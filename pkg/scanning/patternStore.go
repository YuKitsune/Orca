package scanning

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type PatternStore interface {
	GetContentPatterns() []SearchPattern
	GetFileNamePatterns() []SearchPattern
}

type JsonPatternStore struct {
	ContentPatternsJsonFile string
	FileNamePatternsJsonFile string
}

func (store *JsonPatternStore) GetContentPatterns() ([]SearchPattern, error) {
	return getPatternsFromFile(store.ContentPatternsJsonFile)
}

func (store *JsonPatternStore) GetFileNamePatterns() ([]SearchPattern, error) {
	return getPatternsFromFile(store.FileNamePatternsJsonFile)
}

func getPatternsFromFile(file string) ([]SearchPattern, error) {

	jsonFile, err := os.Open(file)
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
