package webhooks

import (
	"Orca/pkg/api"
	"Orca/pkg/handlers"
	"Orca/pkg/patterns"
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
			context, err := getHandlerContext(installation.Installation.ID, appId, privateKey)
			if err != nil {
				log.Fatal(err)
				return
			}

			handlers.HandleInstallation(installation, *context)

		case github.PushPayload:
			push := payload.(github.PushPayload)
			context, err := getHandlerContext(int64(push.Installation.ID), appId, privateKey)
			if err != nil {
				log.Fatal(err)
				return
			}
			handlers.HandlePush(push, *context)

			// Todo: Installation ID
		case github.IssuesPayload:
			issue := payload.(github.IssuesPayload)
			if issue.Action == "opened" || issue.Action == "edited" {
				context, err := getHandlerContext(0, appId, privateKey)
				if err != nil {
					log.Fatal(err)
					return
				}
				handlers.HandleIssue(issue, *context)
			}

			// Todo: Installation ID
		case github.IssueCommentPayload:
			issueComment := payload.(github.IssueCommentPayload)
			if issueComment.Action == "created" || issueComment.Action == "edited" {
				context, err := getHandlerContext(0, appId, privateKey)
				if err != nil {
					log.Fatal(err)
					return
				}
				handlers.HandleIssueComment(issueComment, *context)
			}

		case github.PullRequestPayload:
			pullRequest := payload.(github.PullRequestPayload)
			if pullRequest.Action == "opened" || pullRequest.Action == "edited" {
				context, err := getHandlerContext(pullRequest.Installation.ID, appId, privateKey)
				if err != nil {
					log.Fatal(err)
					return
				}
				handlers.HandlePullRequest(pullRequest, *context)
			}

		case github.PullRequestReviewPayload:
			pullRequestReview := payload.(github.PullRequestReviewPayload)
			if pullRequestReview.Action == "submitted" || pullRequestReview.Action == "edited" {
				context, err := getHandlerContext(pullRequestReview.Installation.ID, appId, privateKey)
				if err != nil {
					log.Fatal(err)
					return
				}
				handlers.HandlePullRequestReview(pullRequestReview, *context)
			}

		case github.PullRequestReviewCommentPayload:
			pullRequestReviewComment := payload.(github.PullRequestReviewCommentPayload)
			if pullRequestReviewComment.Action == "created" || pullRequestReviewComment.Action == "edited" {
				context, err := getHandlerContext(pullRequestReviewComment.Installation.ID, appId, privateKey)
				if err != nil {
					log.Fatal(err)
					return
				}
				handlers.HandlePullRequestReviewComment(pullRequestReviewComment, *context)
			}
		}
	})

	return nil
}

func getHandlerContext(installationId int64, appId int, privateKey rsa.PrivateKey) (*handlers.HandlerContext, error) {

	filePatterns, err := patterns.GetFilePatterns()
	if err != nil {
		return nil, err
	}

	contentPatterns, err := patterns.GetContentPatterns()
	if err != nil {
		return nil, err
	}

	gitHubAPIClient, err := api.GetGitHubApiClient(installationId, appId, privateKey)
	if err != nil {
		return nil, err
	}

	var context = handlers.HandlerContext{
		InstallationId: installationId,
		AppId: appId,
		FilePatterns: filePatterns,
		ContentPatterns: contentPatterns,
		GitHubAPIClient: gitHubAPIClient,
	}

	return &context, nil
}