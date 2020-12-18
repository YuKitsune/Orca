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
	privateKey rsa.PrivateKey
	secret string
}

func NewWebhookHandler(
	webHookPath string,
	appId int,
	patternStore *scanning.PatternStore,
	privateKey rsa.PrivateKey,
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
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	// NOTE: github.ParseWebHook will return a pointer to the webhook payload
	//	Type switches need to switch on pointers of the desired type
	webHookType := github.WebHookType(r)
	parsedPayload, err := github.ParseWebHook(webHookType, body)
	if err != nil {
		log.Printf("Error reading payload: %v", err)
		http.Error(w, "can't read payload", http.StatusBadRequest)
		return
	}

	// TODO: Can this be automated?
	switch payload := parsedPayload.(type) {
	case *github.InstallationEvent:
		payloadHandler, err := NewPayloadHandler(
			*payload.Installation.ID,
			webHookHandler.AppId,
			webHookHandler.privateKey,
			webHookHandler.PatternStore)
		if err != nil {
			log.Fatal(err)
			return
		}

		payloadHandler.HandleInstallation(payload)

	case *github.PushEvent:
		handler, err := NewPayloadHandler(
			*payload.Installation.ID,
			webHookHandler.AppId,
			webHookHandler.privateKey,
			webHookHandler.PatternStore)
		if err != nil {
			log.Fatal(err)
			return
		}

		handler.HandlePush(payload)

	case *github.IssuesEvent:
		if *payload.Action == "opened" || *payload.Action == "edited" {
			handler, err := NewPayloadHandler(
				*payload.Installation.ID,
				webHookHandler.AppId,
				webHookHandler.privateKey,
				webHookHandler.PatternStore)
			if err != nil {
				log.Fatal(err)
				return
			}

			handler.HandleIssue(payload)
		}

	case *github.IssueCommentEvent:
		if *payload.Action == "created" || *payload.Action == "edited" {
			handler, err := NewPayloadHandler(
				*payload.Installation.ID,
				webHookHandler.AppId,
				webHookHandler.privateKey,
				webHookHandler.PatternStore)
			if err != nil {
				log.Fatal(err)
				return
			}

			handler.HandleIssueComment(payload)
		}

	case *github.PullRequestEvent:
		if *payload.Action == "opened" || *payload.Action == "edited" {
			handler, err := NewPayloadHandler(
				*payload.Installation.ID,
				webHookHandler.AppId,
				webHookHandler.privateKey,
				webHookHandler.PatternStore)
			if err != nil {
				log.Fatal(err)
				return
			}

			handler.HandlePullRequest(payload)
		}

	case *github.PullRequestReviewEvent:
		if *payload.Action == "submitted" || *payload.Action == "edited" {
			handler, err := NewPayloadHandler(
				*payload.Installation.ID,
				webHookHandler.AppId,
				webHookHandler.privateKey,
				webHookHandler.PatternStore)
			if err != nil {
				log.Fatal(err)
				return
			}

			handler.HandlePullRequestReview(payload)
		}

	case *github.PullRequestReviewCommentEvent:
		if *payload.Action == "created" || *payload.Action == "edited" {
			handler, err := NewPayloadHandler(
				*payload.Installation.ID,
				webHookHandler.AppId,
				webHookHandler.privateKey,
				webHookHandler.PatternStore)
			if err != nil {
				log.Fatal(err)
				return
			}

			handler.HandlePullRequestReviewComment(payload)
		}
	}
}