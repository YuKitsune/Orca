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
