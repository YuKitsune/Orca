package scanning

import (
	"strings"
)

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

func (matches *ContentMatch) HasMatches() bool {
	return len(matches.LineMatches) > 0
}

type Scanner struct {
	Patterns []SearchPattern
}

func NewScanner(patternStore *PatternStore) (*Scanner, error) {

	patterns, err := (*patternStore).GetPatterns()
	if err != nil {
		return nil, err
	}

	scanner := &Scanner {
		Patterns: patterns,
	}

	return scanner, nil
}

func (scanner *Scanner) CheckFileContent(file File) (*FileContentMatch, error) {

	result := FileContentMatch {
		File: file,
	}

	contentResult, err := scanner.checkContent(file.Content)
	if err != nil {
		return nil, err
	}

	if contentResult.HasMatches() {
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
				LineNumber: i+1,
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
		if Contains(pattern.Exclusions, value) {
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

func Contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}

	return false
}