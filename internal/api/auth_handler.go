package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

func (a *API) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("Unsupported method %s.", r.Method), http.StatusMethodNotAllowed)
		return
	}

	if r.Body == nil || r.Body == http.NoBody {
		http.Error(w, "Request body must not be empty or missing", http.StatusBadRequest)
		return
	}

	var req SignupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	isValidEmail := validateEmail(req.Email)

	if !isValidEmail {
		http.Error(w, "Email is not valid", http.StatusBadRequest)
		return
	}

	isValidPassword := validatePassword(req.Password)

	if !isValidPassword {
		http.Error(w, "Password is not valid", http.StatusBadRequest)
		return
	}

	signUpInput := &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(a.CognitoAppClientID), // Use the injected App Client ID
		Username: aws.String(req.Email),            // Email as the username for Cognito
		Password: aws.String(req.Password),
	}

	result, err := a.CognitoClient.SignUp(context.TODO(), signUpInput)

	if err != nil {
		log.Printf("ERROR: Cognito SignUp failed for user %s: %v", req.Email, err)
		var usernameExistsError *types.UsernameExistsException
		if errors.As(err, &usernameExistsError) {
			http.Error(w, "User with this email already exists", http.StatusConflict)
		} else {
			http.Error(w, "Failed to register user", http.StatusInternalServerError)
		}
		return
	}

	resp := SignupResponse{
		Message: "User registered successfully. Awaiting administrator confirmation.",
		UserID:  *result.UserSub, // UserSub is the immutable ID assigned by Cognito
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)

}

func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("Unsupported method %s.", r.Method), http.StatusMethodNotAllowed)
		return
	}

	if r.Body == nil || r.Body == http.NoBody {
		http.Error(w, "Request body must not be empty or missing", http.StatusBadRequest)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	isValidEmail := validateEmail(req.Email)

	if !isValidEmail {
		http.Error(w, "Email is not valid", http.StatusBadRequest)
		return
	}

	isValidPassword := validatePassword(req.Password)

	if !isValidPassword {
		http.Error(w, "Password is not valid", http.StatusBadRequest)
		return
	}

	initiateAuthInput := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String(a.CognitoAppClientID),
		AuthParameters: map[string]string{
			"USERNAME": req.Email,
			"PASSWORD": req.Password,
		},
	}

	result, err := a.CognitoClient.InitiateAuth(context.TODO(), initiateAuthInput)
	if err != nil {
		log.Printf("ERROR: Cognito Login failed for user %s: %v", req.Email, err)
		var notAuthErr *types.NotAuthorizedException
		var userNotFoundErr *types.UserNotFoundException
		var userNotConfirmedErr *types.UserNotConfirmedException

		if errors.As(err, &notAuthErr) {
			http.Error(w, "Incorrect email or password", http.StatusUnauthorized)
		} else if errors.As(err, &userNotFoundErr) {
			log.Printf("User %s not found", req.Email)
			http.Error(w, "Failed to login user", http.StatusInternalServerError)
		} else if errors.As(err, &userNotConfirmedErr) {
			http.Error(w, "User account is not confirmed by an administrator", http.StatusForbidden)
		} else {
			http.Error(w, "Failed to login user", http.StatusInternalServerError)
		}
		return
	}

	if result.AuthenticationResult == nil {
		log.Printf("WARN: Cognito InitiateAuth did not return AuthenticationResult for user %s. Possible MFA challenge or other flow.", req.Email)
		http.Error(w, "Authentication flow requires additional steps (e.g., MFA) which are not yet implemented.", http.StatusForbidden)
		return
	}

	resp := LoginResponse{
		AccessToken:  *result.AuthenticationResult.AccessToken,
		IdToken:      *result.AuthenticationResult.IdToken,
		RefreshToken: *result.AuthenticationResult.RefreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func validateEmail(email string) bool {

	trimmedEmail := strings.TrimSpace(email)

	if trimmedEmail == "" {
		return false
	}

	if !strings.Contains(trimmedEmail, "@") {
		return false
	}

	emailParts := strings.Split(trimmedEmail, "@")

	if len(emailParts) != 2 {
		return false
	}

	domain := emailParts[1]

	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}

func validatePassword(password string) bool {
	trimmedPassword := strings.TrimSpace(password)

	if trimmedPassword == "" {
		return false
	}

	if utf8.RuneCountInString(trimmedPassword) < 8 {
		return false
	}

	specialChars := `^$*.[\]{}()?-"!@#%&/\,><':;|_~` + "`+="

	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, char := range trimmedPassword {
		if 'a' <= char && char <= 'z' {
			hasLower = true
		} else if 'A' <= char && char <= 'Z' {
			hasUpper = true
		} else if '0' <= char && char <= '9' {
			hasDigit = true
		} else if strings.ContainsRune(specialChars, char) {
			hasSpecial = true
		}

		if hasLower && hasUpper && hasDigit && hasSpecial {
			return true
		}
	}

	return false
}
