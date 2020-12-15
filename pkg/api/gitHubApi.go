package api

import (
	"Orca/pkg/crypto"
	"context"
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/go-github/v33/github"
	"net/http"
	"time"
)

func GetGitHubApiClient(installationId int64, appId int, privateKey rsa.PrivateKey) (*github.Client, error) {

	// Get the GitHub App Installation access token
	accessToken, err := getInstallationAccessToken(installationId, appId, privateKey)
	if err != nil {
		return nil, err
	}

	// Create a new HTTP client with the access token being injected
	httpClient := getHttpClientWithInjectedToken(*accessToken)

	client := github.NewClient(&httpClient)

	return client, nil
}

func getInstallationAccessToken(installationId int64, appId int, privateKey rsa.PrivateKey) (*string, error) {

	// To get the Installation access token, we first need the Apps JWT
	appToken, err := getAppJsonWebToken(appId, privateKey)
	if err != nil {
		return nil, err
	}

	// Create a new HTTP client with the JWT so we can request an Installation access token
	httpClient := getHttpClientWithInjectedToken(*appToken)
	gitHubClient := github.NewClient(&httpClient)

	// Get the Installation access token
	tokenResponse, _, err := gitHubClient.Apps.CreateInstallationToken(context.Background(), installationId, nil)
	if err != nil {
		return nil, err
	}

	return tokenResponse.Token, nil
}

func getAppJsonWebToken(appId int, privatKey rsa.PrivateKey) (*string, error){

	// Build the JWT
	token := jwt.New(jwt.GetSigningMethod("RS256"))
	token.Claims["iat"] = time.Now().Unix()
	token.Claims["exp"] = time.Now().Add(time.Minute * 5).Unix()
	token.Claims["iss"] = appId

	// Sign and get the complete encoded token as a string
	privateKeyBytes := crypto.EncodePrivateKey(&privatKey)
	tokenString, err := token.SignedString(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return &tokenString, nil
}

func getHttpClientWithInjectedToken(token string) http.Client {
	httpClient := http.Client {
		Transport: &authorizedTransport {
			underlyingTransport: http.DefaultTransport,
			bearerToken: token,
		},
	}

	return httpClient
}