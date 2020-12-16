package scanning

import "regexp"

type SearchPattern struct {
	Regex regexp.Regexp
	Kind  string
}

func MakePattern(regexPattern string, kind string) (*SearchPattern, error) {
	var regex, err = regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}

	var pattern = &SearchPattern{
		Regex: *regex,
		Kind: kind,
	}

	return pattern, nil
}

func GetContentPatterns() ([]SearchPattern, error) {
	// Todo: Fetch from API?
	var patterns []SearchPattern

	var gitHubPat, err = MakePattern("[a-z0-9]{40}", "GitHub Personal Access Token")
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, *gitHubPat)

	return patterns, nil
}

func GetFileNamePatterns() ([]SearchPattern, error) {
	// Todo: Fetch from API?
	var patterns []SearchPattern

	var certificatePattern, err = MakePattern("(\\.crt)|(\\.cer)|(\\.ca-bundle)|(\\.p7b)|(\\.p7c)|(\\.p7s)|(\\.pem)|(\\.key)|(\\.keystore)|(\\.jks)|(\\.p12)|(\\.pfx)|(\\.pem)", "Certificate")
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, *certificatePattern)

	return patterns, nil
}
