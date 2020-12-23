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
