package handlers

import (
	"Orca/pkg/patterns"
	"Orca/pkg/scanner"
	"context"
	"fmt"
	gitHubAPI "github.com/google/go-github/v33/github"
	"gopkg.in/go-playground/webhooks.v5/github"
	"log"
)

type HandlerContext struct {
	InstallationId int64
	AppId int
	FilePatterns []patterns.SearchPattern
	ContentPatterns []patterns.SearchPattern
	GitHubAPIClient *gitHubAPI.Client
}

type CommitScanResult struct {
	Commit string
	FileMatches []scanner.FileMatch
	ContentMatches []scanner.ContentMatch
}

func HandleInstallation(installationPayload github.InstallationPayload, handlerContext HandlerContext) {

	// Todo: Scan the repository for any sensitive information
	// 	May not be viable for large repositories with a long history
}

func HandlePush(pushPayload github.PushPayload, handlerContext HandlerContext) {
	log.Println("Handling Push event")

	var commitScanResults []CommitScanResult
	for _, commit := range pushPayload.Commits {

		var commitScanResult = CommitScanResult {
			Commit: commit.Sha,
		}

		var filesToCheck = append(commit.Added, commit.Modified...)

		// Check file names
		var fileScanResults = scanner.FindDangerousFilesForPatterns(filesToCheck, handlerContext.FilePatterns)
		if len(fileScanResults) > 0 {
			commitScanResult.FileMatches = append(commitScanResult.FileMatches, fileScanResults...)
		}

		// Check file contents
		for _, file := range filesToCheck {
			content, _, _, err := handlerContext.GitHubAPIClient.Repositories.GetContents(
				context.Background(),
				pushPayload.Repository.Owner.Login,
				pushPayload.Repository.Name,
				file,
				&gitHubAPI.RepositoryContentGetOptions {
					Ref: commit.Sha,
				})
			if err != nil {
				log.Fatal(err)
				return
			}

			// Search for modified or added files
			var contentScanResults = scanner.ScanContentForPatterns(*content.Content, handlerContext.ContentPatterns)
			if len(contentScanResults.LineMatches) > 0 {
				commitScanResult.ContentMatches = append(commitScanResult.ContentMatches, contentScanResults)
			}
		}

		commitScanResults = append(commitScanResults, commitScanResult)
	}

	if commitScanResults != nil && len(commitScanResults) > 0 {
		// Todo: Remediation
	}
}

func HandleIssue(issuePayload github.IssuesPayload, handlerContext HandlerContext) {

	fmt.Println("Handling new issue")
	// Todo: 1. Scan issue content
	// Todo: 2. Scan attachments
	// Todo: 3. Redact any credentials from issue contents
	// Todo: 4. Modify attachments?
}

func HandleIssueComment(issueCommentPayload github.IssueCommentPayload, handlerContext HandlerContext) {

	fmt.Println("Handling issue comment")
	// Todo: 1. Scan issue comment
	// Todo: 2. Scan comment attachments
	// Todo: 3. Redact any credentials from issue comment
	// Todo: 4. Modify attachments?
}

func HandlePullRequest(pullRequestPayload github.PullRequestPayload, handlerContext HandlerContext) {

	fmt.Println("Handling pull request")
	// Todo: 1. Scan pull request
	// Todo: 2. Checkout tip of branch
	// Todo: 3. Scan files
	// Todo: 4. Scan any previously un-scanned commits on branch
}

func HandlePullRequestReview(pullRequestReviewPayload github.PullRequestReviewPayload, handlerContext HandlerContext) {

	fmt.Println("Handling pull request review")
	// Todo: 1. Scan review content
}

func HandlePullRequestReviewComment(pullRequestReviewCommentPayload github.PullRequestReviewCommentPayload, handlerContext HandlerContext) {

	fmt.Println("Handling pull request review comment")
	// Todo: 1. Scan review content
}