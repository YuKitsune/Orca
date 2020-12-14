package handlers

import (
	"Orca/pkg/patterns"
	"Orca/pkg/remediator"
	"Orca/pkg/scanner"
	"context"
	"encoding/base64"
	"fmt"
	gitHubAPI "github.com/google/go-github/v33/github"
	"gopkg.in/go-playground/webhooks.v5/github"
	"log"
	"strings"
)

type HandlerContext struct {
	InstallationId int64
	AppId int
	FilePatterns []patterns.SearchPattern
	ContentPatterns []patterns.SearchPattern
	GitHubAPIClient *gitHubAPI.Client
}

func HandleInstallation(installationPayload github.InstallationPayload, handlerContext HandlerContext) {

	// Todo: Scan the repository for any sensitive information
	// 	May not be viable for large repositories with a long history
}

func HandlePush(pushPayload github.PushPayload, handlerContext HandlerContext) {
	log.Println("Handling Push event")

	var commitScanResults []scanner.CommitScanResult
	for _, commit := range pushPayload.Commits {

		var commitScanResult = scanner.CommitScanResult {
			Commit: commit.ID,
		}

		var filesToCheck = append(commit.Added, commit.Modified...)

		// Check file contents
		// Todo: Is there a bulk alternative to GetContents?
		// 	Don't want to request for each file, could have a big commit
		for _, filePath := range filesToCheck {
			content, _, _, err := handlerContext.GitHubAPIClient.Repositories.GetContents(
				context.Background(),
				pushPayload.Repository.Owner.Login,
				pushPayload.Repository.Name,
				filePath,
				&gitHubAPI.RepositoryContentGetOptions {
					Ref: commit.ID,
				})
			if err != nil {
				log.Fatal(err)
				return
			}

			contentBytes, err := base64.StdEncoding.DecodeString(*content.Content)
			if err != nil {
				log.Fatal(err)
				return
			}
			contentString := string(contentBytes)

			branchName := getBranchFromRef(pushPayload.Ref)
			permalinkUrl := strings.Replace(*content.HTMLURL, fmt.Sprintf("/%s/", branchName), fmt.Sprintf("/%s/", commit.ID), -1)

			file := scanner.File {
				Path:    content.Path,
				Content: &contentString,
				HTMLURL: content.HTMLURL,
				PermalinkURL: &permalinkUrl,
			}

			// Check file names
			var fileScanResults = scanner.FindDangerousPatternsFromFile(file, handlerContext.FilePatterns)
			if len(fileScanResults) > 0 {
				commitScanResult.FileMatches = append(commitScanResult.FileMatches, fileScanResults...)
			}

			// Search for modified or added files
			var contentScanResults = scanner.ScanContentForPatterns(file, handlerContext.ContentPatterns)
			if contentScanResults.HasMatches() {
				commitScanResult.ContentMatches = append(commitScanResult.ContentMatches, contentScanResults)
			}
		}

		if commitScanResult.HasMatches() {
			commitScanResults = append(commitScanResults, commitScanResult)
		}
	}

	if len(commitScanResults) > 0 {
		err := remediator.RemediateFromPush(pushPayload, commitScanResults, handlerContext.GitHubAPIClient)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Todo: Move this somewhere more sensible
func getBranchFromRef(ref string) string {
	refSplit := strings.Split(ref, "/")
	branchName := refSplit[len(refSplit) - 1]
	return branchName
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