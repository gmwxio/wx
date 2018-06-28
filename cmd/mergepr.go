package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// mergeprCmd represents the mergepr command
var mergeprCmd = &cobra.Command{
	Use:   "mergepr branch comment",
	Short: "Merge outstanding pull requests for all repositories with named branch",
	Long: `
	
Sample runs
>wx --dryrun=true mergepr branchname "message"
	Using config file: .wx.yaml
	repo: csharp   WANRING: pr not open
	repo: fud   WANRING: pr not open
	repo: fus   dryrun: merge owner: wxio name: fus #:1
	repo: modal   WANRING: pr not open
	repo: page-fu   WANRING: pr not open	

>wx mergepr multiple-modal  branchname "message"
	Using config file: .wx.yaml
	repo: csharp            WANRING: pr not open
	repo: fud               WANRING: pr not open
	merge result: true message : Pull Request successfully merged
	repo: modal             WANRING: pr not open
	repo: page-fu           WANRING: pr not open
	`,
	Run: mergePR,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("no branch and title provided %d %v", len(args), args)
		}
		branch = args[0]
		title = strings.Join(args[1:], " ")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mergeprCmd)
}

func mergePR(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: oauth},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	{
		var remoteOps sync.WaitGroup
		for i, _ := range repos.Repos {
			remoteOps.Add(1)
			go getPullRequest(client, ctx, &remoteOps, repos.Repos[i])
		}
		remoteOps.Wait()
	}
	{
		opts := &github.PullRequestOptions{}
		for i, _ := range repos.Repos {
			repo := repos.Repos[i]
			if repo.pullreq != nil {
				// by, _ := json.MarshalIndent(repo.pullreq, "", "\t")
				// fmt.Printf("\n\n%+v\n", string(by))
				if *repo.pullreq.State != "open" {
					fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
					fmt.Fprintf(os.Stderr, "\t\tWANRING: pr not open\n")
					continue
				}
				number := repo.pullreq.GetNumber()
				if *dryrun {
					fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
					fmt.Fprintf(os.Stderr, "\t\tdryrun: merge owner: %s name: %s #:%d\n", repo.Owner(), repo.Name(), number)
					continue
				}
				prmr, _, err := client.PullRequests.Merge(ctx, repo.Owner(), repo.Name(), number, title, opts)
				if err != nil {
					fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
					fmt.Fprintf(os.Stderr, "\t\tERROR: merger : %v\n", err)
					return
				}
				fmt.Fprintf(os.Stderr, "merge result: %v message : %v\n", prmr.GetMerged(), prmr.GetMessage())
			}
		}
	}
}

func getPullRequest(client *github.Client, ctx context.Context, remoteOps *sync.WaitGroup, repo *Repo) {
	opts := &github.PullRequestListOptions{
		State: "all",

		// head user:branch might be useful
	}
	prs, _, err := client.PullRequests.List(ctx, repo.Owner(), repo.Name(), opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
		fmt.Fprintf(os.Stderr, "  ERROR: getting PR list: %v\n", err)
		return
	}
	for _, pr := range prs {
		if branch == pr.Head.GetRef() {
			if repo.pullreq != nil {
				fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
				fmt.Fprintf(os.Stderr, "  ERROR: multiple prs with same head branches\n")
			}
			repo.pullreq = pr
		}
	}
	remoteOps.Done()
}
