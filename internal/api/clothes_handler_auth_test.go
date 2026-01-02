package api

import (
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

type TestServer struct {
	Server             *httptest.Server
	Repo               *DummyClothingRepo
	AuthMw             *AuthMiddleware
	CognitoAppClientID string
	PrivateKey         *rsa.PrivateKey
	Kid                string
}

func setupTestServer(t *testing.T) TestServer {
	t.Helper()

	authMW := &AuthMiddleware{
		CognitoAppClientID: os.Getenv("COGNITO_APP_CLIENT_ID"),
		CognitoUserPoolID:  os.Getenv("COGNITO_USER_POOL_ID"),
		AwsRegion:          os.Getenv("AWS_REGION"),
		JwksUrl:            "http://localhost/mock-jwks.json",
		ValidMethods:       []string{"RS256"},
	}
	authMW.jwksMu.Lock()
	authMW.jwksCache = mockJwks
	authMW.jwksMu.Unlock()

	clothingRepo := &DummyClothingRepo{}

	apiHandler := &API{
		Repo:               clothingRepo,
		CognitoClient:      &DummyCognito{},
		CognitoAppClientID: os.Getenv("COGNITO_APP_CLIENT_ID"),
		CognitoUserPoolID:  os.Getenv("COGNITO_USER_POOL_ID"),
	}

	router := mux.NewRouter()
	router.StrictSlash(true)

	router.HandleFunc("/signup", apiHandler.SignUp).Methods(http.MethodPost)
	router.HandleFunc("/login", apiHandler.Login).Methods(http.MethodPost)

	protectedRouter := router.PathPrefix("/clothes").Subrouter()
	protectedRouter.Use(authMW.Authenticate)

	protectedRouter.HandleFunc("", apiHandler.GetClothing).Methods(http.MethodGet)
	protectedRouter.HandleFunc("", apiHandler.CreateClothing).Methods(http.MethodPost)
	protectedRouter.HandleFunc("/{id}", apiHandler.GetClothingById).Methods(http.MethodGet)
	protectedRouter.HandleFunc("/{id}", apiHandler.UpdateClothing).Methods(http.MethodPost, http.MethodPut, http.MethodPatch)
	protectedRouter.HandleFunc("/{id}", apiHandler.DeleteClothing).Methods(http.MethodDelete)

	testServer := httptest.NewServer(router)

	t.Cleanup(func() {
		testServer.Close()
	})

	ts := TestServer{
		Server:             testServer,
		Repo:               clothingRepo,
		AuthMw:             authMW,
		CognitoAppClientID: apiHandler.CognitoAppClientID,
		PrivateKey:         testPrivateKey,
		Kid:                testKid,
	}

	return ts
}

func TestProtectedClothesEndpoints_Get(t *testing.T) {
	ts := setupTestServer(t)
	baseURL := ts.Server.URL

	t.Run("GET /clothes without Authorization header, should return 401 Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, baseURL+"/clothes", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Authorization header required"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("GET /clothes with bad Authorization header, should return 401 Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, baseURL+"/clothes", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer Bearer Token")
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Authorization header must be in format 'Bearer <token>'"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("GET /clothes with invalid/expired Authorization header, should return 401 Unauthorized", func(t *testing.T) {

		testUserID := "expired-user-123"
		expiredJWT := createTestJWT(t, true, ts.Kid, testIssuer, testUserID, ts.CognitoAppClientID, time.Now().Add(-1*time.Hour), ts.PrivateKey)

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, baseURL+"/clothes", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer "+expiredJWT)
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Invalid token"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("GET /clothes with valid Authorization header, should not return 401", func(t *testing.T) {

		testUserID := "valid-user-123"
		validJWT := createTestJWT(t, true, ts.Kid, testIssuer, testUserID, ts.CognitoAppClientID, time.Now().Add(1*time.Hour), ts.PrivateKey)

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, baseURL+"/clothes", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer "+validJWT)
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code == http.StatusUnauthorized {
			t.Errorf("Expected not to get %d", http.StatusUnauthorized)
		}

	})
}

func TestProtectedClothesEndpoints_Post(t *testing.T) {
	ts := setupTestServer(t)
	baseURL := ts.Server.URL

	t.Run("POST /clothes without Authorization header, should return 401 Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPost, baseURL+"/clothes", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Authorization header required"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("POST /clothes with bad Authorization header, should return 401 Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPost, baseURL+"/clothes", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer Bearer Token")
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Authorization header must be in format 'Bearer <token>'"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("POST /clothes with invalid/expired Authorization header, should return 401 Unauthorized", func(t *testing.T) {

		testUserID := "expired-user-123"
		expiredJWT := createTestJWT(t, true, ts.Kid, testIssuer, testUserID, ts.CognitoAppClientID, time.Now().Add(-1*time.Hour), ts.PrivateKey)

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPost, baseURL+"/clothes", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer "+expiredJWT)
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Invalid token"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("POST /clothes with valid Authorization header, should not return 401", func(t *testing.T) {

		testUserID := "valid-user-123"
		validJWT := createTestJWT(t, true, ts.Kid, testIssuer, testUserID, ts.CognitoAppClientID, time.Now().Add(1*time.Hour), ts.PrivateKey)

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPost, baseURL+"/clothes", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer "+validJWT)
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code == http.StatusUnauthorized {
			t.Errorf("Expected not to get %d", http.StatusUnauthorized)
		}

	})
}

func TestProtectedClothesIdEndpoints_Get(t *testing.T) {
	ts := setupTestServer(t)
	baseURL := ts.Server.URL

	t.Run("GET /clothes/{id} without Authorization header, should return 401 Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Authorization header required"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("GET /clothes/123 with bad Authorization header, should return 401 Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer Bearer Token")
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Authorization header must be in format 'Bearer <token>'"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("GET /clothes/123 with invalid/expired Authorization header, should return 401 Unauthorized", func(t *testing.T) {

		testUserID := "expired-user-123"
		expiredJWT := createTestJWT(t, true, ts.Kid, testIssuer, testUserID, ts.CognitoAppClientID, time.Now().Add(-1*time.Hour), ts.PrivateKey)

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer "+expiredJWT)
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Invalid token"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("GET /clothes/123 with valid Authorization header, should not return 401", func(t *testing.T) {

		testUserID := "valid-user-123"
		validJWT := createTestJWT(t, true, ts.Kid, testIssuer, testUserID, ts.CognitoAppClientID, time.Now().Add(1*time.Hour), ts.PrivateKey)

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer "+validJWT)
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code == http.StatusUnauthorized {
			t.Errorf("Expected not to get %d", http.StatusUnauthorized)
		}

	})
}

func TestProtectedClothesIdEndpoints_Post(t *testing.T) {
	ts := setupTestServer(t)
	baseURL := ts.Server.URL

	t.Run("POST /clothes/{id} without Authorization header, should return 401 Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPost, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Authorization header required"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("POST /clothes/123 with bad Authorization header, should return 401 Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPost, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer Bearer Token")
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Authorization header must be in format 'Bearer <token>'"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("POST /clothes/123 with invalid/expired Authorization header, should return 401 Unauthorized", func(t *testing.T) {

		testUserID := "expired-user-123"
		expiredJWT := createTestJWT(t, true, ts.Kid, testIssuer, testUserID, ts.CognitoAppClientID, time.Now().Add(-1*time.Hour), ts.PrivateKey)

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPost, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer "+expiredJWT)
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Invalid token"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("POST /clothes/123 with valid Authorization header, should not return 401", func(t *testing.T) {

		testUserID := "valid-user-123"
		validJWT := createTestJWT(t, true, ts.Kid, testIssuer, testUserID, ts.CognitoAppClientID, time.Now().Add(1*time.Hour), ts.PrivateKey)

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPost, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer "+validJWT)
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code == http.StatusUnauthorized {
			t.Errorf("Expected not to get %d", http.StatusUnauthorized)
		}

	})
}

func TestProtectedClothesIdEndpoints_Put(t *testing.T) {
	ts := setupTestServer(t)
	baseURL := ts.Server.URL

	t.Run("PUT /clothes/{id} without Authorization header, should return 401 Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPut, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Authorization header required"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("PUT /clothes/123 with bad Authorization header, should return 401 Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPut, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer Bearer Token")
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Authorization header must be in format 'Bearer <token>'"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("PUT /clothes/123 with invalid/expired Authorization header, should return 401 Unauthorized", func(t *testing.T) {

		testUserID := "expired-user-123"
		expiredJWT := createTestJWT(t, true, ts.Kid, testIssuer, testUserID, ts.CognitoAppClientID, time.Now().Add(-1*time.Hour), ts.PrivateKey)

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPut, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer "+expiredJWT)
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Invalid token"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("PUT /clothes/123 with valid Authorization header, should not return 401", func(t *testing.T) {

		testUserID := "valid-user-123"
		validJWT := createTestJWT(t, true, ts.Kid, testIssuer, testUserID, ts.CognitoAppClientID, time.Now().Add(1*time.Hour), ts.PrivateKey)

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPut, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer "+validJWT)
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code == http.StatusUnauthorized {
			t.Errorf("Expected not to get %d", http.StatusUnauthorized)
		}

	})
}

func TestProtectedClothesIdEndpoints_Patch(t *testing.T) {
	ts := setupTestServer(t)
	baseURL := ts.Server.URL

	t.Run("PATCH /clothes/{id} without Authorization header, should return 401 Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPatch, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Authorization header required"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("PATCH /clothes/123 with bad Authorization header, should return 401 Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPatch, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer Bearer Token")
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Authorization header must be in format 'Bearer <token>'"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("PATCH /clothes/123 with invalid/expired Authorization header, should return 401 Unauthorized", func(t *testing.T) {

		testUserID := "expired-user-123"
		expiredJWT := createTestJWT(t, true, ts.Kid, testIssuer, testUserID, ts.CognitoAppClientID, time.Now().Add(-1*time.Hour), ts.PrivateKey)

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPatch, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer "+expiredJWT)
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Invalid token"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("PATCH /clothes/123 with valid Authorization header, should not return 401", func(t *testing.T) {

		testUserID := "valid-user-123"
		validJWT := createTestJWT(t, true, ts.Kid, testIssuer, testUserID, ts.CognitoAppClientID, time.Now().Add(1*time.Hour), ts.PrivateKey)

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPatch, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer "+validJWT)
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code == http.StatusUnauthorized {
			t.Errorf("Expected not to get %d", http.StatusUnauthorized)
		}

	})
}

func TestProtectedClothesIdEndpoints_Delete(t *testing.T) {
	ts := setupTestServer(t)
	baseURL := ts.Server.URL

	t.Run("DELETE /clothes/{id} without Authorization header, should return 401 Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodDelete, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Authorization header required"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("DELETE /clothes/123 with bad Authorization header, should return 401 Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodDelete, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer Bearer Token")
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Authorization header must be in format 'Bearer <token>'"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("DELETE /clothes/123 with invalid/expired Authorization header, should return 401 Unauthorized", func(t *testing.T) {

		testUserID := "expired-user-123"
		expiredJWT := createTestJWT(t, true, ts.Kid, testIssuer, testUserID, ts.CognitoAppClientID, time.Now().Add(-1*time.Hour), ts.PrivateKey)

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodDelete, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer "+expiredJWT)
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d; got %d", http.StatusUnauthorized, w.Code)
		}
		expectedBody := "Invalid token"
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected response body to contain %q; got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("DELETE /clothes/123 with valid Authorization header, should not return 401", func(t *testing.T) {

		testUserID := "valid-user-123"
		validJWT := createTestJWT(t, true, ts.Kid, testIssuer, testUserID, ts.CognitoAppClientID, time.Now().Add(1*time.Hour), ts.PrivateKey)

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodDelete, baseURL+"/clothes/123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		r.Header.Set("Authorization", "Bearer "+validJWT)
		r.Header.Set("Content-Type", "application/json")

		ts.Server.Config.Handler.ServeHTTP(w, r)

		if w.Code == http.StatusUnauthorized {
			t.Errorf("Expected not to get %d", http.StatusUnauthorized)
		}

	})
}
