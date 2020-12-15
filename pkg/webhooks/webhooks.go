package webhooks

import (
	"Orca/pkg/handlers"
	"crypto/rsa"
	"gopkg.in/go-playground/webhooks.v5/github"
	"log"
	"net/http"
)

func SetupHandlers(webHookPath string, privateKey rsa.PrivateKey, gitHubSecret string, appId int) error {

	hook, _ := github.New(github.Options.Secret(gitHubSecret))

	http.HandleFunc(webHookPath, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(
			r,
			github.InstallationEvent,
			github.PushEvent,
			github.IssuesEvent,
			github.IssueCommentEvent,
			github.PullRequestEvent,
			github.PullRequestReviewEvent,
			github.PullRequestReviewCommentEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				// Event wasn't one of the ones asked to be parsed
			}
		}

		// Todo: 1. Verify webhook signature

		// Note:
		// 	getHandlerContext is invoked within each case as it requires an installation ID which is not known until
		//  the payload has been parsed
		//  If it can be moved outside of the switch, that would be nice

		// TODO: Can this be automated?
		switch payload.(type) {
		case github.InstallationPayload:
			installation := payload.(github.InstallationPayload)
			handler, err := handlers.NewHandler(installation.Installation.ID, appId, privateKey)
			if err != nil {
				log.Fatal(err)
				return
			}

			handler.HandleInstallation(installation)

		case github.PushPayload:
			push := payload.(github.PushPayload)
			handler, err := handlers.NewHandler(int64(push.Installation.ID), appId, privateKey)
			if err != nil {
				log.Fatal(err)
				return
			}

			handler.HandlePush(push)

			// Todo: The go-playground/webhooks package didn't include the Installation ID with Issues for some reason
			// 	Will need to make a PR and remove the hard-coded 0 once it's been merged
		case github.IssuesPayload:
			issue := payload.(github.IssuesPayload)
			if issue.Action == "opened" || issue.Action == "edited" {
				handler, err := handlers.NewHandler(0, appId, privateKey)
				if err != nil {
					log.Fatal(err)
					return
				}

				handler.HandleIssue(issue)
			}

		case github.IssueCommentPayload:
			issueComment := payload.(github.IssueCommentPayload)
			if issueComment.Action == "created" || issueComment.Action == "edited" {
				handler, err := handlers.NewHandler(0, appId, privateKey)
				if err != nil {
					log.Fatal(err)
					return
				}

				handler.HandleIssueComment(issueComment)
			}

		case github.PullRequestPayload:
			pullRequest := payload.(github.PullRequestPayload)
			if pullRequest.Action == "opened" || pullRequest.Action == "edited" {
				handler, err := handlers.NewHandler(pullRequest.Installation.ID, appId, privateKey)
				if err != nil {
					log.Fatal(err)
					return
				}

				handler.HandlePullRequest(pullRequest)
			}

		case github.PullRequestReviewPayload:
			pullRequestReview := payload.(github.PullRequestReviewPayload)
			if pullRequestReview.Action == "submitted" || pullRequestReview.Action == "edited" {
				handler, err := handlers.NewHandler(pullRequestReview.Installation.ID, appId, privateKey)
				if err != nil {
					log.Fatal(err)
					return
				}

				handler.HandlePullRequestReview(pullRequestReview)
			}

		case github.PullRequestReviewCommentPayload:
			pullRequestReviewComment := payload.(github.PullRequestReviewCommentPayload)
			if pullRequestReviewComment.Action == "created" || pullRequestReviewComment.Action == "edited" {
				handler, err := handlers.NewHandler(pullRequestReviewComment.Installation.ID, appId, privateKey)
				if err != nil {
					log.Fatal(err)
					return
				}

				handler.HandlePullRequestReviewComment(pullRequestReviewComment)
			}
		}
	})

	return nil
}