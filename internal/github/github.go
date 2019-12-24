package github

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"github.com/jpillora/opts"
	"github.com/wxio/wx/internal/types"
	"golang.org/x/oauth2"
)

// New constructor
func New(rt *types.Root) opts.Opts {
	gc := &githubCmd{rt: rt}
	return opts.New(gc).Name("github").
		AddCommand(opts.New(&listCmd{gc: gc}).Name("list"))
}

type githubCmd struct {
	rt *types.Root
}

type listCmd struct {
	gc         *githubCmd
	HostSuffix string
	User       string
}

func (cc *listCmd) Run() error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	// list all repositories for the authenticated user
	repos, _, err := client.Repositories.List(ctx, cc.User, nil)
	if err != nil {
		return err
	}
	for _, rep := range repos {
		// fmt.Printf("%s\n", *rep.CloneURL)
		// continue
		fmt.Printf(`- path: go/src/github.com/%[1]s/%[2]s
  url: git@github.com%[3]s:%[1]s/%[2]s.git
`, *rep.Owner.Login, *rep.Name, cc.HostSuffix)
	}
	return nil
}
