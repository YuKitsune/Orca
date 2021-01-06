package scanning

import (
	"context"
	"encoding/base64"
	"github.com/google/go-github/v33/github"
	"log"
	"strings"
)

type File struct {
	Path         *string
	Content      *string
	HTMLURL      *string
	PermalinkURL *string
}

type FileMatch struct {
	File
	Kind  string
	Error error
}

type FileContentMatch struct {
	File
	ContentMatch
}

type ContentMatch struct {
	LineMatches []LineMatch
}

type LineMatch struct {
	LineNumber int
	Matches    []Match
}

type Match struct {
	StartIndex int
	EndIndex   int
	value      string
	Kind       string
	Resolved bool
}

type Scanner struct {
	Patterns []SearchPattern
}

func NewScanner(patternStore *PatternStore) (*Scanner, error) {

	patterns, err := (*patternStore).GetPatterns()
	if err != nil {
		return nil, err
	}

	scanner := &Scanner{
		Patterns: patterns,
	}

	return scanner, nil
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

			// If the file was removed, then mark any previous matches as resolved
			// Todo: Fix the nesting!!!
			if *file.Status == "removed" {
				for _, previousScanResult := range commitScanResults {
					for _, previousFileMatch := range previousScanResult.Matches {
						if previousFileMatch.Path == file.Filename {
							for _, previousLineMatch := range previousFileMatch.LineMatches {
								for _, previousMatch := range previousLineMatch.Matches {
									previousMatch.Resolved = true
								}
							}
						}
					}
				}

				continue
			}

			// Can only scan contents of added and modified files
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

			// Remove any already known matches
			fileContentMatch = RemoveKnownMatches(GetMatches(commitScanResults), *fileContentMatch)
			if len(fileContentMatch.LineMatches) > 0 {
				commitScanResult.Matches = append(commitScanResult.Matches, *fileContentMatch)
			} else {
				// No matches, all previous matches are also probably resolved
				for _, previousScanResult := range commitScanResults {
					for _, previousFileMatch := range previousScanResult.Matches {
						if previousFileMatch.Path == file.Filename {
							for _, previousLineMatch := range previousFileMatch.LineMatches {
								for _, previousMatch := range previousLineMatch.Matches {
									previousMatch.Resolved = true
								}
							}
						}
					}
				}
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

func (scanner *Scanner) CheckFileContentFromCommit(
	githubClient *github.Client,
	repoOwner *string,
	repoName *string,
	commit *string,
	filePath *string) (*FileContentMatch, error) {

	// Todo: Is there a bulk alternative to GetContents?
	// 	Don't want to request for each file, could have a big commit
	content, _, _, err := githubClient.Repositories.GetContents(
		context.Background(),
		*repoOwner,
		*repoName,
		*filePath,
		&github.RepositoryContentGetOptions{
			Ref: *commit,
		})
	if err != nil {
		return nil, err
	}

	contentBytes, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		return nil, err
	}
	contentString := string(contentBytes)

	permalinkUrl := *content.HTMLURL

	file := File{
		Path:         content.Path,
		Content:      &contentString,
		HTMLURL:      content.HTMLURL,
		PermalinkURL: &permalinkUrl,
	}

	return scanner.CheckFileContent(file)
}

func (scanner *Scanner) CheckFileContent(file File) (*FileContentMatch, error) {

	result := FileContentMatch{
		File: file,
	}

	contentResult, err := scanner.checkContent(file.Content)
	if err != nil {
		return nil, err
	}

	if len(contentResult.LineMatches) > 0 {
		result.LineMatches = contentResult.LineMatches
	}

	return &result, nil
}

func (scanner *Scanner) checkContent(content *string) (*ContentMatch, error) {

	var result ContentMatch

	// Todo: Multi-line scan first, then single-line scan around any multi-line match ranges
	var lines = strings.Split(*content, "\n")
	for i, line := range lines {
		matchesOnLine, err := scanner.scanLineForPatterns(line)
		if err != nil {
			return nil, err
		}

		if len(matchesOnLine) > 0 {
			var lineMatches = LineMatch{
				LineNumber: i + 1,
				Matches:    matchesOnLine,
			}

			result.LineMatches = append(result.LineMatches, lineMatches)
		}
	}

	return &result, nil
}

func (scanner *Scanner) scanLineForPatterns(line string) ([]Match, error) {
	var matches []Match
	for _, pattern := range scanner.Patterns {
		currentPatternMatches, err := scanLineForPattern(line, pattern)
		if err != nil {
			return nil, err
		}

		if len(currentPatternMatches) > 0 {
			matches = append(matches, currentPatternMatches...)
		}
	}

	return matches, nil
}

func scanLineForPattern(line string, pattern SearchPattern) ([]Match, error) {
	var matches []Match
	regex, err := pattern.GetRegexp()
	if err != nil {
		return nil, err
	}

	var regexMatches = regex.FindAllStringIndex(line, -1)
	for _, match := range regexMatches {
		var startIndex = match[0]
		var endIndex = match[1]
		value := line[startIndex:endIndex]

		// Ignore if the matched string is allowed to be excluded from checks
		if pattern.CanIgnore(value) {
			continue
		}

		matches = append(matches, Match{
			StartIndex: startIndex,
			EndIndex:   endIndex,
			value:      value,
			Kind:       pattern.Kind,
		})
	}

	return matches, nil
}

func GetMatches(commitScanResults []CommitScanResult) []FileContentMatch{
	var result []FileContentMatch
	for _, commitScanResult := range commitScanResults {
		result = append(result, commitScanResult.Matches...)
	}

	return result
}

func RemoveKnownMatches(knownFileContentMatches []FileContentMatch, newFileContentMatch FileContentMatch) *FileContentMatch {

	// Find the indexes of matches we already know about
	// Todo: Please for the love of god find a better way to do this without creating a skate park
	var matchesToRemove [][]int
	for _, knownFileContentMatch := range knownFileContentMatches {
		if knownFileContentMatch.Path == newFileContentMatch.Path {
			for _, knownLineMatch := range knownFileContentMatch.LineMatches {
				for i, newLineMatch := range newFileContentMatch.LineMatches {
					for _, knownMatch := range knownLineMatch.Matches {
						for j, newMatch := range newLineMatch.Matches {
							if newMatch.value == knownMatch.value {
								matchesToRemove = append(matchesToRemove, []int{i, j})
							}
						}
					}
				}
			}
		}
	}

	// Copy the new results and remove the matches so we can manually add the ones we care about
	// Todo: Is this assignment by value or reference?
	resultingMatch := newFileContentMatch
	resultingMatch.LineMatches = []LineMatch{}
	if len(matchesToRemove) > 0 {

		// Note: Because the index of each match will be moved back by 1 each time we remove something,
		//	the index of the item to remove needs to be moved back by 1 for each iteration
		//	We can use the index of the current loop to help us
		//  Example:
		//  First iteration,  0 changes,          index 5 stays as is
		//  Second iteration, 1 previous change,  index 6 moved to 5
		//  Third iteration,  2 previous changes, index 7 moved to 5

		for i, indexOfLineMatch := range matchesToRemove {

			lineMatch := newFileContentMatch.LineMatches[i]
			for j, indexOfMatch := range indexOfLineMatch {
				indexToRemove := indexOfMatch - j
				lineMatch.Matches = append(lineMatch.Matches[:indexToRemove], lineMatch.Matches[indexToRemove+1:]...)
			}

			if len(lineMatch.Matches) > 0 {
				resultingMatch.LineMatches = append(resultingMatch.LineMatches, lineMatch)
			}
		}
	}

	return &resultingMatch
}

// File removed: 							Keep match, mark as resolved
// Match no longer present in new commit:	Keep old match, mark as resolved
// Match is present in new commit:			Remove new match, old match has required data