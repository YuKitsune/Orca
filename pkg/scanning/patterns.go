package scanning

import "regexp"

type SearchPattern struct {
	Pattern string
	Kind  string
	Exclusions []string
}

func (pattern *SearchPattern) GetRegexp() (*regexp.Regexp, error) {
	return regexp.Compile(pattern.Pattern)
}

func (pattern *SearchPattern) CanIgnore(value string) bool {
	for _, exclusionPatternString := range pattern.Exclusions {
		exclusionPattern := regexp.MustCompile(exclusionPatternString)
		return exclusionPattern.MatchString(value)
	}
	return false
}
