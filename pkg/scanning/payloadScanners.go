package scanning

import (
	"github.com/google/go-github/v33/github"
	"sort"
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
	Matches []LineMatch
}

func (result *IssueScanResult) HasMatches() bool {
	return result != nil && len(result.Matches) > 0
}

type PullRequestScanResult struct {
	Matches []LineMatch
}

func (result *PullRequestScanResult) HasMatches() bool {
	return len(result.Matches) > 0
}

type PullRequestReviewScanResult struct {
	Matches []LineMatch
}

func (result *PullRequestReviewScanResult) HasMatches() bool {
	return len(result.Matches) > 0
}

type PullRequestReviewCommentScanResult struct {
	Matches []LineMatch
}

func (result *PullRequestReviewCommentScanResult) HasMatches() bool {
	return len(result.Matches) > 0
}

func (scanner *Scanner) CheckPush(push *github.PushEvent, githubClient *github.Client) ([]CommitScanResult, error) {

	// Sort the commits by their date
	sort.Slice(push.Commits, func(i, j int) bool {
		return push.Commits[i].Timestamp.Unix() < push.Commits[j].Timestamp.Unix()
	})

	// Get a list of commit SHAs
	var commitSHAs []string
	for _, commit := range push.Commits {
		commitSHAs = append(commitSHAs, *commit.ID)
	}

	return scanner.CheckCommits(push.Repo.Owner.Login, push.Repo.Name, githubClient, commitSHAs)
}

func (scanner *Scanner) CheckIssue(issue *github.IssuesEvent) (*IssueScanResult, error) {

	// Check the Issue body
	matches, err := scanner.CheckContent(issue.Issue.Body)
	if err != nil {
		return nil, err
	}

	result := IssueScanResult{
		Matches: matches,
	}

	return &result, nil
}

func (scanner *Scanner) CheckIssueComment(issueComment *github.IssueCommentEvent) (*IssueScanResult, error) {

	// Check the Issue Comment body
	matches, err := scanner.CheckContent(issueComment.Comment.Body)
	if err != nil {
		return nil, err
	}

	result := IssueScanResult{
		Matches: matches,
	}

	return &result, nil
}

func (scanner *Scanner) CheckPullRequest(pullRequest *github.PullRequestEvent) (*PullRequestScanResult, error) {

	// NOTE: commits are checked via a CI check, see checkSuiteHandler.go

	// Check the Pull Request body
	matches, err := scanner.CheckContent(pullRequest.PullRequest.Body)
	if err != nil {
		return nil, err
	}

	result := PullRequestScanResult{
		Matches: matches,
	}

	return &result, nil
}

func (scanner *Scanner) CheckPullRequestReview(
	pullRequestReview *github.PullRequestReviewEvent) (*PullRequestReviewScanResult, error) {

	// Check the Pull Request Review body
	matches, err := scanner.CheckContent(pullRequestReview.Review.Body)
	if err != nil {
		return nil, err
	}

	result := PullRequestReviewScanResult{
		Matches: matches,
	}

	return &result, nil
}

func (scanner *Scanner) CheckPullRequestReviewComment(
	pullRequestReviewComment *github.PullRequestReviewCommentEvent) (*PullRequestReviewCommentScanResult, error) {

	// Check the Pull Request Review Comment body
	matches, err := scanner.CheckContent(pullRequestReviewComment.Comment.Body)
	if err != nil {
		return nil, err
	}

	result := PullRequestReviewCommentScanResult{
		Matches: matches,
	}

	return &result, nil
}
