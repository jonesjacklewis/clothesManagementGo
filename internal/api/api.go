package api

import (
	"clothes_management/internal/repository"
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

type CognitoAPI interface {
	SignUp(ctx context.Context, params *cognitoidentityprovider.SignUpInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.SignUpOutput, error)
	InitiateAuth(ctx context.Context, params *cognitoidentityprovider.InitiateAuthInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.InitiateAuthOutput, error)
}

type API struct {
	Repo               repository.ClothingRepository
	CognitoClient      CognitoAPI
	CognitoAppClientID string
	CognitoUserPoolID  string
}
