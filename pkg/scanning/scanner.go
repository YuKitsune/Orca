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
	FileNamePatterns []SearchPattern
	ContentPatterns  []SearchPattern
}

func NewScanner() (*Scanner, error) {
	fileNamePatterns, err := GetFileNamePatterns()
	if err != nil {
		return nil, err
	}

	contentPatterns, err := GetContentPatterns()
	if err != nil {
		return nil, err
	}

	scanner := &Scanner {
		FileNamePatterns: fileNamePatterns,
		ContentPatterns:  contentPatterns,
	}

	return scanner, nil
}

func (scanner *Scanner) CheckFileName(file File) []FileMatch {

	var matches []FileMatch
	for _, pattern := range scanner.FileNamePatterns {
		if pattern.Regex.MatchString(*file.Path) {
			matches = append(matches, FileMatch{
				File: file,
				Kind: pattern.Kind,
			})
		}
	}

	return matches
}

func (scanner *Scanner) CheckFileContent(file File) FileContentMatch {

	result := FileContentMatch {
		File: file,
	}

	contentResult := scanner.checkContent(*file.Content)
	if contentResult.HasMatches() {
		result.LineMatches = contentResult.LineMatches
	}

	return result
}

func (scanner *Scanner) checkContent(content string) ContentMatch {

	var result ContentMatch

	var lines = strings.Split(content, "\n")
	for i, line := range lines {
		var matchesOnLine = scanner.scanLineForPatterns(line)
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

func (scanner *Scanner) scanLineForPatterns(line string) []Match {
	var matches []Match
	for _, pattern := range scanner.ContentPatterns {
		var currentPatternMatches = scanLineForPattern(line, pattern)
		if len(currentPatternMatches) > 0 {
			matches = append(matches, currentPatternMatches...)
		}
	}

	return matches
}

func scanLineForPattern(line string, pattern SearchPattern) []Match {
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
