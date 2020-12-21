package handlers

import (
	"Orca/pkg/scanning"
	"context"
	"fmt"
	"github.com/google/go-github/v33/github"
	"log"
	"strings"
)

type MatchHandler struct {
	GitHubApiClient *github.Client
}

func NewMatchHandler(gitHubApiClient *github.Client) *MatchHandler {
	return &MatchHandler{
		GitHubApiClient: gitHubApiClient,
	}
}

func (matchHandler *MatchHandler) HandleMatchesFromPush(pushPayload *github.PushEvent, results []scanning.CommitScanResult) error {

	// Open a new issue
	var title string
	if len(results) > 1 {
		title = fmt.Sprintf("Potentially sensitive data found in %d commits", len(results))
	} else {
		title = "Potentially sensitive data found in a commit"
	}

	log.Printf("Opening a new issue \"%s\"\n", title)

	body := "Potentially sensitive data has recently been pushed to this repository.\n\n"

	for _, result := range results {
		body += fmt.Sprintf("Introduced in %s:\n", result.Commit)

		// Add content matches
		if len(result.Matches) > 0 {

			body += "Files containing potentially sensitive data:\n"
			for _, contentMatch := range result.Matches {

				// Todo: Group lines which are directly below each other into one permalink (e.g. #L2-L4)
				body += fmt.Sprintf("### %s\n", *contentMatch.Path)
				for _, lineMatch := range contentMatch.LineMatches {

					// TODO: Add a buffer around the line for extra context
					var matchKinds []string
					for _, match := range lineMatch.Matches {
						matchKinds = append(matchKinds, match.Kind)
					}

					body += fmt.Sprintf("#### %s:", strings.Join(matchKinds, ", "))
					body += fmt.Sprintf("%s#L%d\n", *contentMatch.PermalinkURL, lineMatch.LineNumber)
				}
			}
		}
	}

	issue, _, err := matchHandler.GitHubApiClient.Issues.Create(
		context.Background(),
		*pushPayload.Repo.Owner.Login,
		*pushPayload.Repo.Name,
		&github.IssueRequest{
			Title:     &title,
			Body:      &body,
			Assignee:  pushPayload.Pusher.Name,
		})
	if err != nil {
		return err
	}

	log.Printf("Issue #%d opened\n", issue.Number)

	return nil
}

func (matchHandler *MatchHandler) HandleMatchesFromIssue(issue *github.IssuesEvent, result *scanning.IssueScanResult) error {

	log.Printf("Redacting matches from #%d\n", issue.Issue.Number)
	newBody := redactMatchesFromContent(*issue.Issue.Body, result, '*')

	// Replace the issue body with the new body with redacted matches
	_, _, err := matchHandler.GitHubApiClient.Issues.Edit(
		context.Background(),
		*issue.Issue.Repository.Owner.Login,
		*issue.Issue.Repository.Name,
		*issue.Issue.Number,
		&github.IssueRequest {
			Body: &newBody,
		})
	if err != nil {
		return err
	}
	log.Printf("Matches from #%d redacted\n", issue.Issue.Number)

	return nil
}

func (matchHandler *MatchHandler) HandleMatchesFromIssueComment(issue *github.IssueCommentEvent, result *scanning.IssueScanResult) error {

	log.Printf("Redacting matches from #%d (comment %d)\n", issue.Issue.Number, issue.Comment.ID)
	newBody := redactMatchesFromContent(*issue.Comment.Body, result, '*')

	// Replace the issue body with the new body with redacted matches
	_, _, err := matchHandler.GitHubApiClient.Issues.EditComment(
		context.Background(),
		*issue.Repo.Owner.Login,
		*issue.Repo.Name,
		*issue.Comment.ID,
		&github.IssueComment{
			Body: &newBody,
		})
	if err != nil {
		return err
	}
	log.Printf("Matches from #%d (comment %d) redacted\n", issue.Issue.Number, issue.Comment.ID)

	return nil
}

func redactMatchesFromContent(content string, lineMatches []scanning.LineMatch, replacementCharacter rune) string {

	contentRunes := []rune(content)
	for _, lineMatch := range lineMatches {
		lineNumber := lineMatch.LineNumber

		// Skip to the specified line number
		currentLineNumber := 1
		indexInLine := 0
		for i, ch := range contentRunes {

			// Counting each new line we find to find the line number
			startOfNewLine := false
			if ch == '\n' {
				currentLineNumber++
				indexInLine = 0
				startOfNewLine = true
			}

			// Make sure we're on the right line
			if currentLineNumber == lineNumber {

				// Check if the current index on the line is within the range of one of the matches
				for _, match := range lineMatch.Matches {
					if indexInLine >= match.StartIndex && indexInLine < match.EndIndex {
						contentRunes[i] = replacementCharacter
					}
				}
			}

			// Don't increment if we just decided we're on a new line
			if !startOfNewLine {
				indexInLine++
			}
		}
	}

	contentString := string(contentRunes)

	return contentString
}
