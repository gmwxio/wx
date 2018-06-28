package cmd

import (
	"os"
	"strings"
	"sync"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
)

type WXConfig struct {
	Project string
	Matches struct {
		GitOrigin *string `yaml:"git.origin"`
	}
}

type Repo struct {
	Path     string
	GitOwner string `yaml:"owner,omitempty"`
	RepoName string `yaml:"name,omitempty"`
	Url      string `yaml:"url,omitempty"`

	gitRepo     *git.Repository
	localBranch struct {
		exits bool
		hash  [20]byte
	}
	remoteBranch struct {
		comment string
		exits   bool
		hash    [20]byte
	}
	pullreq *github.PullRequest
	out     string
}
type Repos struct {
	DefaultGitOwner *string `yaml:"defaultGitOwner,omitempty"`
	Repos           []*Repo
}

func (repo *Repo) Name() string {
	name := repo.RepoName
	if repo.Path == "." || repo.RepoName == "." {
		name, _ = os.Getwd()
		name = strings.Replace(name, "\\", "/", -1)
		if idx := strings.LastIndex(name, "/"); idx > -1 {
			name = name[idx+1:]
		}
	}
	if name == "" {
		if idx := strings.LastIndex(repo.Path, "/"); idx > -1 {
			name = repo.Path[idx+1:]
		} else {
			name = repo.Path
		}
	}
	return name
}

func (repo *Repo) Owner() string {
	owner := repo.GitOwner
	if owner == "" {
		owner = *repos.DefaultGitOwner
	}
	return owner
}

func (repo *Repo) Address() string {
	url := repo.Url
	if url == "" {
		url = "git@github.com:" + repo.Owner() + "/" + repo.Name() + ".git"
	}
	return url
}

func pexec(cmd *cobra.Command, args []string, f func(*Repo)) {
	pa := len(repos.Repos)
	if parallelism != nil && *parallelism != 0 {
		pa = *parallelism
	}
	sem := make(chan bool, pa)
	var wg = sync.WaitGroup{}
	for i, _ := range repos.Repos {
		sem <- true
		wg.Add(1)
		go func(repo *Repo) {
			defer func() {
				<-sem
			}()
			f(repo)
			wg.Done()
		}(repos.Repos[i])
	}
	wg.Wait()
}

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool { return &v }

// Int is a helper routine that allocates a new int value
// to store v and returns a pointer to it.
func Int(v int) *int { return &v }

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string { return &v }
