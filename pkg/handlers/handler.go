package handlers

import (
	"Orca/pkg/api"
	"Orca/pkg/payloads"
	"Orca/pkg/rectifier"
	"Orca/pkg/scanning"
	"crypto/rsa"
	"fmt"
	gitHubApi "github.com/google/go-github/v33/github"
	"gopkg.in/go-playground/webhooks.v5/github"
	"log"
)

type Handler struct {
	InstallationId  int64
	AppId           int
	GitHubApiClient *gitHubApi.Client
	Scanner 		*scanning.Scanner
}

func NewHandler(installationId int64, appId int, privateKey rsa.PrivateKey) (*Handler, error) {

	scanner, err := scanning.NewScanner()
	if err != nil {
		return nil, err
	}

	gitHubApiClient, err := api.GetGitHubApiClient(installationId, appId, privateKey)
	if err != nil {
		return nil, err
	}

	handler := Handler {
		InstallationId:  installationId,
		AppId:           appId,
		GitHubApiClient: gitHubApiClient,
		Scanner: scanner,
	}

	return &handler, nil
}

func (handler *Handler) HandleInstallation(installationPayload github.InstallationPayload) {

	// Todo: Scan the repository for any sensitive information
	// 	May not be viable for large repositories with a long history
}

// Todo: Move payload conversion outside of this file

func (handler *Handler) HandlePush(pushPayload github.PushPayload) {
	log.Println("Handling Push event")

	// Convert the payload
	payload := payloads.NewPushPayload(pushPayload)

	// Check the commits
	commitScanResults, err := handler.Scanner.CheckPush(payload, handler.GitHubApiClient)
	if err != nil {
		log.Fatal(err)
		return
	}

	// If anything shows up in the results, take action
	if len(*commitScanResults) > 0 {
		rectifier := rectifier.NewRectifier(handler.GitHubApiClient)
		err := rectifier.RemediateFromPush(payload, *commitScanResults)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (handler *Handler) HandleIssue(issuePayload github.IssuesPayload) {

	// Convert the payload
	payload := payloads.NewIssuePayload(issuePayload)

	// Check the contents of the issue
	issueScanResult := handler.Scanner.CheckIssue(payload)

	// If anything shows up in the results, take action
	if issueScanResult.HasMatches() {
		rectifier := rectifier.NewRectifier(handler.GitHubApiClient)
		err := rectifier.RemediateFromIssue(payload, issueScanResult)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (handler *Handler) HandleIssueComment(issueCommentPayload github.IssueCommentPayload) {

	// Convert the payload
	payload := payloads.NewIssueCommentPayload(issueCommentPayload)

	// Check the contents of the comment
	issueScanResult := handler.Scanner.CheckIssueComment(payload)

	// If anything shows up in the results, take action
	if issueScanResult.HasMatches() {
		rectifier := rectifier.NewRectifier(handler.GitHubApiClient)
		err := rectifier.RemediateFromIssueComment(payload, issueScanResult)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (handler *Handler) HandlePullRequest(pullRequestPayload github.PullRequestPayload) {

	fmt.Println("Handling pull request")
	// Todo: 1. Scan pull request
	// Todo: 2. Checkout tip of branch
	// Todo: 3. Scan files
	// Todo: 4. Scan any previously un-scanned commits on branch
}

func (handler *Handler) HandlePullRequestReview(pullRequestReviewPayload github.PullRequestReviewPayload) {

	fmt.Println("Handling pull request review")
	// Todo: 1. Scan review content
}

func (handler *Handler) HandlePullRequestReviewComment(pullRequestReviewCommentPayload github.PullRequestReviewCommentPayload) {

	fmt.Println("Handling pull request review comment")
	// Todo: 1. Scan review content
}