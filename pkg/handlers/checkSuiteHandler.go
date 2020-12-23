package handlers

import (
	"context"
	"fmt"
	"github.com/google/go-github/v33/github"
	"log"
)

type checkRunStatus string
type checkRunConclusion string

const (
	checkRunStatusInProgress  checkRunStatus     = "in_progress"
	checkRunStatusCompleted   checkRunStatus     = "completed"
	checkRunConclusionSuccess checkRunConclusion = "success"
	checkRunConclusionSkipped checkRunConclusion = "skipped"
	checkRunConclusionFailure checkRunConclusion = "failure"
)

// BUG: This will trigger a failure even if the issue has been fixed

func (handler *PayloadHandler) HandleCheckSuite(checkSuitePayload *github.CheckSuiteEvent) {
	fmt.Println("Handling Check Suite request...")

	// Create a new Check Run
	fmt.Println("Creating new check run")
	inProgressString := string(checkRunStatusInProgress)
	checkRun, _, err := handler.GitHubApiClient.Checks.CreateCheckRun(
		context.Background(),
		*checkSuitePayload.Repo.Owner.Login,
		*checkSuitePayload.Repo.Name,
		github.CreateCheckRunOptions{
			Name:    "Orca Checks",
			HeadSHA: *checkSuitePayload.CheckSuite.HeadSHA,
			Status:  &inProgressString,
		})
	if err != nil {
		log.Fatal(err)
		return
	}

	// Bring over some of the properties we want to access later
	checkRun.CheckSuite.Repository = checkSuitePayload.Repo

	// Execute the check
	// TODO: Checks currently only supported on Pull Requests, need to find a better way to deal with this
	if len(checkSuitePayload.CheckSuite.PullRequests) > 0 {
		for _, pullRequest := range checkSuitePayload.CheckSuite.PullRequests {
			commits, _, err := handler.GitHubApiClient.PullRequests.ListCommits(
				context.Background(),
				*checkSuitePayload.Repo.Owner.Login,
				*checkSuitePayload.Repo.Name,
				*pullRequest.Number,
				nil)
			if err != nil {
				handler.handleFailure(checkRun, "Failed to list commits from Pull Request", err)
				return
			}

			commitScanResults, err := handler.Scanner.CheckCommits(checkSuitePayload.Repo, handler.GitHubApiClient, commits)
			if err != nil {
				handler.handleFailure(checkRun, "Failed to scan commits from Pull Request", err)
				return
			}

			if len(commitScanResults) > 0 {
				log.Printf("Potentially sensitive information detected in pull request #%d. Failing check.\n", pullRequest.Number)
				title, text := BuildMessage(commitScanResults)
				handler.completeCheckRun(checkRun, checkRunConclusionFailure, title, &text)
				return
			} else {
				log.Printf("No matches to address in pull request #%d.\n", pullRequest.Number)
			}
		}

		// Made it here, all is well
		handler.completeCheckRun(checkRun, checkRunConclusionSuccess, "No issues detected", nil)
	} else {
		handler.completeCheckRun(
			checkRun,
			checkRunConclusionSkipped,
			"No Pull Requests found. Orca Checks are currently only supported from Pull Requests",
			nil)
		log.Println("No pull requests. Skipping.")
		return
	}
}

func (handler *PayloadHandler) handleFailure(checkRun *github.CheckRun, summary string, err error) {
	handler.updateCheckRun(
		checkRun,
		checkRunStatusCompleted,
		checkRunConclusionFailure,
		summary,
		nil)
	log.Fatal(err)
}

func (handler *PayloadHandler) completeCheckRun(checkRun *github.CheckRun, conclusion checkRunConclusion, summary string, text *string) {
	handler.updateCheckRun(
		checkRun,
		checkRunStatusCompleted,
		conclusion,
		summary,
		text)
}

func (handler *PayloadHandler) updateCheckRun(
	checkRun *github.CheckRun,
	status checkRunStatus,
	conclusion checkRunConclusion,
	summary string,
	text *string) {

	statusString := string(status)
	conclusionString := string(conclusion)
	outputTitle := "Orca Checks"

	_, _, err := handler.GitHubApiClient.Checks.UpdateCheckRun(
		context.Background(),
		*checkRun.CheckSuite.Repository.Owner.Login,
		*checkRun.CheckSuite.Repository.Name,
		*checkRun.ID,
		github.UpdateCheckRunOptions{
			Status:     &statusString,
			Conclusion: &conclusionString,
			Output: &github.CheckRunOutput{
				Title:   &outputTitle,
				Summary: &summary,
				Text:    text,
			},
		})

	if err != nil {
		// TODO: At this point we're going to have a dangling check,
		// 	need to persist these checks somewhere so we can clean them up after a failure
		log.Fatalf("Could not mark check run as failed: %v", err)
	}
}
