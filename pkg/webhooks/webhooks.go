package webhooks

import (
	"Orca/pkg/handlers"
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

func (webHookHandler *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {

	// Gets the body as bytes and validates the signature
	body, err := github.ValidatePayload(r, []byte(webHookHandler.secret))
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	webHookType := github.WebHookType(r)
	payload, err := github.ParseWebHook(webHookType, body)

	// TODO: Can this be automated?
	switch payload.(type) {
	case github.InstallationEvent:
		installation := payload.(github.InstallationEvent)
		handler, err := handlers.NewPayloadHandler(
			*installation.Installation.ID,
			webHookHandler.AppId,
			webHookHandler.privateKey,
			webHookHandler.PatternStore)
		if err != nil {
			log.Fatal(err)
			return
		}

		handler.HandleInstallation(installation)

	case github.PushEvent:
		push := payload.(github.PushEvent)
		handler, err := handlers.NewPayloadHandler(
			*push.Installation.ID,
			webHookHandler.AppId,
			webHookHandler.privateKey,
			webHookHandler.PatternStore)
		if err != nil {
			log.Fatal(err)
			return
		}

		handler.HandlePush(push)

	case github.IssuesEvent:
		issue := payload.(github.IssuesEvent)
		if *issue.Action == "opened" || *issue.Action == "edited" {
			handler, err := handlers.NewPayloadHandler(
				*issue.Installation.ID,
				webHookHandler.AppId,
				webHookHandler.privateKey,
				webHookHandler.PatternStore)
			if err != nil {
				log.Fatal(err)
				return
			}

			handler.HandleIssue(issue)
		}

	case github.IssueCommentEvent:
		issueComment := payload.(github.IssueCommentEvent)
		if *issueComment.Action == "created" || *issueComment.Action == "edited" {
			handler, err := handlers.NewPayloadHandler(
				*issueComment.Installation.ID,
				webHookHandler.AppId,
				webHookHandler.privateKey,
				webHookHandler.PatternStore)
			if err != nil {
				log.Fatal(err)
				return
			}

			handler.HandleIssueComment(issueComment)
		}

	case github.PullRequestEvent:
		pullRequest := payload.(github.PullRequestEvent)
		if *pullRequest.Action == "opened" || *pullRequest.Action == "edited" {
			handler, err := handlers.NewPayloadHandler(
				*pullRequest.Installation.ID,
				webHookHandler.AppId,
				webHookHandler.privateKey,
				webHookHandler.PatternStore)
			if err != nil {
				log.Fatal(err)
				return
			}

			handler.HandlePullRequest(pullRequest)
		}

	case github.PullRequestReviewEvent:
		pullRequestReview := payload.(github.PullRequestReviewEvent)
		if *pullRequestReview.Action == "submitted" || *pullRequestReview.Action == "edited" {
			handler, err := handlers.NewPayloadHandler(
				*pullRequestReview.Installation.ID,
				webHookHandler.AppId,
				webHookHandler.privateKey,
				webHookHandler.PatternStore)
			if err != nil {
				log.Fatal(err)
				return
			}

			handler.HandlePullRequestReview(pullRequestReview)
		}

	case github.PullRequestReviewCommentEvent:
		pullRequestReviewComment := payload.(github.PullRequestReviewCommentEvent)
		if *pullRequestReviewComment.Action == "created" || *pullRequestReviewComment.Action == "edited" {
			handler, err := handlers.NewPayloadHandler(
				*pullRequestReviewComment.Installation.ID,
				webHookHandler.AppId,
				webHookHandler.privateKey,
				webHookHandler.PatternStore)
			if err != nil {
				log.Fatal(err)
				return
			}

			handler.HandlePullRequestReviewComment(pullRequestReviewComment)
		}
	}
}