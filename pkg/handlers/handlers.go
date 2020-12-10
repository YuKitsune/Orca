package handlers

import (
	"fmt"
	"gopkg.in/go-playground/webhooks.v5/github"
)

func HandlePush(pushPayload github.PushPayload) {

	// Todo: 1. Checkout tip of branch
	// Todo: 2. Scan files
	// Todo: 3. Scan any previously un-scanned commits on branch
	fmt.Println("Handling Push")
}

func HandleIssue(issuePayload github.IssuesPayload) {

	// Todo: 1. Scan issue content
	fmt.Println("Handling new issue")
}

func HandleIssueComment(issueCommentPayload github.IssueCommentPayload) {

	// Todo: 1. Scan issue comment content
	fmt.Println("Handling issue comment")
}

func HandlePullRequest(pullRequestPayload github.PullRequestPayload) {

	// Todo: 1. Scan pull request
	// Todo: 2. Checkout tip of branch
	// Todo: 3. Scan files
	// Todo: 4. Scan any previously unscanned commits on branch
	fmt.Println("Handling pull request")
}

func HandlePullRequestReview(pullRequestReviewPayload github.PullRequestReviewPayload) {

	// Todo: 1. Scan review content
	fmt.Println("Handling pull request review")
}

func HandlePullRequestReviewComment(pullRequestReviewCommentPayload github.PullRequestReviewCommentPayload) {

	// Todo: 1. Scan review content
	fmt.Println("Handling pull request review comment")
}