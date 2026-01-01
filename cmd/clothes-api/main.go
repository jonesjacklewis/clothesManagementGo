package main

import (
	"clothes_management/internal/api"
	"clothes_management/internal/repository"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Printf("INFO: No .env file found, or error loading .env: %v. Relying on system environment variables.", err)
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		log.Fatal("ERROR: AWS_REGION environment variable not set. Please set it in .env or your shell.")
	}
	dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if dynamoTableName == "" {
		log.Fatal("ERROR: DYNAMODB_TABLE_NAME environment variable not set. Please set it in .env or your shell.")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		log.Fatalf("ERROR: Failed to load AWS SDK config: %v", err)
	}
	dynamoClient := dynamodb.NewFromConfig(cfg)

	_, err = dynamoClient.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String(dynamoTableName),
	})
	if err != nil {
		var notFoundEx *types.ResourceNotFoundException
		if errors.As(err, &notFoundEx) {
			log.Fatalf("ERROR: DynamoDB table '%s' not found in region '%s'. Please create it. Error: %v", dynamoTableName, awsRegion, err)
		} else {
			log.Fatalf("ERROR: Failed to describe DynamoDB table '%s'. Check region, credentials, and permissions: %v", dynamoTableName, err)
		}
	}
	log.Printf("SUCCESS: Successfully connected to DynamoDB table '%s' in region '%s'.", dynamoTableName, awsRegion)

	cognitoUserPoolId := os.Getenv("COGNITO_USER_POOL_ID")
	if cognitoUserPoolId == "" {
		log.Fatal("ERROR: COGNITO_USER_POOL_ID environment variable not set. Please set it in .env or your shell.")
	}

	cognitoAppClientId := os.Getenv("COGNITO_APP_CLIENT_ID")
	if cognitoAppClientId == "" {
		log.Fatal("ERROR: COGNITO_APP_CLIENT_ID environment variable not set. Please set it in .env or your shell.")
	}

	jwksUrl := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", awsRegion, cognitoUserPoolId)

	cognitoClient := cognitoidentityprovider.NewFromConfig(cfg)

	port := 8080
	portStr := fmt.Sprintf(":%d", port)

	repo, err := repository.NewDynamoDBClothingRepository(dynamoClient, dynamoTableName)

	if err != nil {
		log.Fatalf("ERROR: Failed to create instance of DynamoDBClothingRepository %v", err)
	}

	apiHandler := &api.API{
		Repo:               repo,
		CognitoClient:      cognitoClient,
		CognitoAppClientID: cognitoAppClientId,
		CognitoUserPoolID:  cognitoUserPoolId,
	}

	authMiddleware, err := api.NewAuthMiddleware(cognitoAppClientId, cognitoUserPoolId, jwksUrl, awsRegion)

	if err != nil {
		log.Fatalf("ERROR: Failed to create AuthMiddleware: %v", err)
	}

	router := mux.NewRouter()

	router.StrictSlash(true)

	router.HandleFunc("/signup", apiHandler.SignUp).Methods(http.MethodPost)
	router.HandleFunc("/login", apiHandler.Login).Methods(http.MethodPost)

	protectedRouter := router.PathPrefix("/clothes").Subrouter()
	protectedRouter.Use(authMiddleware.Authenticate)

	protectedRouter.HandleFunc("", apiHandler.GetClothing).Methods(http.MethodGet)
	protectedRouter.HandleFunc("", apiHandler.CreateClothing).Methods(http.MethodPost)
	protectedRouter.HandleFunc("/{id}", apiHandler.GetClothingById).Methods(http.MethodGet)
	protectedRouter.HandleFunc("/{id}", apiHandler.UpdateClothing).Methods(http.MethodPost, http.MethodPut, http.MethodPatch)
	protectedRouter.HandleFunc("/{id}", apiHandler.DeleteClothing).Methods(http.MethodDelete)

	srv := &http.Server{
		Addr:         portStr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	log.Printf("INFO: Server starting on port %d...", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ERROR: HTTP server failed: %v", err)
	}
	log.Println("INFO: Server gracefully stopped.")
}
