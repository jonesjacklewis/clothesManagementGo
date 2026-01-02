package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type DummyCognito struct {
	ShouldErrOnSignUp bool
	SignUpErr         error
	SignUpUserSub     string
}

func (d *DummyCognito) SignUp(ctx context.Context, params *cognitoidentityprovider.SignUpInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.SignUpOutput, error) {

	if d.ShouldErrOnSignUp {
		return nil, d.SignUpErr
	}

	var result cognitoidentityprovider.SignUpOutput = cognitoidentityprovider.SignUpOutput{
		UserSub: aws.String(d.SignUpUserSub),
	}

	return &result, nil
}

func (d *DummyCognito) InitiateAuth(ctx context.Context, params *cognitoidentityprovider.InitiateAuthInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.InitiateAuthOutput, error) {
	return nil, nil
}

func TestValidateEmail(t *testing.T) {
	t.Run("Given email is empty/whitespace, should return false", func(t *testing.T) {
		email := "   "
		isValid := validateEmail(email)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given email doesn't contain @, should return false", func(t *testing.T) {
		email := "usernameATdomain.com"
		isValid := validateEmail(email)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given email contains more than one @, should return false", func(t *testing.T) {
		email := "user@name@domain.com"
		isValid := validateEmail(email)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given email doesn't contain ., should return false", func(t *testing.T) {
		email := "user@domainDOTcom"
		isValid := validateEmail(email)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given valid email, should return true", func(t *testing.T) {
		email := "user@domain.com"
		isValid := validateEmail(email)

		if !isValid {
			t.Error("Expected isValid = true")
		}
	})
}

func TestValidatePassword(t *testing.T) {
	t.Run("Given password is empty/whitespace, should return false", func(t *testing.T) {
		password := "   "
		isValid := validatePassword(password)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given password is less than 8 characters, should return false", func(t *testing.T) {
		password := "A1!defg"
		isValid := validatePassword(password)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given password is less than 8 characters (where e.g. emoji 😀 = 1), should return false", func(t *testing.T) {
		password := "A1!def😀"
		isValid := validatePassword(password)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given password doesn't contain a lowercase letter, should return false", func(t *testing.T) {
		password := "A1!BCDEFG"
		isValid := validatePassword(password)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given password doesn't contain an upper letter, should return false", func(t *testing.T) {
		password := "a1!bcdefg"
		isValid := validatePassword(password)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given password doesn't contain a number, should return false", func(t *testing.T) {
		password := "a!!bcdefG"
		isValid := validatePassword(password)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given password doesn't contain a special character, should return false", func(t *testing.T) {
		password := "a12bcdefG"
		isValid := validatePassword(password)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given password meets criteria, should return true", func(t *testing.T) {
		password := "a1!bcdefG"
		isValid := validatePassword(password)

		if !isValid {
			t.Error("Expected isValid = true")
		}
	})

}

func TestSignUp(t *testing.T) {
	t.Run("Given method is not POST, should return StatusMethodNotAllowed", func(t *testing.T) {
		dummyCognito := &DummyCognito{}

		apiHandler := &API{
			CognitoClient: dummyCognito,
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/signup", nil)

		apiHandler.SignUp(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected %d got %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}

		expectedMessage := fmt.Sprintf("Unsupported method %s.", http.MethodGet)

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given method is POST, but body is nil, should return StatusBadRequest", func(t *testing.T) {
		dummyCognito := &DummyCognito{}

		apiHandler := &API{
			CognitoClient: dummyCognito,
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/signup", nil)

		apiHandler.SignUp(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Request body must not be empty or missing"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given method is POST, but body is invalid JSON, should return StatusBadRequest", func(t *testing.T) {
		dummyCognito := &DummyCognito{}

		apiHandler := &API{
			CognitoClient: dummyCognito,
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader("data"))

		apiHandler.SignUp(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Request body must be JSON"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given method is POST, but body does not contain email, should return StatusBadRequest", func(t *testing.T) {
		dummyCognito := &DummyCognito{}

		apiHandler := &API{
			CognitoClient: dummyCognito,
		}

		w := httptest.NewRecorder()
		var jsonMap map[string]any = map[string]any{
			"notEmail": "valid@domain.com",
			"password": "Password123!",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))

		apiHandler.SignUp(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Request body missing email field"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given method is POST, but body does not contain password, should return StatusBadRequest", func(t *testing.T) {
		dummyCognito := &DummyCognito{}

		apiHandler := &API{
			CognitoClient: dummyCognito,
		}

		w := httptest.NewRecorder()
		var jsonMap map[string]any = map[string]any{
			"email":       "valid@domain.com",
			"notPassword": "Password123!",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))

		apiHandler.SignUp(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Request body missing password field"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given method is POST, but email is invalid, should return StatusBadRequest", func(t *testing.T) {
		dummyCognito := &DummyCognito{}

		apiHandler := &API{
			CognitoClient: dummyCognito,
		}

		w := httptest.NewRecorder()
		var jsonMap map[string]any = map[string]any{
			"email":    "v@lid@domain.com",
			"password": "Password123!",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))

		apiHandler.SignUp(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Email is not valid"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given method is POST, but password is invalid, should return StatusBadRequest", func(t *testing.T) {
		dummyCognito := &DummyCognito{}

		apiHandler := &API{
			CognitoClient: dummyCognito,
		}

		w := httptest.NewRecorder()
		var jsonMap map[string]any = map[string]any{
			"email":    "valid@domain.com",
			"password": "Password123",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))

		apiHandler.SignUp(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected %d got %d", http.StatusBadRequest, resp.StatusCode)
		}

		expectedMessage := "Password is not valid"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given method is POST, but email already exists, should return StatusConflict", func(t *testing.T) {
		dummyCognito := &DummyCognito{
			ShouldErrOnSignUp: true,
			SignUpErr: &types.UsernameExistsException{
				Message: aws.String("User with this email already exists"),
			},
		}

		apiHandler := &API{
			CognitoClient: dummyCognito,
		}

		w := httptest.NewRecorder()
		var jsonMap map[string]any = map[string]any{
			"email":    "valid@domain.com",
			"password": "Password123!",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))

		apiHandler.SignUp(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected %d got %d", http.StatusConflict, resp.StatusCode)
		}

		expectedMessage := "User with this email already exists"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given method is POST, but issue with signup, should return StatusInternalServerError", func(t *testing.T) {
		dummyCognito := &DummyCognito{
			ShouldErrOnSignUp: true,
			SignUpErr:         errors.New("Something went wrong"),
		}

		apiHandler := &API{
			CognitoClient: dummyCognito,
		}

		w := httptest.NewRecorder()
		var jsonMap map[string]any = map[string]any{
			"email":    "valid@domain.com",
			"password": "Password123!",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))

		apiHandler.SignUp(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected %d got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		expectedMessage := "Failed to register user"

		if !strings.Contains(w.Body.String(), expectedMessage) {
			t.Errorf("Expected %s got %s", expectedMessage, w.Body.String())
		}
	})

	t.Run("Given method is POST, with valid data, should return StatusCreated", func(t *testing.T) {

		expectedUserSub := "a1b2c3d4-5678-90ab-cdef-EXAMPLE11111"

		dummyCognito := &DummyCognito{
			ShouldErrOnSignUp: false,
			SignUpErr:         nil,
			SignUpUserSub:     expectedUserSub,
		}

		apiHandler := &API{
			CognitoClient: dummyCognito,
		}

		w := httptest.NewRecorder()
		var jsonMap map[string]any = map[string]any{
			"email":    "valid@domain.com",
			"password": "Password123!",
		}

		body, err := json.Marshal(jsonMap)
		if err != nil {
			t.Fatalf("failed to marshal json: %v", err)
		}

		r := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))

		apiHandler.SignUp(w, r)

		resp := w.Result()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected %d got %d", http.StatusCreated, resp.StatusCode)
		}

		var responseBody map[string]any
		err = json.Unmarshal(w.Body.Bytes(), &responseBody)

		if err != nil {
			t.Fatalf("Response body is not valid JSON: %v", err)
		}

		message, exists := responseBody["message"]

		if !exists {
			t.Error("Missing message property")
		}

		expectedMessage := "User registered successfully. Awaiting administrator confirmation."

		if message.(string) != expectedMessage {
			t.Errorf("Expected %s got %s", expectedMessage, message.(string))
		}

		userId, exists := responseBody["userId"]

		if !exists {
			t.Error("Missing userId property")
		}

		if userId.(string) != expectedUserSub {
			t.Errorf("Expected %s got %s", expectedUserSub, userId.(string))
		}

	})
}
