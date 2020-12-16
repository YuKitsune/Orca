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
	Commit  string
	Matches []FileContentMatch
}

func (result *CommitScanResult) HasMatches() bool {
	return len(result.Matches) > 0
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

		// Search for modified or added files
		contentScanResults, err := scanner.CheckFileContent(file)
		if err != nil {
			return nil, err
		}

		if contentScanResults.HasMatches() {
			commitScanResult.Matches = append(commitScanResult.Matches, *contentScanResults)
		}
	}

	return &commitScanResult, nil
}

func (scanner *Scanner) CheckIssue(issue payloads.IssuePayload) (*IssueScanResult, error) {
	return scanner.checkIssueBody(issue.Body)
}

func (scanner *Scanner) CheckIssueComment(issueComment payloads.IssueCommentPayload) (*IssueScanResult, error) {
	return scanner.checkIssueBody(issueComment.Body)

}

func (scanner *Scanner) checkIssueBody(issueBody string) (*IssueScanResult, error) {
	var issueScanResult IssueScanResult
	contentResult, err := scanner.checkContent(issueBody)
	if err != nil {
		return nil, err
	}
	if contentResult.HasMatches() {
		issueScanResult.LineMatches = contentResult.LineMatches
	}

	return &issueScanResult, nil
}