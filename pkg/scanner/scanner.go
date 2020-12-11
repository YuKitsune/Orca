package scanner

import (
	"Orca/pkg/patterns"
	"strings"
)

type FileMatch struct {
	FileName string
	Kind string
	Error error
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

func FindDangerousFilesForPatterns(filePaths []string, patterns []patterns.SearchPattern) []FileMatch {

	var matches []FileMatch
	for i := 0; i < len(filePaths); i++ {
		var currentFilePath = filePaths[i]
		for j := 0; j < len(patterns); j++ {
			var currentPattern = patterns[j]
			if currentPattern.Regex.MatchString(currentFilePath) {
				matches = append(matches, FileMatch{
					FileName: currentFilePath,
					Kind:     currentPattern.Kind,
				})
			}
		}
	}

	return matches
}

func ScanContentForPatterns(content string, patterns []patterns.SearchPattern) ContentMatch {

	var result ContentMatch

	var lines = strings.Split(content, "\n")
	for i := 0; i < len(lines); i++ {
		var matchesOnLine = scanLineForPatterns(lines[i], patterns)
		var lineMatches = LineMatch{
			LineNumber: i+1,
			Matches:    matchesOnLine,
		}

		result.LineMatches = append(result.LineMatches, lineMatches)
	}

	return result
}

func scanLineForPatterns(line string, patterns []patterns.SearchPattern) []Match {
	var matches []Match
	for i := 0; i < len(patterns); i++ {
		var currentPatternMatches = scanLineForPattern(line, patterns[i])
		if len(currentPatternMatches) > 0 {
			matches = append(matches, currentPatternMatches...)
		}
	}

	return matches
}

func scanLineForPattern(line string, pattern patterns.SearchPattern) []Match {
	var matches []Match
	var regexMatches = pattern.Regex.FindAllStringIndex(line, -1)
	for i := 0; i < len(regexMatches); i++ {
		var startIndex = regexMatches[i][0]
		var endIndex = regexMatches[i][1]
		matches = append(matches, Match{
			StartIndex: startIndex,
			EndIndex:   endIndex,
			value:      line[startIndex:endIndex],
			Kind:       pattern.Kind,
		})
	}

	return matches
}