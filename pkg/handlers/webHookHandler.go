package handlers

import (
	"Orca/pkg/scanning"
	"crypto/rsa"
	"github.com/google/go-github/v33/github"
	"log"
	"net/http"
)

type WebhookHandler struct {
	Path string
	AppId int
	PatternStore *scanning.PatternStore
	privateKey *rsa.PrivateKey
	secret string
}

func NewWebhookHandler(
	webHookPath string,
	appId int,
	patternStore *scanning.PatternStore,
	privateKey *rsa.PrivateKey,
	gitHubSecret string) *WebhookHandler {
	handler := WebhookHandler{
		Path:         webHookPath,
		AppId:        appId,
		PatternStore: patternStore,
		privateKey:   privateKey,
		secret:       gitHubSecret,
	}

	return &handler
}

func (webHookHandler *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Gets the body as bytes and validates the signature
	body, err := github.ValidatePayload(r, []byte(webHookHandler.secret))
	if err != nil {
		webHookHandler.HandleError(w, err)
	}

	// NOTE: github.ParseWebHook will return a pointer to the webhook payload
	//	Type switches need to switch on pointers of the desired type otherwise they won't work
	webHookType := github.WebHookType(r)
	parsedPayload, err := github.ParseWebHook(webHookType, body)
	if err != nil {
		webHookHandler.HandleError(w, err)
	}

	// TODO: Can this be automated?
	switch payload := parsedPayload.(type) {
	case *github.InstallationEvent:
		payloadHandler, err := webHookHandler.MakePayloadHandler(payload.Installation.ID)
		if err != nil {
			webHookHandler.HandleError(w, err)
		}

		payloadHandler.HandleInstallation(payload)

	case *github.PushEvent:
		payloadHandler, err := webHookHandler.MakePayloadHandler(payload.Installation.ID)
		if err != nil {
			webHookHandler.HandleError(w, err)
		}

		payloadHandler.HandlePush(payload)

	case *github.IssuesEvent:
		if *payload.Action == "opened" || *payload.Action == "edited" {
			payloadHandler, err := webHookHandler.MakePayloadHandler(payload.Installation.ID)
			if err != nil {
				webHookHandler.HandleError(w, err)
			}

			payloadHandler.HandleIssue(payload)
		}

	case *github.IssueCommentEvent:
		if *payload.Action == "created" || *payload.Action == "edited" {
			payloadHandler, err := webHookHandler.MakePayloadHandler(payload.Installation.ID)
			if err != nil {
				webHookHandler.HandleError(w, err)
			}

			payloadHandler.HandleIssueComment(payload)
		}

	case *github.PullRequestEvent:
		if *payload.Action == "opened" || *payload.Action == "edited" {
			payloadHandler, err := webHookHandler.MakePayloadHandler(payload.Installation.ID)
			if err != nil {
				webHookHandler.HandleError(w, err)
			}

			payloadHandler.HandlePullRequest(payload)
		}

	case *github.PullRequestReviewEvent:
		if *payload.Action == "submitted" || *payload.Action == "edited" {
			payloadHandler, err := webHookHandler.MakePayloadHandler(payload.Installation.ID)
			if err != nil {
				webHookHandler.HandleError(w, err)
			}

			payloadHandler.HandlePullRequestReview(payload)
		}

	case *github.PullRequestReviewCommentEvent:
		if *payload.Action == "created" || *payload.Action == "edited" {
			payloadHandler, err := webHookHandler.MakePayloadHandler(payload.Installation.ID)
			if err != nil {
				webHookHandler.HandleError(w, err)
			}

			payloadHandler.HandlePullRequestReviewComment(payload)
		}
	}
}

func (webHookHandler *WebhookHandler) MakePayloadHandler(installationId *int64) (*PayloadHandler, error) {
	payloadHandler, err := NewPayloadHandler(*installationId, webHookHandler.AppId, webHookHandler.privateKey, webHookHandler.PatternStore)
	if err != nil {
		return nil, err
	}

	return payloadHandler, nil
}

func (webHookHandler *WebhookHandler) HandleError(w http.ResponseWriter, err error) {
	http.Error(w, "failed to handle payload", http.StatusBadRequest)
	log.Fatalf("Error handling WebHook request: %v", err)
}
