package git

import (
	"fmt"
	"os"
	"sync"

	"github.com/Masterminds/vcs"
	"github.com/google/go-github/github"
	"github.com/jpillora/opts"
	"github.com/wxio/wx/internal/types"
	gogit "gopkg.in/src-d/go-git.v4"
)

func New(rt *types.Root) opts.Opts {
	gc := &gitCmd{rt: rt}
	return opts.New(rt).Name("git").
		AddCommand(opts.New(&cloneCmd{gc}).Name("clone"))
}

type gitCmd struct {
	rt    *types.Root
	repos []repo
}

type cloneCmd struct {
	gc *gitCmd
}

type repo struct {
	gitRepo     *gogit.Repository
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

func (cc *cloneCmd) Run() {
	l := len(cc.gc.rt.Workspaces)
	cc.gc.repos = make([]repo, l, l)
	fmt.Printf("%v\n", cc.gc.rt)
	var wg sync.WaitGroup
	for i, _ := range cc.gc.rt.Workspaces {
		ws := cc.gc.rt.Workspaces[i]
		repo := cc.gc.repos[i]
		var err error
		if repo.gitRepo, err = gogit.PlainOpen(ws.Path); err == nil {
			fmt.Fprintf(os.Stderr, "repo: %s ", ws.Name())
			fmt.Fprintf(os.Stderr, "  exists\n")
			continue
		}

		// vcs.Logger.SetOutput(os.Stdout)
		owner := ws.GitOwner
		if owner == "" {
			owner = cc.gc.rt.DefaultGitOwner
		}

		fmt.Printf("cloning %s %s\n", ws.Name(), ws.Address(owner))
		gitRepo, err := vcs.NewGitRepo(ws.Address(owner), ws.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "repo: %s ", ws.Name())
			fmt.Fprintf(os.Stderr, "  clone error : %v\n", err)
			continue
		}
		wg.Add(1)
		go func(repo types.Workspace) {
			defer wg.Done()
			err = gitRepo.Get()
			if err != nil {
				fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
				fmt.Fprintf(os.Stderr, "  clone error : %v\n", err)
				return
			}
			fmt.Printf("cloned %v\n", repo.Name())
		}(ws)
	}
	wg.Wait()
	fmt.Printf("-------\n")

}
