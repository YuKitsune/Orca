package scanning

import (
	"context"
	"github.com/google/go-github/v33/github"
)

type Result interface {
	HasMatches() bool
}

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

func (result *IssueScanResult) HasMatches() bool {
	return len(result.LineMatches) > 0
}

type PullRequestScanResult struct {
	ContentMatch
	Commits []CommitScanResult
}

func (result *PullRequestScanResult) HasMatches() bool {
	return len(result.LineMatches) > 0 || len(result.Commits) > 0
}

func (scanner *Scanner) CheckPush(push *github.PushEvent, githubClient *github.Client) ([]CommitScanResult, error) {

	var commitScanResults []CommitScanResult
	for _, commit := range push.Commits {

		var commitScanResult = CommitScanResult {
			Commit: *commit.ID,
		}

		// Only want to check added or modified files
		// Deleted files should already have been checked before hand
		var filesToCheck = append(commit.Added, commit.Modified...)

		// Check file contents
		for _, filePath := range filesToCheck {
			contentScanResults, err := scanner.CheckFileContentFromCommit(
				githubClient,
				push.Repo.Owner.Login,
				push.Repo.Name,
				commit.ID,
				&filePath)
			if err != nil {
				return nil, err
			}

			if len(contentScanResults.LineMatches) > 0 {
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

func (scanner *Scanner) checkIssueBody(body *string) (*IssueScanResult, error) {
	result, err := scanner.CheckTextBody(body)
	if err != nil {
		return nil, err
	}

	if len(result.LineMatches) > 1 {
		issueResult := IssueScanResult{*result}
		return &issueResult, nil
	}

	return nil, nil
}

func (scanner *Scanner) CheckPullRequest(pullRequest *github.PullRequestEvent, githubClient *github.Client) (*PullRequestScanResult, error) {

	// Check the Pull Request body
	contentMatch, err := scanner.CheckTextBody(pullRequest.PullRequest.Body)
	if err != nil {
		return nil, err
	}

	// Get a list of the commits on the PRs branch
	commits, _, err := githubClient.Repositories.ListCommits(
		context.Background(),
		*pullRequest.Repo.Owner.Login,
		*pullRequest.Repo.Name,
		&github.CommitsListOptions{

			// TODO: this will get all commits on the PRs branch, but this won't be very good if we're making a PR from
			// 	master into something else
			SHA: *pullRequest.PullRequest.Head.Ref,
		})

	// Check each (added, modified, or removed) file in each commit
	var commitScanResults []CommitScanResult
	for _, commit := range commits {

		commitScanResult := CommitScanResult { Commit: *commit.SHA }
		for _, file := range commit.Files {
			fileContentMatch, err := scanner.CheckFileContentFromCommit(
				githubClient,
				pullRequest.Repo.Owner.Login,
				pullRequest.Repo.Name,
				commit.Commit.SHA,
				file.Filename)
			if err != nil {
				return nil, err
			}

			if len(fileContentMatch.LineMatches) > 0 {
				commitScanResult.Matches = append(commitScanResult.Matches, *fileContentMatch)
			}
		}

		if commitScanResult.HasMatches() {
			commitScanResults = append(commitScanResults, commitScanResult)
		}
	}

	result := PullRequestScanResult {
		ContentMatch: *contentMatch,
		Commits: commitScanResults,
	}

	return &result, nil
}

func (scanner *Scanner) CheckTextBody(body *string) (*ContentMatch, error) {
	var result ContentMatch
	contentResult, err := scanner.checkContent(body)
	if err != nil {
		return nil, err
	}
	if len(contentResult.LineMatches) > 0 {
		result.LineMatches = contentResult.LineMatches
	}

	return &result, nil
}
