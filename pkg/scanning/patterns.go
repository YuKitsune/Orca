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

	var secretPattern, err = MakePattern("(secret)", "Secret")
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, *secretPattern)

	return patterns, nil
}

func GetFilePatterns() ([]SearchPattern, error) {
	// Todo: Fetch from API?
	var patterns []SearchPattern

	var certificatePattern, err = MakePattern("(\\.crt)|(\\.cer)|(\\.ca-bundle)|(\\.p7b)|(\\.p7c)|(\\.p7s)|(\\.pem)|(\\.key)|(\\.keystore)|(\\.jks)|(\\.p12)|(\\.pfx)|(\\.pem)", "Certificate")
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, *certificatePattern)

	return patterns, nil
}