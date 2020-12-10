package main

import (
	"Orca/pkg/crypto"
	"Orca/pkg/webhooks"
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"net/http"
	"os"
)

func main() {

	var path string
	var port int
	var privateKeyFile string
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
				Required: true,
				Usage:   "Path to listen for WebHook requests", // Todo: Fix wording
				Destination: &path,
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"pt"},
				Value: 80,
				Required: true,
				Usage:   "Port to listen on", // Todo: Fix wording
				Destination: &port,
			},
			&cli.StringFlag{
				Name:    "private-key-file",
				Aliases: []string{"pk"},
				Required: true,
				EnvVars: []string{"GITHUB_ORCA_PRIVATE_KEY"},
				Usage:   "GitHub private key for the Orca app. Expects that the private key in PEM format. Converts the newlines",
				Destination: &privateKeyFile,
			},
			&cli.StringFlag{
				Name:    "secret",
				Aliases: []string{"s"},
				Required: true,
				EnvVars: []string{"GITHUB_ORCA_WEBHOOK_SECRET"},
				Usage:   "The secret is used to verify that webhooks are sent by GitHub.",
				Destination: &secret,
			},
			&cli.IntFlag{
				Name:    "app-id",
				Aliases: []string{"id"},
				Required: true,
				EnvVars: []string{"GITHUB_ORCA_APP_ID"},
				Usage:   "The GitHub App's identifier (type integer) set when registering an app.",
				Destination: &appId,
			},
		},
		Action: func(c *cli.Context) error {

			// Get the private key
			// BUG: Private key file won't parse correctly
			var privateKey, certErr = crypto.DecodePrivateKeyFromFile(privateKeyFile)
			if certErr != nil {
				return certErr
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