package scanning

import (
	"context"
	"encoding/base64"
	"github.com/google/go-github/v33/github"
	"log"
)

var (
	cache *fileCache
)

const (
	FileAdded    FileState = "added"
	FileModified FileState = "modified"
	FileRemoved  FileState = "removed"
)

type FileState string

type File struct {
	CommitSHA    string
	Path         string
	Content      string
	PermalinkURL string
	Status       FileState
}

// Todo: If this is going to run as a serverless application, then it will make more sense to use Redis or Memcached
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

func getFileCache() *fileCache {
	if cache == nil {
		cache = &fileCache{
			files: []File{},
		}
	}

	return cache
}

func (cache *fileCache) getFileFromCommit(commitSHA string, fileName string) (int, *File) {
	for i, file := range cache.files {
		if file.CommitSHA == commitSHA && file.Path == fileName {
			return i, &file
		}
	}

	return -1, nil
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
			log.Printf("%s from %s not available in cache, fetching from API\n", query.FileName, query.CommitSHA)
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
	} else {
		log.Printf("%s from %s fetched from cache\n", query.FileName, query.CommitSHA)
	}

	return file, nil
}
