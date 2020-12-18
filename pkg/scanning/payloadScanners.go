package scanning

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/google/go-github/v33/github"
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

func (scanner *Scanner) CheckPush(push *github.PushEvent, githubClient *github.Client) ([]CommitScanResult, error) {

	var commitScanResults []CommitScanResult
	for _, commit := range push.Commits {

		var commitScanResult = CommitScanResult {
			Commit: *commit.ID,
		}

		var filesToCheck = append(commit.Added, commit.Modified...)

		// Check file contents
		// Todo: Is there a bulk alternative to GetContents?
		// 	Don't want to request for each file, could have a big commit
		for _, filePath := range filesToCheck {
			content, _, _, err := githubClient.Repositories.GetContents(
				context.Background(),
				*push.Repo.Owner.Login,
				*push.Repo.Name,
				filePath,
				&github.RepositoryContentGetOptions {
					Ref: *commit.ID,
				})
			if err != nil {
				return nil, err
			}

			contentBytes, err := base64.StdEncoding.DecodeString(*content.Content)
			if err != nil {
				return nil, err
			}
			contentString := string(contentBytes)

			permalinkUrl := strings.Replace(*content.HTMLURL, fmt.Sprintf("/%s/", *push.Ref), fmt.Sprintf("/%s/", *commit.ID), -1)

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

		if commitScanResult.HasMatches() {
			commitScanResults = append(commitScanResults, commitScanResult)
		}
	}

	return commitScanResults, nil
}

func (scanner *Scanner) CheckIssue(issue *github.IssuesEvent) (*IssueScanResult, error) {
	return scanner.checkIssueBody(issue.Issue.Body)
}

func (scanner *Scanner) CheckIssueComment(issueComment *github.IssueCommentEvent) (*IssueScanResult, error) {
	return scanner.checkIssueBody(issueComment.Comment.Body)
}

func (scanner *Scanner) checkIssueBody(issueBody *string) (*IssueScanResult, error) {
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