package rectifier

import (
	"Orca/pkg/payloads"
	"Orca/pkg/scanning"
	"context"
	"fmt"
	gitHubApi "github.com/google/go-github/v33/github"
	"log"
)

type Rectifier struct {
	GitHubApiClient *gitHubApi.Client
}

func NewRectifier(gitHubApiClient *gitHubApi.Client) *Rectifier {
	return &Rectifier {
		GitHubApiClient: gitHubApiClient,
	}
}

func (rectifier *Rectifier) RectifyFromPush(pushPayload payloads.PushPayload, results []scanning.CommitScanResult) error {

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

		// Add dangerous files
		if len(result.FileMatches) > 0 {

			body += "Potentially sensitive files:\n"
			for _, dangerousFile := range result.FileMatches {
				body += fmt.Sprintf("- [%s](%s)\n", *dangerousFile.Path, *dangerousFile.HTMLURL)
			}

			body += "\n\n"
		}

		// Add content matches
		if len(result.ContentMatches) > 0 {

			body += "Files containing potentially sensitive data:\n"
			for _, contentMatch := range result.ContentMatches {

				// Todo: Group lines which are directly below each other into one permalink (e.g. #L2-L4)
				body += fmt.Sprintf("### %s\n", *contentMatch.Path)
				for _, lineMatch := range contentMatch.LineMatches {

					// TODO: Add a buffer around the line for extra context
					body += fmt.Sprintf("%s#L%d\n", *contentMatch.PermalinkURL, lineMatch.LineNumber)
				}
			}
		}
	}

	issue, _, err := rectifier.GitHubApiClient.Issues.Create(
		context.Background(),
		pushPayload.Repository.Owner.Login,
		pushPayload.Repository.Name,
		&gitHubApi.IssueRequest{
			Title:     &title,
			Body:      &body,
			Assignee:  &pushPayload.Author.Name,
		})
	if err != nil {
		return err
	}

	log.Printf("Issue #%d opened\n", issue.Number)

	return nil
}

func (rectifier *Rectifier) RemediateFromIssue(issue payloads.IssuePayload, result scanning.IssueScanResult) error {

	log.Printf("Redacting matches from #%d\n", issue.Number)
	newBody := redactMatchesFromContent(issue.Body, result, '*')

	// Replace the issue body with the new body with redacted matches
	_, _, err := rectifier.GitHubApiClient.Issues.Edit(
		context.Background(),
		issue.Repository.Owner.Login,
		issue.Repository.Name,
		issue.Number,
		&gitHubApi.IssueRequest {
			Body: &newBody,
		})
	if err != nil {
		return err
	}
	log.Printf("Matches from #%d redacted\n", issue.Number)

	return nil
}

func (rectifier *Rectifier) RemediateFromIssueComment(issue payloads.IssueCommentPayload, result scanning.IssueScanResult) error {

	log.Printf("Redacting matches from #%d (comment %d)\n", issue.Number, issue.CommentId)
	newBody := redactMatchesFromContent(issue.Body, result, '*')

	// Replace the issue body with the new body with redacted matches
	_, _, err := rectifier.GitHubApiClient.Issues.EditComment(
		context.Background(),
		issue.Repository.Owner.Login,
		issue.Repository.Name,
		issue.CommentId,
		&gitHubApi.IssueComment{
			Body: &newBody,
		})
	if err != nil {
		return err
	}
	log.Printf("Matches from #%d (comment %d) redacted\n", issue.Number, issue.CommentId)

	return nil
}

func redactMatchesFromContent(content string, result scanning.IssueScanResult, replacementCharacter rune) string {

	contentRunes := []rune(content)
	for _, lineMatch := range result.LineMatches {
		lineNumber := lineMatch.LineNumber

		// Skip to the specified line number
		currentLineNumber := 1
		indexInLine := 0
		for i, ch := range contentRunes {

			// Counting each new line we find to find the line number
			if ch == '\n' {
				currentLineNumber++
			}

			// Don't care about the rest of this unless we're on the right line number
			if currentLineNumber != lineNumber {
				continue
			}

			// At this point we're on the right line
			// Check if the current index on the line is within the range of one of the matches
			for _, match := range lineMatch.Matches {
				if i >= match.StartIndex && i < match.EndIndex {
					contentRunes[i] = replacementCharacter
				}
			}

			indexInLine++
		}
	}

	contentString := string(contentRunes)

	return contentString
}
