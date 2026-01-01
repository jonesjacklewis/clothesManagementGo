package api

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/lestrrat-go/jwx/jwk"
)

var (
	testPrivateKey *rsa.PrivateKey
	testKid        string = "test-key-id-123"
	testIssuer     string
	mockJwks       jwk.Set
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../.env_test"); err != nil {
		fmt.Printf("INFO (test): No .env_test file found or error loading: %v. Relying on system environment variables.\n", err)
	} else {
		fmt.Println("INFO (test): .env_test file loaded for tests.")
	}

	var err error
	testPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("ERROR: Failed to generate RSA private key for tests: %v\n", err)
		os.Exit(1)
	}

	publicKey := &testPrivateKey.PublicKey

	testJwkKey, err := jwk.New(publicKey)
	if err != nil {
		fmt.Printf("ERROR: Failed to create JWK from public key: %v\n", err)
		os.Exit(1)
	}
	testJwkKey.Set(jwk.KeyIDKey, testKid)

	mockJwks = jwk.NewSet()
	mockJwks.Add(testJwkKey)

	awsRegion := os.Getenv("AWS_REGION")
	cognitoUserPoolId := os.Getenv("COGNITO_USER_POOL_ID")
	if awsRegion == "" || cognitoUserPoolId == "" {
		fmt.Println("WARNING: AWS_REGION or COGNITO_USER_POOL_ID not set, test issuer will be incomplete.")
	}
	testIssuer = fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", awsRegion, cognitoUserPoolId)

	// Run all tests in the package
	os.Exit(m.Run())
}

func setupMockedAuthMiddleware(t *testing.T) *AuthMiddleware {
	cognitoAppClientId := os.Getenv("COGNITO_APP_CLIENT_ID")
	if cognitoAppClientId == "" {
		t.Fatal("ERROR: COGNITO_APP_CLIENT_ID environment variable not set. Please set it in .env or your shell.")
	}

	cognitoUserPoolId := os.Getenv("COGNITO_USER_POOL_ID")
	if cognitoUserPoolId == "" {
		t.Fatal("ERROR: COGNITO_USER_POOL_ID environment variable not set. Please set it in .env or your shell.")
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		t.Fatal("ERROR: AWS_REGION environment variable not set. Please set it in .env or your shell.")
	}

	jwksUrl := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", awsRegion, cognitoUserPoolId)

	mw := &AuthMiddleware{
		CognitoAppClientID: cognitoAppClientId,
		CognitoUserPoolID:  cognitoUserPoolId,
		AwsRegion:          awsRegion,
		JwksUrl:            jwksUrl,
	}

	mw.jwksMu.Lock()
	mw.jwksCache = mockJwks
	mw.jwksMu.Unlock()

	return mw
}

// createTestJWT is a helper function to generate a signed JWT for testing.
func createTestJWT(t *testing.T, addKid bool, kid, issuer, subject, audience string, expiration time.Time, signingKey *rsa.PrivateKey) string {
	t.Helper() // Mark as a test helper

	// 1. Define the claims for the JWT
	claims := jwt.MapClaims{
		"iss": issuer,            // The issuer (e.g., Cognito User Pool URL)
		"aud": audience,          // The audience (e.g., Cognito App Client ID)
		"sub": subject,           // The subject (the UserID)
		"exp": expiration.Unix(), // Expiration time as Unix timestamp
		"iat": time.Now().Unix(), // Issued at time
		// Add any other claims your middleware might check
	}

	// 2. Create the JWT token object
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// 3. Set the Key ID (kid) in the token header
	if addKid {
		token.Header["kid"] = kid
	}

	// 4. Sign the token with the private key
	signedToken, err := token.SignedString(signingKey)
	if err != nil {
		t.Fatalf("Failed to sign test JWT: %v", err)
	}

	return signedToken
}

func createHS256JWT(t *testing.T, issuer, audience, subject string) string {
	t.Helper()

	claims := jwt.MapClaims{
		"iss": issuer,
		"aud": audience,
		"sub": subject,
		"exp": time.Now().Add(5 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok.Header["kid"] = "anything" // doesn’t matter; your alg check happens first

	s, err := tok.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign HS256 token: %v", err)
	}
	return s
}

func TestAuthenticate(t *testing.T) {
	t.Run("Given missing Authorization header, should return 401 Unauthorized", func(t *testing.T) {
		mw := setupMockedAuthMiddleware(t)

		dummyNextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDContextKey).(string)
			if !ok || userID == "" {
				http.Error(w, "UserID not found in context (next handler error)", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Next handler reached, UserID: %s", userID)))
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/clothes", nil)

		mw.Authenticate(dummyNextHandler).ServeHTTP(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, resp.StatusCode)
		}

		expected := "Authorization header required"

		if !strings.Contains(w.Body.String(), expected) {
			t.Errorf("Expected %s got %s", expected, w.Body.String())
		}
	})

	t.Run("Given Authorization header contains more than 1 space, should return 401 Unauthorized", func(t *testing.T) {
		mw := setupMockedAuthMiddleware(t)

		dummyNextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDContextKey).(string)
			if !ok || userID == "" {
				http.Error(w, "UserID not found in context (next handler error)", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Next handler reached, UserID: %s", userID)))
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/clothes", nil)
		r.Header.Set("Authorization", "Bearer TokenValue Extra")

		mw.Authenticate(dummyNextHandler).ServeHTTP(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, resp.StatusCode)
		}

		expected := "Authorization header must be in format 'Bearer <token>'"

		if !strings.Contains(w.Body.String(), expected) {
			t.Errorf("Expected %s got %s", expected, w.Body.String())
		}
	})

	t.Run("Given Authorization header doesn't contain bearer, should return 401 Unauthorized", func(t *testing.T) {
		mw := setupMockedAuthMiddleware(t)

		dummyNextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDContextKey).(string)
			if !ok || userID == "" {
				http.Error(w, "UserID not found in context (next handler error)", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Next handler reached, UserID: %s", userID)))
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/clothes", nil)
		r.Header.Set("Authorization", "Not TokenValue")

		mw.Authenticate(dummyNextHandler).ServeHTTP(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, resp.StatusCode)
		}

		expected := "Authorization header must be in format 'Bearer <token>'"

		if !strings.Contains(w.Body.String(), expected) {
			t.Errorf("Expected %s got %s", expected, w.Body.String())
		}
	})

	t.Run("Given invalid JWT method, should return 401 Unauthorized", func(t *testing.T) {
		mw := setupMockedAuthMiddleware(t)
		mw.ValidMethods = []string{"RS256", "HS256"}

		expectedIss := testIssuer
		aud := mw.CognitoAppClientID

		badToken := createHS256JWT(t, expectedIss, aud, "user-123")

		dummyNextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDContextKey).(string)
			if !ok || userID == "" {
				http.Error(w, "UserID not found in context (next handler error)", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Next handler reached, UserID: %s", userID)))
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/clothes", nil)
		r.Header.Set("Authorization", "Bearer "+badToken)

		mw.Authenticate(dummyNextHandler).ServeHTTP(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, resp.StatusCode)
		}

		expected := "unexpected signing method"

		if !strings.Contains(w.Body.String(), expected) {
			t.Errorf("Expected %s got %s", expected, w.Body.String())
		}
	})

	t.Run("Given missing kid header, should return 401 Unauthorized", func(t *testing.T) {
		mw := setupMockedAuthMiddleware(t)

		testUserID := "valid-user-id-abc"
		token := createTestJWT(t, false, "", testIssuer, testUserID, mw.CognitoAppClientID, time.Now().Add(1*time.Hour), testPrivateKey)

		dummyNextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDContextKey).(string)
			if !ok || userID == "" {
				http.Error(w, "UserID not found in context (next handler error)", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Next handler reached, UserID: %s", userID)))
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/clothes", nil)
		r.Header.Set("Authorization", "Bearer "+token)

		mw.Authenticate(dummyNextHandler).ServeHTTP(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, resp.StatusCode)
		}

		expected := "token does not have a 'kid' header"

		if !strings.Contains(w.Body.String(), expected) {
			t.Errorf("Expected %s got %s", expected, w.Body.String())
		}
	})

	t.Run("Given unknown kid, should return 401 Unauthorized", func(t *testing.T) {
		mw := setupMockedAuthMiddleware(t)

		testUserID := "valid-user-id-abc"
		badKid := testKid + uuid.New().String()
		token := createTestJWT(t, true, badKid, testIssuer, testUserID, mw.CognitoAppClientID, time.Now().Add(1*time.Hour), testPrivateKey)

		dummyNextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDContextKey).(string)
			if !ok || userID == "" {
				http.Error(w, "UserID not found in context (next handler error)", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Next handler reached, UserID: %s", userID)))
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/clothes", nil)
		r.Header.Set("Authorization", "Bearer "+token)

		mw.Authenticate(dummyNextHandler).ServeHTTP(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, resp.StatusCode)
		}

		expected := fmt.Sprintf("public key with KID '%s' not found", badKid)

		if !strings.Contains(w.Body.String(), expected) {
			t.Errorf("Expected %s got %s", expected, w.Body.String())
		}
	})

	t.Run("Given empty user ID, should return 401 Unauthorized", func(t *testing.T) {
		mw := setupMockedAuthMiddleware(t)

		testUserID := ""
		token := createTestJWT(t, true, testKid, testIssuer, testUserID, mw.CognitoAppClientID, time.Now().Add(1*time.Hour), testPrivateKey)

		dummyNextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDContextKey).(string)
			if !ok || userID == "" {
				http.Error(w, "UserID not found in context (next handler error)", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Next handler reached, UserID: %s", userID)))
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/clothes", nil)
		r.Header.Set("Authorization", "Bearer "+token)

		mw.Authenticate(dummyNextHandler).ServeHTTP(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, resp.StatusCode)
		}

		expected := "User ID not found in token"

		if !strings.Contains(w.Body.String(), expected) {
			t.Errorf("Expected %s got %s", expected, w.Body.String())
		}
	})

	t.Run("Given valid, should return 200 OK", func(t *testing.T) {
		mw := setupMockedAuthMiddleware(t)

		testUserID := "valid-user-id-abc"
		token := createTestJWT(t, true, testKid, testIssuer, testUserID, mw.CognitoAppClientID, time.Now().Add(1*time.Hour), testPrivateKey)

		dummyNextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDContextKey).(string)
			if !ok || userID == "" {
				http.Error(w, "UserID not found in context (next handler error)", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Next handler reached, UserID: %s", userID)))
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/clothes", nil)
		r.Header.Set("Authorization", "Bearer "+token)

		mw.Authenticate(dummyNextHandler).ServeHTTP(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d; got %d", http.StatusOK, resp.StatusCode)
		}

		expected := "Next handler reached"

		if !strings.Contains(w.Body.String(), expected) {
			t.Errorf("Expected %s got %s", expected, w.Body.String())
		}
	})

}
