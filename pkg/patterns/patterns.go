package patterns

import "regexp"

type SearchPattern struct {
	Regex *regexp.Regexp
	Kind  string
}

func MakePattern(regexPattern string, kind string) (SearchPattern, error) {
	var regex, err = regexp.Compile(regexPattern)
	if err != nil {
		return SearchPattern{}, err
	}

	var pattern = SearchPattern{
		Regex: regex,
		Kind: kind,
	}

	return pattern, nil
}

func MustMakePattern(regexPattern string, kind string) SearchPattern {
	var pattern, err = MakePattern(regexPattern, kind)
	if err != nil {
		return SearchPattern{}
	}

	return pattern
}
