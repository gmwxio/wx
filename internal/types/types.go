package types

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// Set by build tool chain
var (
	Version string = "dev"
	Date    string = "na"
	Commit  string = "na"
)

// Root struct for commands
// Field tags control where the values come from
// If opts:"-" yaml:"-" are set in object creation
//    opts:="-" come from config file
//    yaml:="-" come from command line flags
type Root struct {
	WorkspaceRoot   string      `opts:"-" yaml:"-"`
	CWD             string      `opts:"-" yaml:"-"`
	RootCfg         *Root       `opts:"-" yaml:"-"`
	GroupOutput     bool        `yaml:"-" help:"group by output. eg useful in 'wx -g sh git status --short'"`
	Parallelism     int         `yaml:"-" help:"max number of tasks to run in parallel. Defaults to 0 ie all"`
	DefaultGitOwner string      `opts:"-"`
	EmptyTag        bool        `yaml:"-" help:"if set the 'tags' are ignored"`
	Tags            []string    `yaml:",flow" help:"selects workspace based on tags. Tags are ANDed. Leading '~' is not.\n    eg tagged t1 and not t2 'wx -t t1 -t ~t2 sh pwd'"`
	NoHead          bool        `yaml:"-"`
	Workspaces      []Workspace `opts:"-"`
	// //
	// workspaces []*Workspace
}

// WorkspaceRoot for completions
var WorkspaceRoot string

type VersionCmd struct{}

type Workspace struct {
	Path     string
	GitOwner string   `json:"owner,omitempty" yaml:",omitempty"`
	RepoName string   `json:"name,omitempty"  yaml:",omitempty"`
	Url      string   `json:"url,omitempty"`
	Tags     []string `json:"tags,omitempty"  yaml:",omitempty,flow"`
	Out      string   `json:"-" opts:"-"  yaml:"-"`
}

func (r *VersionCmd) Run() {
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
func (root *Root) Configure(rcfg, rflg *Root) {
	root.GroupOutput = rflg.GroupOutput
	root.Parallelism = rflg.Parallelism
	root.NoHead = rflg.NoHead
	//
	root.DefaultGitOwner = rcfg.DefaultGitOwner
	//
	if !rflg.EmptyTag {
		if len(rflg.Tags) != 0 {
			root.Tags = rflg.Tags
		} else {
			root.Tags = rcfg.Tags
		}
	}
	root.tagMatcher(rcfg)
}

func (root *Root) tagMatcher(rcfg *Root) {
	if len(root.Tags) > 0 {
	OUTER:
		for i := range rcfg.Workspaces {
			w := rcfg.Workspaces[i]
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
			root.Workspaces = append(root.Workspaces, w)
		}
	} else {
		for i := range rcfg.Workspaces {
			w := rcfg.Workspaces[i]
			root.Workspaces = append(root.Workspaces, w)
		}
	}
}

func (root *Root) Pexec(args []string, f func(ws *Workspace)) {
	if len(root.Workspaces) == 0 {
		fmt.Printf("No workspaces for tags '%v'\n", root.Tags)
		return
	}
	pa := len(root.Workspaces)
	if root.Parallelism != 0 {
		pa = root.Parallelism
	}
	sem := make(chan bool, pa)
	var wg = sync.WaitGroup{}
	for i := range root.Workspaces {
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
		for i := range root.Workspaces {
			r := root.Workspaces[i]
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
		for i := range root.Workspaces {
			r := root.Workspaces[i]
			if !root.NoHead {
				fmt.Fprintf(os.Stderr, "--------- %s ---------\n", r.Name())
			}
			fmt.Printf("%s", r.Out)
		}
	}
}
