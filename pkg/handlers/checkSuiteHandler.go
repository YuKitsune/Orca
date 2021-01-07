package handlers

import (
	"Orca/pkg/scanning"
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

// BUG: This will trigger a failure even if the issue has been fixed in a more recent commit

func (handler *PayloadHandler) HandleCheckSuite(checkSuitePayload *github.CheckSuiteEvent) {
	fmt.Println("Handling Check Suite request...")

	// Create a new Check Run
	fmt.Println("Creating new check run")
	inProgressString := string(checkRunStatusInProgress)
	checkRun, _, err := handler.GitHubClient.Checks.CreateCheckRun(
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
	if len(checkSuitePayload.CheckSuite.PullRequests) > 0 {
		for _, pullRequest := range checkSuitePayload.CheckSuite.PullRequests {
			commits, _, err := handler.GitHubClient.PullRequests.ListCommits(
				context.Background(),
				*checkSuitePayload.Repo.Owner.Login,
				*checkSuitePayload.Repo.Name,
				*pullRequest.Number,
				nil)
			if err != nil {
				handler.handleFailure(checkRun, "Failed to list commits from Pull Request", err)
				return
			}

			// Note: Timestamp not available in these commits for some reason (but they are in the Push event???)
			//	Have to assume the commits are in the correct order.

			// Get a list of commit SHAs
			var commitSHAs []string
			for _, commit := range commits {
				commitSHAs = append(commitSHAs, *commit.SHA)
			}

			commitScanResults, err := handler.Scanner.CheckCommits(
				checkSuitePayload.Repo.Owner.Login,
				checkSuitePayload.Repo.Name,
				handler.GitHubClient,
				commitSHAs)
			if err != nil {
				handler.handleFailure(checkRun, "Failed to scan commits from Pull Request", err)
				return
			}

			if len(commitScanResults) > 0 {

				// Todo: Once scan results are persisted, only act on new scan results

				// If all matches are resolved, pass the check, but reply with a reminder that the matches can still be
				//	viewed in the commit history
				var conclusion checkRunConclusion
				if AllMatchesAreResolved(commitScanResults) {
					log.Printf("Matches found but resolved in pull request #%d. Passing check with reminder.\n", pullRequest.Number)
					conclusion = checkRunConclusionSuccess

					// Reply with reminder
					body := "## :warning: Heads up!\n"
					body += "It looks like there is _potentially_ sensitive information in the commit history, but it appears to have since been removed.\n"
					body += fmt.Sprintf("See the [Orca check results](%s) for more information.\n", *checkRun.HTMLURL)
					body += "If any sensitive information is in the history, please make sure it is addressed appropriately." // Todo: Reword this line
					_, _, err := handler.GitHubClient.Issues.CreateComment(
						context.Background(),
						*checkSuitePayload. Repo.Owner.Login,
						*checkSuitePayload.Repo.Name,
						*pullRequest.Number,
						&github.IssueComment {
							Body: &body,
						})
					if err != nil {
						handler.handleFailure(checkRun, "Failed to reply to Pull Request with commit history warning", err)
						return
					}
				} else {
					log.Printf("Potentially sensitive information detected in pull request #%d. Failing check.\n", pullRequest.Number)
					conclusion = checkRunConclusionFailure
				}

				title, text := BuildMessage(commitScanResults)
				handler.completeCheckRun(checkRun, conclusion, title, &text)

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

	_, _, err := handler.GitHubClient.Checks.UpdateCheckRun(
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
		// TODO: At this point we're going to have an abandoned check,
		// 	need to persist these checks somewhere so we can clean them up after a failure
		log.Fatalf("Could not mark check run as failed: %v", err)
	}
}

func AllMatchesAreResolved(scanResults []scanning.CommitScanResult) bool {
	for _, result := range scanResults {
		for _, match := range result.Matches {
			if !match.Resolved {
				return false
			}
		}
	}

	return true
}