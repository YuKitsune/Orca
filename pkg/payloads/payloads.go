package payloads

import (
	"gopkg.in/go-playground/webhooks.v5/github"
	"strings"
)

type Author struct {
	Name string
}

type Commit struct {
	ID       string
	Repository Repository
	Branch   string
	Author   Author
	Added    []string
	Modified []string
	Removed  []string
}

type RepositoryOwner struct {
	Login string
}

type Repository struct {
	Name  string
	Owner RepositoryOwner
}

type PushPayload struct {
	Ref        string
	Commits    []Commit
	Repository Repository
	Author Author
}

func (pushPayload *PushPayload) GetBranch() string {
	refSplit := strings.Split(pushPayload.Ref, "/")
	branchName := refSplit[len(refSplit) - 1]
	return branchName
}

type IssuePayload struct {
	Repository Repository
	Number int
	Author Author
	Body string
}

type IssueCommentPayload struct {
	IssuePayload
	CommentId int64
}

func NewPushPayload(payload github.PushPayload) PushPayload {

	pushPayload := PushPayload{
		Ref:     payload.Ref,
		Repository: Repository{
			Name: payload.Repository.Name,
			Owner: RepositoryOwner{
				Login: payload.Repository.Owner.Login,
			},
		},
		Author: Author{
			Name:  payload.Pusher.Name,
		},
	}

	var commits []Commit
	for _, commit := range payload.Commits {
		commits = append(commits, Commit{
			ID: commit.ID,
			Branch: pushPayload.GetBranch(),
			Repository: pushPayload.Repository,
			Author: Author{
				Name:     commit.Author.Name,
			},
			Added: commit.Added,
			Modified: commit.Modified,
			Removed: commit.Removed,
		})
	}

	pushPayload.Commits = commits

	return pushPayload
}

func NewIssuePayload(payload github.IssuesPayload) IssuePayload {
	return IssuePayload{
		Repository: Repository{
			Name: payload.Repository.Name,
			Owner: RepositoryOwner{
				Login: payload.Repository.Owner.Login,
			},
		},
		Number: int(payload.Issue.Number),
		Author: Author{
			Name:     payload.Issue.User.Login,
		},
		Body: payload.Issue.Body,
	}
}

func NewIssueCommentPayload(payload github.IssueCommentPayload) IssueCommentPayload {
	return IssueCommentPayload {
		IssuePayload: IssuePayload {
			Repository: Repository{
				Name:  payload.Repository.Name,
				Owner: RepositoryOwner{
					Login: payload.Repository.Owner.Login,
				},
			},
			Number: int(payload.Issue.Number),
			Author: Author{
				Name:     payload.Comment.User.Login,
			},
			Body: payload.Comment.Body,
		},
		CommentId: payload.Comment.ID,
	}
}
