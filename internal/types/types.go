package types

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	Version string = "dev"
	Date    string = "na"
	Commit  string = "na"
)

type Root struct {
	WorkspaceRoot   string `opts:"-"`
	CWD             string `opts:"-"`
	GroupOutput     bool
	Parallelism     int
	DefaultGitOwner string
	Workspaces      []Workspace `opts:"-" json:"repo,omitempty"`
}

type Workspace struct {
	Path     string
	GitOwner string `json:"owner,omitempty"`
	RepoName string `json:"name,omitempty"`
	Url      string `json:"url,omitempty"`
	Out      string `json:"-" opts:"-"`
}

func (r *Root) Run() {
	fmt.Printf("version: %s\ndate: %s\ncommit: %s\n", Version, Date, Commit)
}

func (repo *Workspace) Name() string {
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

func (repo *Workspace) Address(owner string) string {
	url := repo.Url
	if url == "" {
		url = "git@github.com:" + owner + "/" + repo.Name() + ".git"
	}
	return url
}

func (root *Root) Pexec(args []string, f func(*Workspace)) {
	pa := len(root.Workspaces)
	if root.Parallelism != 0 {
		pa = root.Parallelism
	}
	sem := make(chan bool, pa)
	var wg = sync.WaitGroup{}
	for i, _ := range root.Workspaces {
		sem <- true
		wg.Add(1)
		go func(repo *Workspace) {
			defer func() {
				<-sem
			}()
			f(repo)
			wg.Done()
		}(&root.Workspaces[i])
	}
	wg.Wait()
}

func (root *Root) PrintOutput() {
	if root.GroupOutput {
		omap := make(map[string][]string)
		omapO := make([]string, 0)
		for _, r := range root.Workspaces {
			if _, ex := omap[r.Out]; !ex {
				omapO = append(omapO, r.Out)
			}
			omap[r.Out] = append(omap[r.Out], r.Name())
		}
		for _, k := range omapO {
			fmt.Fprintf(os.Stderr, `
---- Repositories: -------------------------------------------------------------------
%[2]v
----    message:   -------------------------------------------------------------------
%[1]v--------------------------------------------------------------------------------------
`, k, omap[k])
		}
	} else {
		for _, r := range root.Workspaces {
			fmt.Fprintf(os.Stderr, "--------- %s ---------:\n%s", r.Name(), r.Out)
		}
	}
}
