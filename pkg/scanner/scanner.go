package scanner

import (
	"Orca/pkg/patterns"
	"strings"
)

type CommitScanResult struct {
	Commit string
	FileMatches []FileMatch
	ContentMatches []ContentMatch
}

func (result *CommitScanResult) HasMatches() bool {
	return len(result.FileMatches) > 0 || len(result.ContentMatches) > 0
}
type File struct {
	Path    *string
	Content *string
	HTMLURL *string
	PermalinkURL *string
}

type FileMatch struct {
	File
	Kind string
	Error error
}

type ContentMatch struct {
	File
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

func (matches *ContentMatch) HasMatches() bool {
	return len(matches.LineMatches) > 0
}

func FindDangerousPatternsFromFile(file File, patterns []patterns.SearchPattern) []FileMatch {

	var matches []FileMatch
	for _, pattern := range patterns {
		if pattern.Regex.MatchString(*file.Path) {
			matches = append(matches, FileMatch{
				File: file,
				Kind: pattern.Kind,
			})
		}
	}

	return matches
}

func ScanContentForPatterns(file File, patterns []patterns.SearchPattern) ContentMatch {

	result := ContentMatch {
		File: file,
	}

	var lines = strings.Split(*file.Content, "\n")
	for i, line := range lines {
		var matchesOnLine = scanLineForPatterns(line, patterns)
		if len(matchesOnLine) > 0 {
			var lineMatches = LineMatch{
				LineNumber: i+1,
				Matches:    matchesOnLine,
			}

			result.LineMatches = append(result.LineMatches, lineMatches)
		}
	}

	return result
}

func scanLineForPatterns(line string, patterns []patterns.SearchPattern) []Match {
	var matches []Match
	for _, pattern := range patterns {
		var currentPatternMatches = scanLineForPattern(line, pattern)
		if len(currentPatternMatches) > 0 {
			matches = append(matches, currentPatternMatches...)
		}
	}

	return matches
}

func scanLineForPattern(line string, pattern patterns.SearchPattern) []Match {
	var matches []Match
	var regexMatches = pattern.Regex.FindAllStringIndex(line, -1)
	for _, match := range regexMatches {
		var startIndex = match[0]
		var endIndex = match[1]
		matches = append(matches, Match{
			StartIndex: startIndex,
			EndIndex:   endIndex,
			value:      line[startIndex:endIndex],
			Kind:       pattern.Kind,
		})
	}

	return matches
}