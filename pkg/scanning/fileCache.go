package scanning

import (
	"context"
	"encoding/base64"
	"github.com/google/go-github/v33/github"
)

var (
	cache *fileCache
)

type FileState string

const (
	FileAdded    FileState = "added"
	FileModified FileState = "modified"
	FileRemoved  FileState = "removed"
)

type File struct {
	CommitSHA    string
	Path         string
	Content      string
	PermalinkURL string
	Status       FileState
}

type fileCache struct {
	files []File
}

func (cache *fileCache) addFile(file File) {

	// Remove any conflicting file
	index, existingFile := cache.getFileFromCommit(file.CommitSHA, file.Path)
	if existingFile != nil {
		cache.files = append(cache.files[:index], cache.files[index+1:]...)
	}

	// Add the file
	cache.files = append(cache.files, file)
}

func (cache *fileCache) getFilesFromCommit(commitSHA string) []*File {
	var results []*File
	for _, file := range cache.files {
		if file.CommitSHA == commitSHA {
			results = append(results, &file)
		}
	}

	return results
}

func (cache *fileCache) getFileFromCommit(commitSHA string, fileName string) (int, *File) {
	for i, file := range cache.files {
		if file.CommitSHA == commitSHA && file.Path == fileName {
			return i, &file
		}
	}

	return -1, nil
}

func getFileCache() *fileCache {
	if cache == nil {
		cache = &fileCache{
			files: []File{},
		}
	}

	return cache
}

// TODO: Split the file into cache and scanner?

func GetFile(query GitHubFileQuery, client *github.Client) (*File, error) {

	// Check the cache
	cache := getFileCache()
	_, file := cache.getFileFromCommit(query.CommitSHA, query.FileName)

	// If not in the cache, then send a request and cache the result for later
	if file == nil {

		file = &File{
			CommitSHA: query.CommitSHA,
			Path:      query.FileName,
			Status:    query.Status,
		}

		// If the file was not removed, then we can go ahead and get it's content and permalink
		if query.Status != FileRemoved {
			content, _, _, err := client.Repositories.GetContents(
				context.Background(),
				query.RepoOwner,
				query.RepoName,
				query.FileName,
				&github.RepositoryContentGetOptions{
					Ref: query.CommitSHA,
				})
			if err != nil {
				return nil, err
			}

			contentBytes, err := base64.StdEncoding.DecodeString(*content.Content)
			if err != nil {
				return nil, err
			}
			contentString := string(contentBytes)

			file.Content = contentString
			file.PermalinkURL = *content.HTMLURL
		}

		cache.addFile(*file)
	}

	return file, nil
}