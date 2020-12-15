package scanning

import (
	"Orca/pkg/payloads"
	"context"
	"encoding/base64"
	"fmt"
	gitHubApi "github.com/google/go-github/v33/github"
	"strings"
)

type CommitScanResult struct {
	Commit string
	FileMatches []FileMatch
	ContentMatches []FileContentMatch
}

func (result *CommitScanResult) HasMatches() bool {
	return len(result.FileMatches) > 0 || len(result.ContentMatches) > 0
}

type IssueScanResult struct {
	ContentMatch
}

func (scanner *Scanner) CheckPush(push payloads.PushPayload, gitHubApiClient *gitHubApi.Client) (*[]CommitScanResult, error) {

	return scanner.CheckCommits(push.Commits, gitHubApiClient)
}

func (scanner *Scanner) CheckCommits(commits []payloads.Commit, gitHubApiClient *gitHubApi.Client) (*[]CommitScanResult, error) {

	var commitScanResults []CommitScanResult
	for _, commit := range commits {

		commitScanResult, err := scanner.CheckCommit(commit, gitHubApiClient)
		if err != nil {
			return nil, err
		}

		if commitScanResult.HasMatches() {
			commitScanResults = append(commitScanResults, *commitScanResult)
		}
	}

	return &commitScanResults, nil
}

func (scanner *Scanner) CheckCommit(commit payloads.Commit, gitHubApiClient *gitHubApi.Client) (*CommitScanResult, error) {

	var commitScanResult = CommitScanResult {
		Commit: commit.ID,
	}

	var filesToCheck = append(commit.Added, commit.Modified...)

	// Check file contents
	// Todo: Is there a bulk alternative to GetContents?
	// 	Don't want to request for each file, could have a big commit
	for _, filePath := range filesToCheck {
		content, _, _, err := gitHubApiClient.Repositories.GetContents(
			context.Background(),
			commit.Repository.Owner.Login,
			commit.Repository.Name,
			filePath,
			&gitHubApi.RepositoryContentGetOptions {
				Ref: commit.ID,
			})
		if err != nil {
			return nil, err
		}

		contentBytes, err := base64.StdEncoding.DecodeString(*content.Content)
		if err != nil {
			return nil, err
		}
		contentString := string(contentBytes)

		permalinkUrl := strings.Replace(*content.HTMLURL, fmt.Sprintf("/%s/", commit.Branch), fmt.Sprintf("/%s/", commit.ID), -1)

		file := File {
			Path:    content.Path,
			Content: &contentString,
			HTMLURL: content.HTMLURL,
			PermalinkURL: &permalinkUrl,
		}

		// Check file names
		var fileScanResults = scanner.CheckFileName(file)
		if len(fileScanResults) > 0 {
			commitScanResult.FileMatches = append(commitScanResult.FileMatches, fileScanResults...)
		}

		// Search for modified or added files
		var contentScanResults = scanner.CheckFileContent(file)
		if contentScanResults.HasMatches() {
			commitScanResult.ContentMatches = append(commitScanResult.ContentMatches, contentScanResults)
		}
	}

	return &commitScanResult, nil
}

func (scanner *Scanner) CheckIssue(issue payloads.IssuePayload) IssueScanResult {
	return scanner.checkIssueBody(issue.Body)
}

func (scanner *Scanner) CheckIssueComment(issueComment payloads.IssueCommentPayload) IssueScanResult {
	return scanner.checkIssueBody(issueComment.Body)

}

func (scanner *Scanner) checkIssueBody(issueBody string) IssueScanResult {
	var issueScanResult IssueScanResult
	contentResult := scanner.checkContent(issueBody)
	if contentResult.HasMatches() {
		issueScanResult.LineMatches = contentResult.LineMatches
	}

	return issueScanResult
}