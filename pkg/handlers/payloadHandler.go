package handlers

import (
	"Orca/pkg/api"
	"Orca/pkg/scanning"
	"crypto/rsa"
	"fmt"
	"github.com/google/go-github/v33/github"
	"log"
)

type PayloadHandler struct {
	InstallationId  int64
	AppId           int
	GitHubApiClient *github.Client
	Scanner         *scanning.Scanner
}

func NewPayloadHandler(
	installationId int64,
	appId int,
	privateKey *rsa.PrivateKey,
	patternStore *scanning.PatternStore) (*PayloadHandler, error) {

	scanner, err := scanning.NewScanner(patternStore)
	if err != nil {
		return nil, err
	}

	gitHubApiClient, err := api.GetGitHubApiClient(installationId, appId, privateKey)
	if err != nil {
		return nil, err
	}

	handler := PayloadHandler{
		InstallationId:  installationId,
		AppId:           appId,
		GitHubApiClient: gitHubApiClient,
		Scanner:         scanner,
	}

	return &handler, nil
}

func (handler *PayloadHandler) HandleInstallation(installationPayload *github.InstallationEvent) {

	// Todo: Scan the repository for any sensitive information
	// 	May not be viable for large repositories with a long history
}

func (handler *PayloadHandler) HandlePush(pushPayload *github.PushEvent) {
	log.Println("Handling push...")

	// Check the commits
	commitScanResults, err := handler.Scanner.CheckPush(pushPayload, handler.GitHubApiClient)
	if err != nil {
		log.Fatal(err)
		return
	}

	// If anything shows up in the results, take action
	if len(commitScanResults) > 0 {
		log.Println("Potentially sensitive information detected. Rectifying...")
		matchHandler := NewMatchHandler(handler.GitHubApiClient)
		err := matchHandler.HandleMatchesFromPush(pushPayload, commitScanResults)
		if err != nil {
			log.Fatal(err)
			return
		}

		log.Println("Push has been addressed")
	} else {
		log.Println("No matches to address")
	}
}

func (handler *PayloadHandler) HandleIssue(issuePayload *github.IssuesEvent) {
	log.Println("Handling issue...")

	// Check the contents of the issue
	issueScanResult, err := handler.Scanner.CheckIssue(issuePayload)
	if err != nil {
		log.Fatal(err)
		return
	}

	// If anything shows up in the results, take action
	if issueScanResult.HasMatches() {
		log.Println("Potentially sensitive information detected. Rectifying...")
		matchHandler := NewMatchHandler(handler.GitHubApiClient)
		err := matchHandler.HandleMatchesFromIssue(issuePayload, issueScanResult)
		if err != nil {
			log.Fatal(err)
			return
		}

		log.Println("Issue has been addressed")
	} else {
		log.Println("No matches to address")
	}
}

func (handler *PayloadHandler) HandleIssueComment(issueCommentPayload *github.IssueCommentEvent) {
	log.Println("Handling issue...")

	// Check the contents of the comment
	issueScanResult, err := handler.Scanner.CheckIssueComment(issueCommentPayload)
	if err != nil {
		log.Fatal(err)
		return
	}

	// If anything shows up in the results, take action
	if issueScanResult.HasMatches() {
		log.Println("Potentially sensitive information detected. Rectifying...")
		matchHandler := NewMatchHandler(handler.GitHubApiClient)
		err := matchHandler.HandleMatchesFromIssueComment(issueCommentPayload, issueScanResult)
		if err != nil {
			log.Fatal(err)
			return
		}

		log.Println("Issue comment has been addressed")
	} else {
		log.Println("No matches to address")
	}
}

func (handler *PayloadHandler) HandlePullRequest(pullRequestPayload *github.PullRequestEvent) {
	log.Println("Handling pull request...")

	// Check the contents of the pull request
	pullRequestScanResult, err := handler.Scanner.CheckPullRequest(pullRequestPayload)
	if err != nil {
		log.Fatal(err)
		return
	}

	// If anything shows up in the results, take action
	if pullRequestScanResult.HasMatches() {
		log.Println("Potentially sensitive information detected. Rectifying...")
		matchHandler := NewMatchHandler(handler.GitHubApiClient)
		err := matchHandler.HandleMatchesFromPullRequest(pullRequestPayload, pullRequestScanResult)
		if err != nil {
			log.Fatal(err)
			return
		}

		log.Println("Pull request has been addressed")
	} else {
		log.Println("No matches to address")
	}
}

func (handler *PayloadHandler) HandlePullRequestReview(pullRequestReviewPayload *github.PullRequestReviewEvent) {
	fmt.Println("Handling pull request review...")

	// Check the contents of the pull request review
	pullRequestReviewScanResult, err := handler.Scanner.CheckPullRequestReview(pullRequestReviewPayload)
	if err != nil {
		log.Fatal(err)
		return
	}

	// If anything shows up in the results, take action
	if pullRequestReviewScanResult.HasMatches() {
		log.Println("Potentially sensitive information detected. Rectifying...")
		matchHandler := NewMatchHandler(handler.GitHubApiClient)
		err := matchHandler.HandleMatchesFromPullRequestReview(pullRequestReviewPayload, pullRequestReviewScanResult)
		if err != nil {
			log.Fatal(err)
			return
		}

		log.Println("Pull request review has been addressed")
	} else {
		log.Println("No matches to address")
	}
}

func (handler *PayloadHandler) HandlePullRequestReviewComment(
	pullRequestReviewCommentPayload *github.PullRequestReviewCommentEvent) {
	fmt.Println("Handling pull request review comment...")

	// Check the contents of the pull request review
	pullRequestReviewCommentScanResult, err := handler.Scanner.CheckPullRequestReviewComment(pullRequestReviewCommentPayload)
	if err != nil {
		log.Fatal(err)
		return
	}

	// If anything shows up in the results, take action
	if pullRequestReviewCommentScanResult.HasMatches() {
		log.Println("Potentially sensitive information detected. Rectifying...")
		matchHandler := NewMatchHandler(handler.GitHubApiClient)
		err := matchHandler.HandleMatchesFromPullRequestReviewComment(
			pullRequestReviewCommentPayload,
			pullRequestReviewCommentScanResult)
		if err != nil {
			log.Fatal(err)
			return
		}

		log.Println("Pull request review comment has been addressed")
	} else {
		log.Println("No matches to address")
	}
}
