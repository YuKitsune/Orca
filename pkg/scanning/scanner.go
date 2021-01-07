package scanning

import (
	"context"
	"encoding/base64"
	"github.com/google/go-github/v33/github"
	"log"
	"strings"
)

type FileContentMatch struct {
	File
	LineMatch
}

type File struct {
	Path         *string
	Content      *string
	HTMLURL      *string
	PermalinkURL *string
}

type LineMatch struct {
	LineNumber int
	Match
}

type Match struct {
	StartIndex int
	EndIndex   int
	value      string
	Kind       string
	Resolved   bool
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
			if *file.Status == "removed" {
				for i, previousScanResult := range commitScanResults {
					for j, previousFileMatch := range previousScanResult.Matches {
						if *previousFileMatch.Path == *file.Filename {
							commitScanResults[i].Matches[j].Resolved = true
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

			fileContentMatches, err := scanner.CheckFileContentFromCommit(
				githubClient,
				repo.Owner.Login,
				repo.Name,
				commit.SHA,
				file.Filename)
			if err != nil {
				return nil, err
			}

			if len(fileContentMatches) > 0 {

				// Ignore previously known matches
				for _, fileContentMatch := range fileContentMatches {
					if !MatchIsKnown(getMatches(commitScanResults), fileContentMatch) {
						commitScanResult.Matches = append(commitScanResult.Matches, fileContentMatch)
					}
				}

			} else {

				// No matches found, previous matches in this file should be resolved
				for i, previousScanResult := range commitScanResults {
					for j, previousFileMatch := range previousScanResult.Matches {
						if *previousFileMatch.Path == *file.Filename {
							commitScanResults[i].Matches[j].Resolved = true
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

func (scanner *Scanner) CheckFileContentFromCommit(
	githubClient *github.Client,
	repoOwner *string,
	repoName *string,
	commit *string,
	filePath *string) ([]FileContentMatch, error) {

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

func (scanner *Scanner) CheckFileContent(file File) ([]FileContentMatch, error) {

	var result []FileContentMatch

	lineMatches, err := scanner.CheckContent(file.Content)
	if err != nil {
		return nil, err
	}

	if len(lineMatches) > 0 {
		for _, lineMatch := range lineMatches {
			fileMatch := FileContentMatch{
				File:      file,
				LineMatch: lineMatch,
			}

			result = append(result, fileMatch)
		}
	}

	return result, nil
}

func (scanner *Scanner) CheckContent(content *string) ([]LineMatch, error) {

	var result []LineMatch

	// Todo: Multi-line scan first, then single-line scan around any multi-line match ranges
	var lines = strings.Split(*content, "\n")
	for i, line := range lines {
		lineNumber := i + 1
		matchesOnLine, err := scanner.scanLineForPatterns(line)
		if err != nil {
			return nil, err
		}

		// Todo: Another loop, any way around this?
		if len(matchesOnLine) > 0 {
			for _, matchOnLine := range matchesOnLine {
				lineMatch := LineMatch{
					LineNumber: lineNumber,
					Match:      matchOnLine,
				}

				result = append(result, lineMatch)
			}
		}
	}

	return result, nil
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

func getMatches(commitScanResults []CommitScanResult) []FileContentMatch {
	var result []FileContentMatch
	for _, commitScanResult := range commitScanResults {
		result = append(result, commitScanResult.Matches...)
	}

	return result
}

func MatchIsKnown(knownFileContentMatches []FileContentMatch, newFileContentMatch FileContentMatch) bool {
	for _, knownFileContentMatch := range knownFileContentMatches {
		if *knownFileContentMatch.Path == *newFileContentMatch.Path {
			if newFileContentMatch.value == knownFileContentMatch.value &&
				!knownFileContentMatch.Resolved {
				return true
			}
		}
	}

	return false
}
