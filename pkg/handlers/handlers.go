package handlers

import (
	"Orca/pkg/patterns"
	"Orca/pkg/scanner"
	"fmt"
	"gopkg.in/go-playground/webhooks.v5/github"
	"log"
)

type HandlerContext struct {
	AppId int
	FilePatterns []patterns.SearchPattern
	ContentPatterns []patterns.SearchPattern
}

type CommitScanResult struct {
	Commit string
	FileMatches []scanner.FileMatch
	ContentMatches []scanner.ContentMatch
}

func HandleInstallation(installationPayload github.InstallationPayload, context HandlerContext) {

	// Todo: Scan the repository for any sensitive information
	// 	May not be viable for large repositories with a long history
}

func HandlePush(pushPayload github.PushPayload, context HandlerContext) {
	log.Println("Handling Push event")

	var commitScanResults []CommitScanResult
	for i := 0; i < len(pushPayload.Commits); i++ {

		var commit = pushPayload.Commits[i]
		var commitScanResult = CommitScanResult {
			Commit: commit.Sha,
		}

		var filesToCheck = append(commit.Added, commit.Modified...)

		// Check file names
		var fileScanResults = scanner.FindDangerousFilesForPatterns(filesToCheck, context.FilePatterns)
		if len(fileScanResults) > 0 {
			commitScanResult.FileMatches = append(commitScanResult.FileMatches, fileScanResults...)
		}

		// Todo: Check file contents

		commitScanResults = append(commitScanResults, commitScanResult)
	}

	if commitScanResults != nil && len(commitScanResults) > 0 {
		// Todo: Remediation
	}
}

func HandleIssue(issuePayload github.IssuesPayload, context HandlerContext) {

	// Todo: 1. Scan issue content
	fmt.Println("Handling new issue")
}

func HandleIssueComment(issueCommentPayload github.IssueCommentPayload, context HandlerContext) {

	// Todo: 1. Scan issue comment content
	fmt.Println("Handling issue comment")
}

func HandlePullRequest(pullRequestPayload github.PullRequestPayload, context HandlerContext) {

	// Todo: 1. Scan pull request
	// Todo: 2. Checkout tip of branch
	// Todo: 3. Scan files
	// Todo: 4. Scan any previously unscanned commits on branch
	fmt.Println("Handling pull request")
}

func HandlePullRequestReview(pullRequestReviewPayload github.PullRequestReviewPayload, context HandlerContext) {

	// Todo: 1. Scan review content
	fmt.Println("Handling pull request review")
}

func HandlePullRequestReviewComment(pullRequestReviewCommentPayload github.PullRequestReviewCommentPayload, context HandlerContext) {

	// Todo: 1. Scan review content
	fmt.Println("Handling pull request review comment")
}