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
	Tags            []string `help:"selects workspace based on tags. Tags are ANDed. Leading '~' is not.\n    eg tagged t1 and not t2 'gb -t t1 -t ~t2 sh pwd'"`
	NoHead          bool
	Workspaces      []Workspace `opts:"-" json:"repo,omitempty"`
	//
	workspaces []*Workspace
}

type Workspace struct {
	Path     string
	GitOwner string   `json:"owner,omitempty"`
	RepoName string   `json:"name,omitempty"`
	Url      string   `json:"url,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Out      string   `json:"-" opts:"-"`
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

func (root *Root) TagMatcher() {
	if len(root.Tags) > 0 {
	OUTER:
		for i := range root.Workspaces {
			w := root.Workspaces[i]
			// && the tags
		HAS:
			for _, rt := range root.Tags {
				if rt[0] == '~' {
					for _, t := range w.Tags {
						if t == rt[1:] {
							continue OUTER
						}
					}
					continue HAS
				} else {
					for _, t := range w.Tags {
						if t == rt {
							continue HAS
						}
					}
				}
				continue OUTER
			}
			root.workspaces = append(root.workspaces, &w)
		}
	} else {
		for i := range root.Workspaces {
			w := root.Workspaces[i]
			root.workspaces = append(root.workspaces, &w)
		}
	}
}

func (root *Root) Pexec(args []string, f func(ws *Workspace)) {
	if len(root.workspaces) == 0 {
		fmt.Printf("No workspaces for tags '%v'\n", root.Tags)
		return
	}
	pa := len(root.workspaces)
	if root.Parallelism != 0 {
		pa = root.Parallelism
	}
	sem := make(chan bool, pa)
	var wg = sync.WaitGroup{}
	for _, w := range root.workspaces {
		sem <- true
		wg.Add(1)
		go func(repo *Workspace) {
			defer func() {
				<-sem
			}()
			f(repo)
			wg.Done()
		}(w)
	}
	wg.Wait()
}

func (root *Root) PrintOutput() {
	if root.GroupOutput {
		omap := make(map[string][]string)
		omapO := make([]string, 0)
		for _, r := range root.workspaces {
			if _, ex := omap[r.Out]; !ex {
				omapO = append(omapO, r.Out)
			}
			omap[r.Out] = append(omap[r.Out], r.Name())
		}
		for _, k := range omapO {
			if !root.NoHead {
				fmt.Fprintf(os.Stderr, `
---- Repositories: -------------------------------------------------------------------
%v
----    message:   -------------------------------------------------------------------
`, omap[k])
			}
			fmt.Printf("%s", k)
		}
	} else {
		for _, r := range root.workspaces {
			if !root.NoHead {
				fmt.Fprintf(os.Stderr, "--------- %s ---------\n", r.Name())
			}
			fmt.Printf("%s", r.Out)
		}
	}
}
