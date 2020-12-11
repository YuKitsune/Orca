package webhooks

import (
	"Orca/pkg/handlers"
	"Orca/pkg/patterns"
	"crypto/rsa"
	"gopkg.in/go-playground/webhooks.v5/github"
	"net/http"
)

func SetupHandlers(webHookPath string, privateKey rsa.PrivateKey, gitHubSecret string, appId int) error {

	hook, _ := github.New(github.Options.Secret(gitHubSecret))

	var filePatterns, filePatternsErr = patterns.GetFilePatterns()
	if filePatternsErr != nil {
		return filePatternsErr
	}

	var contentPatterns, contentPatternsErr = patterns.GetContentPatterns()
	if contentPatternsErr != nil {
		return contentPatternsErr
	}

	var context = handlers.HandlerContext{
		AppId: appId,
		FilePatterns: filePatterns,
		ContentPatterns: contentPatterns,
	}

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
		// Todo: 2. Authenticate App
		// Todo: 3. Authenticate Installation (in order to run API operations)

		// TODO: Can this be automated?
		switch payload.(type) {
		case github.InstallationPayload:
			installation := payload.(github.InstallationPayload)
			handlers.HandleInstallation(installation, context)

		case github.PushPayload:
			push := payload.(github.PushPayload)
			handlers.HandlePush(push, context)

		case github.IssuesPayload:
			issue := payload.(github.IssuesPayload)
			if issue.Action == "opened" || issue.Action == "edited" {
				handlers.HandleIssue(issue, context)
			}

		case github.IssueCommentPayload:
			issueComment := payload.(github.IssueCommentPayload)
			if issueComment.Action == "created" || issueComment.Action == "edited" {
				handlers.HandleIssueComment(issueComment, context)
			}

		case github.PullRequestPayload:
			pullRequest := payload.(github.PullRequestPayload)
			if pullRequest.Action == "opened" || pullRequest.Action == "edited" {
				handlers.HandlePullRequest(pullRequest, context)
			}

		case github.PullRequestReviewPayload:
			pullRequestReview := payload.(github.PullRequestReviewPayload)
			if pullRequestReview.Action == "submitted" || pullRequestReview.Action == "edited" {
				handlers.HandlePullRequestReview(pullRequestReview, context)
			}

		case github.PullRequestReviewCommentPayload:
			pullRequestReviewComment := payload.(github.PullRequestReviewCommentPayload)
			if pullRequestReviewComment.Action == "created" || pullRequestReviewComment.Action == "edited" {
				handlers.HandlePullRequestReviewComment(pullRequestReviewComment, context)
			}
		}
	})

	return nil
}