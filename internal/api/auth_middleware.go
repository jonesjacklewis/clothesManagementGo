package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/jwk"
)

type AuthMiddleware struct {
	CognitoAppClientID string
	CognitoUserPoolID  string
	JwksUrl            string
	Region             string

	jwksCache jwk.Set
	jwksMu    sync.RWMutex
}

type contextKey string

const UserIDContextKey contextKey = "userID"

func NewAuthMiddleware(appClientID, userPoolID, jwksURL, region string) (*AuthMiddleware, error) {
	if strings.TrimSpace(appClientID) == "" ||
		strings.TrimSpace(userPoolID) == "" ||
		strings.TrimSpace(jwksURL) == "" ||
		strings.TrimSpace(region) == "" {
		return nil, fmt.Errorf("appClientID, userPoolID, jwksURL, and region cannot be empty")
	}

	mw := &AuthMiddleware{
		CognitoAppClientID: appClientID,
		CognitoUserPoolID:  userPoolID,
		JwksUrl:            jwksURL,
		Region:             region,
	}

	// Initial fetch of JWKS
	if err := mw.refreshJwksCache(); err != nil {
		return nil, fmt.Errorf("failed to fetch initial JWKS from %s: %w", jwksURL, err)
	}
	log.Printf("INFO: Initial JWKS fetched successfully from %s", jwksURL)

	// Start a background goroutine to periodically refresh the JWKS cache
	go mw.startJwksRefresher()

	return mw, nil
}

func (a *AuthMiddleware) refreshJwksCache() error {
	set, err := jwk.Fetch(context.Background(), a.JwksUrl)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	a.jwksMu.Lock() // Write lock
	a.jwksCache = set
	a.jwksMu.Unlock()

	return nil
}

func (a *AuthMiddleware) startJwksRefresher() {
	ticker := time.NewTicker(5 * time.Minute) // Refresh every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		log.Println("INFO: Refreshing JWKS cache...")
		if err := a.refreshJwksCache(); err != nil {
			log.Printf("ERROR: Failed to refresh JWKS cache: %v", err)
		} else {
			log.Println("INFO: JWKS cache refreshed successfully.")
		}
	}
}

func (a *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Authorization header must be in format 'Bearer <token>'", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]
		expectedIss := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", a.Region, a.CognitoUserPoolID)

		// --- JWT Parsing and Validation ---
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			// Ensure token is signed with an expected algorithm
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				log.Printf("WARN: Unexpected signing method: %v", token.Header["alg"])
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Look up the key in the JWKS cache
			kid, ok := token.Header["kid"].(string)
			if !ok {
				log.Printf("WARN: Token does not have a 'kid' header")
				return nil, fmt.Errorf("token does not have a 'kid' header")
			}

			a.jwksMu.RLock() // Read lock
			defer a.jwksMu.RUnlock()

			key, found := a.jwksCache.LookupKeyID(kid)
			if !found {
				log.Printf("WARN: Public key with KID '%s' not found in JWKS cache", kid)
				return nil, fmt.Errorf("public key with KID '%s' not found", kid)
			}

			// Extract the RSA public key from the JWK
			var publicKey any
			if err := key.Raw(&publicKey); err != nil {
				log.Printf("ERROR: Failed to get raw public key from JWK: %v", err)
				return nil, fmt.Errorf("failed to get public key for validation: %w", err)
			}
			return publicKey, nil
		},
			jwt.WithValidMethods([]string{"RS256"}), // Cognito ID Tokens typically use RS256
			jwt.WithIssuer(expectedIss),             // Validate the 'iss' claim
			jwt.WithExpirationRequired(),            // Validate the 'exp' claim
			jwt.WithTimeFunc(time.Now),              // Use current time for expiry checks
		)

		if err != nil {
			log.Printf("ERROR: JWT validation failed: %v", err)
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized) // Return specific error to client for debugging
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid token signature or claims", http.StatusUnauthorized)
			return
		}

		// --- Extract UserID from Claims ---
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Printf("ERROR: Could not get claims from token as MapClaims")
			http.Error(w, "Invalid token claims format", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["sub"].(string) // 'sub' claim is the user's immutable ID in Cognito
		if !ok || strings.TrimSpace(userID) == "" {
			log.Printf("ERROR: 'sub' claim not found or empty in token for user %v", claims["username"])
			http.Error(w, "User ID not found in token", http.StatusUnauthorized)
			return
		}

		// --- Inject UserID into Context ---
		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)

		// Call the next handler in the chain with the new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
