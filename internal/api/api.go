package api

import (
	"clothes_management/internal/repository"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

type API struct {
	Repo               repository.ClothingRepository
	CognitoClient      *cognitoidentityprovider.Client
	CognitoAppClientID string
	CognitoUserPoolID  string
}
