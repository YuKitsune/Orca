package main

import (
	"Orca/pkg/crypto"
	"Orca/pkg/webhooks"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {

	var path string
	var port int
	var privateKeyFile string
	var privateKeyString string
	var secret string
	var appId int

	app := &cli.App{
		Name: "Orca",
		Usage: "A GitHub App that hunts for potential credentials in GitHub repositories, issues and pull requests.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "path",
				Aliases: []string{"ph"},
				Value: "/webhooks",
				EnvVars: []string{"ORCA_PATH"},
				Usage:   "Path to listen for WebHook requests", // Todo: Fix wording
				Destination: &path,
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"pt"},
				Value: 80,
				EnvVars: []string{"ORCA_PORT"},
				Usage:   "Port to listen on", // Todo: Fix wording
				Destination: &port,
			},
			&cli.StringFlag{
				Name:    "private-key-file",
				Aliases: []string{"pkf"},
				Usage:   "GitHub private key file for the Orca app. Expects that the private key in PEM format.",
				Destination: &privateKeyFile,
			},
			&cli.StringFlag{
				Name:    "private-key",
				Aliases: []string{"pk"},
				EnvVars: []string{"GITHUB_ORCA_PRIVATE_KEY"},
				Usage:   "GitHub private key for the Orca app. Expects that the private key in PEM format.",
				Destination: &privateKeyString,
			},
			&cli.StringFlag{
				Name:    "secret",
				Aliases: []string{"s"},
				EnvVars: []string{"GITHUB_ORCA_WEBHOOK_SECRET"},
				Usage:   "The secret is used to verify that webhooks are sent by GitHub.",
				Destination: &secret,
			},
			&cli.IntFlag{
				Name:    "app-id",
				Aliases: []string{"id"},
				EnvVars: []string{"GITHUB_ORCA_APP_ID"},
				Usage:   "The GitHub App's identifier (type integer) set when registering an app.",
				Destination: &appId,
			},
		},
		Action: func(c *cli.Context) error {

			// Check the webhook path
			if len(path) < 1 {
				return errors.New("a webhook path must be provided")
			}

			// Check the port number
			if port > 65535 || port < 1 {
				return errors.New("a valid port number must be provided")
			}

			// Check and decode the private key
			var privateKey *rsa.PrivateKey
			var keyErr error
			if len(privateKeyFile) > 0 {
				privateKey, keyErr = crypto.DecodePrivateKeyFromFile(privateKeyFile)
			} else if len(privateKeyString) > 0 {
				// Replace escaped newlines with actual new lines
				// Todo: I wonder if this is a bad idea? GitHubs own ruby template does this so it should be fine
				var privateKeyWithNewLines = strings.ReplaceAll(privateKeyString, "\\n", "\n")
				var privateKeyBytes = []byte(privateKeyWithNewLines)
				privateKey, keyErr = crypto.DecodePrivateKey(privateKeyBytes)
			} else {
				return errors.New("a private key must be provided")
			}

			if keyErr != nil {
				return keyErr
			}

			// Check the secret
			if len(secret) < 1 {
				return errors.New("a secret is must be provided")
			}

			// Check the app ID
			if appId < 1 {
				return errors.New("an app id must be provided")
			}

			// Setup webhook handlers
			webhooks.SetupHandlers(path, *privateKey, secret, appId)

			// Start HTTP webhooks
			log.Printf("Starting webhooks at port %d\n", port)
			var address = fmt.Sprintf(":%d", port)
			if err := http.ListenAndServe(address, nil); err != nil {
				return err
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}