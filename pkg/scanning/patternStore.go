package scanning

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type PatternStore interface {
	GetPatterns() ([]SearchPattern, error)
}

type FilePatternStore struct {
	PatternsJsonFile string
}

func (store *FilePatternStore) GetPatterns() ([]SearchPattern, error) {

	jsonFile, err := os.Open(store.PatternsJsonFile)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var result []SearchPattern
	if err := json.Unmarshal(byteValue, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func NewPatternStore(patternsLocation string) (PatternStore, error) {
	if strings.HasPrefix(patternsLocation, "http") {
		// Todo
		return nil, errors.New("fetching patterns from a URL is not yet implemented")
	} else if fileExists(patternsLocation) {
		store := &FilePatternStore { PatternsJsonFile: patternsLocation }
		return store, nil
	} else {
		errorMessage := fmt.Sprintf("unsupported patterns location \"%s\"\n", patternsLocation)
		return nil, errors.New(errorMessage)
	}
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}