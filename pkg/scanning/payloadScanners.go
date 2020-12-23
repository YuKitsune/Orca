package scanning

import (
	"context"
	"github.com/google/go-github/v33/github"
	"log"
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
	return result != nil && len(result.LineMatches) > 0
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
	commits, _, err := githubClient.PullRequests.ListCommits(
		context.Background(),
		*pullRequest.Repo.Owner.Login,
		*pullRequest.Repo.Name,
		*pullRequest.PullRequest.Number,
		nil)

	// Check each (added, modified, or removed) file in each commit
	commitScanResults, err := scanner.CheckCommits(pullRequest.Repo, githubClient, commits)
	if err != nil {
		return nil, err
	}

	result := PullRequestScanResult{
		ContentMatch: *contentMatch,
		Commits:      commitScanResults,
	}

	return &result, nil
}

func (scanner *Scanner) CheckCommits(
	repo *github.Repository,
	githubClient *github.Client,
	commits []*github.RepositoryCommit) ([]CommitScanResult, error) {

	var commitScanResults []CommitScanResult
	for _, commit := range commits {

		// NOTE: ListCommits does not include any references to which files were changed (commit.Files is always nil),
		//	so we need to send another request specifically for the commit
		// TODO: Find a way around this to prevent getting rate limited
		commitScanResult := CommitScanResult{Commit: *commit.SHA}
		commitWithFiles, _, err := githubClient.Repositories.GetCommit(
			context.Background(),
			*repo.Owner.Login,
			*repo.Name,
			*commit.SHA)
		if err != nil {
			return nil, err
		}

		for _, file := range commitWithFiles.Files {

			// Only care about added and modified files
			if *file.Status != "added" && *file.Status != "modified" {
				continue
			}

			log.Printf("Checking %s from %s", *file.Filename, *commit.SHA)

			fileContentMatch, err := scanner.CheckFileContentFromCommit(
				githubClient,
				repo.Owner.Login,
				repo.Name,
				commit.SHA,
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
	return commitScanResults, nil
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
