package webhooks

import (
	"Orca/pkg/handlers"
	"crypto/rsa"
	"gopkg.in/go-playground/webhooks.v5/github"
	"net/http"
)

func SetupHandlers(webHookPath string, privateKey rsa.PrivateKey, gitHubSecret string, appId int) {

	hook, _ := github.New(github.Options.Secret(gitHubSecret))

	http.HandleFunc(webHookPath, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(
			r,
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
		// Todo: 2. Authenticate App
		// Todo: 3. Authenticate Installation (in order to run API operations)

		// TODO: Can this be automated?
		switch payload.(type) {
		case github.PushPayload:
			push := payload.(github.PushPayload)
			handlers.HandlePush(push)

		case github.IssuesPayload:
			issue := payload.(github.IssuesPayload)
			if issue.Action == "opened" || issue.Action == "edited" {
				handlers.HandleIssue(issue)
			}

		case github.IssueCommentPayload:
			issueComment := payload.(github.IssueCommentPayload)
			if issueComment.Action == "created" || issueComment.Action == "edited" {
				handlers.HandleIssueComment(issueComment)
			}

		case github.PullRequestPayload:
			pullRequest := payload.(github.PullRequestPayload)
			if pullRequest.Action == "opened" || pullRequest.Action == "edited" {
				handlers.HandlePullRequest(pullRequest)
			}

		case github.PullRequestReviewPayload:
			pullRequestReview := payload.(github.PullRequestReviewPayload)
			if pullRequestReview.Action == "submitted" || pullRequestReview.Action == "edited" {
				handlers.HandlePullRequestReview(pullRequestReview)
			}

		case github.PullRequestReviewCommentPayload:
			pullRequestReviewComment := payload.(github.PullRequestReviewCommentPayload)
			if pullRequestReviewComment.Action == "created" || pullRequestReviewComment.Action == "edited" {
				handlers.HandlePullRequestReviewComment(pullRequestReviewComment)
			}
		}
	})
}