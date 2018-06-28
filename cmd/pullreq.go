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
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// pullreqCmd represents the pullreq command
var pullreqCmd = &cobra.Command{
	Use:   "pullreq branch \"title goes here\"",
	Short: "create pull requests from all repos that contain the named branch",
	Long: `
Create a pr from branch to master.
Only for remote repo that contain the branch.
If --check-local is specified then it check that the local repos and remote are is sync.

Checks to see that a pr doesn't already exist.
examples:
1.
wx pullreq <branchname> "pull request message"
2.
wx --dryrun pullreq --check-local <branchname> "pull request message"

Sample output:
Using config file: .wx.yaml
---------------
Not included
 repo: wxpb
---------------
Included
 repo: csharp   comment : moved info message
---------------
repo: csharp            WARNING:skipping - pull request already exist
	`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("no branch and title provided %d %v", len(args), args)
		}
		branch = args[0]
		title = strings.Join(args[1:], " ")
		return nil
	},
	Run: runPR,
}

func runPR(cmd *cobra.Command, args []string) {
	hasErr := false
	var remoteOps sync.WaitGroup
	// fmt.Printf("Not including\n")
	for i, _ := range repos.Repos {
		repo := repos.Repos[i]
		var err error
		if repo.gitRepo, err = git.PlainOpen(repo.Path); err != nil {
			fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
			fmt.Fprintf(os.Stderr, "  can't find repo. err: %v\n", err)
			// hasErr = true
			continue
		}

		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: oauth},
		)
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)

		remoteOps.Add(1)
		go remoteBranchCheck(repo, &remoteOps, &hasErr)
		remoteOps.Add(1)
		go getPullRequest(client, ctx, &remoteOps, repo)

		if *checkLocal {
			refIt, err := repo.gitRepo.Branches()
			if err != nil {
				fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
				fmt.Fprintf(os.Stderr, "  local repo cannot get branches err: %v\n", err)
				hasErr = true
				continue
			}
			err = refIt.ForEach(func(ref *plumbing.Reference) error {
				if ref.Name().String() == "refs/heads/"+branch {
					repo.localBranch.exits = true
					repo.localBranch.hash = ref.Hash()
					// repo.headsMatch = repo.remoteBranch.hash == ref.Hash()
				}
				return nil
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
				fmt.Fprintf(os.Stderr, "  local repo branch iteration err: %v\n", err)
				hasErr = true
				continue
			}
		}
	}
	remoteOps.Wait()
	fmt.Printf("---------------\n")
	fmt.Printf("Not included\n")
	for _, repo := range repos.Repos {
		if !repo.localBranch.exits && !repo.remoteBranch.exits {
			fmt.Printf(" repo: %v\n", repo.Name())
		}
	}
	fmt.Printf("---------------\n")
	fmt.Printf("Included\n")
	for _, repo := range repos.Repos {
		if repo.localBranch.exits || repo.remoteBranch.exits {
			fmt.Printf(" repo: %v \tcomment : %v\n", repo.Name(), strings.Trim(repo.remoteBranch.comment, " \n"))
		}
	}
	fmt.Printf("---------------\n")
	if hasErr {
		fmt.Fprintf(os.Stderr, "ERRORs:  not is a state for PR (check your ducks). Exiting\n")
		os.Exit(1)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: oauth},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	pull := &github.NewPullRequest{
		Base:  String("master"),
		Head:  String(branch),
		Title: String(title),
	}
	for i, _ := range repos.Repos {
		repo := repos.Repos[i]
		if *checkLocal && repo.localBranch.exits && !repo.remoteBranch.exits {
			fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
			fmt.Fprintf(os.Stderr, "\t\tWARNING:skipping - branch in local repo but not present in remote\n")
			continue
		}
		if *checkLocal && repo.localBranch.hash != repo.remoteBranch.hash {
			fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
			fmt.Fprintf(os.Stderr, "\t\tWARNING:skipping - local and remote tips don't match - push?\n")
			continue
		}
		if !repo.remoteBranch.exits {
			continue
		}
		if repo.pullreq != nil {
			fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
			fmt.Fprintf(os.Stderr, "\t\tWARNING:skipping - pull request already exist\n")
			continue
		}
		name := repo.RepoName
		owner := repo.GitOwner
		if name == "" {
			if idx := strings.LastIndex(repo.Path, "/"); idx > -1 {
				name = repo.Path[idx+1:]
			} else {
				name = repo.Path
			}
		}
		if owner == "" {
			owner = *repos.DefaultGitOwner
		}
		if *dryrun {
			fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
			fmt.Printf("\t\tdryrun - create pull request %v %v\n", owner, name)
			continue
		}

		pr, resp, err := client.PullRequests.Create(ctx, owner, name, pull)
		if err != nil {
			fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
			fmt.Fprintf(os.Stderr, "  ERROR: creating pull request: %v\n", err)
			hasErr = true
		} else {
			fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
			fmt.Fprintf(os.Stderr, "  pull request created\n")
		}
		fmt.Printf("  response %v\n", resp)
		fmt.Printf("  pull request %v\n", pr.GetHTMLURL())
	}
	if hasErr {
		fmt.Fprintf(os.Stderr, "ERRORs: couldn't do pull request. Exiting\n")
		os.Exit(1)
	}

}

func remoteBranchCheck(thisrepo *Repo, remoteOps *sync.WaitGroup, hasErr *bool) {
	defer remoteOps.Done()
	remote, err := thisrepo.gitRepo.Remote("origin")
	// remotes, err := thisrepo.gitRepo.Remotes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "repo: %s ", thisrepo.Name())
		fmt.Fprintf(os.Stderr, "  ERROR: %v\n", err)
		*hasErr = true
		return
	}
	list, err := remote.List(&git.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "repo: %s ", thisrepo.Name())
		fmt.Fprintf(os.Stderr, "  ERROR: %v\n", err)
		*hasErr = true
		return
	}
	for _, item := range list {
		if item.Name().String() == "refs/heads/"+branch {
			thisrepo.remoteBranch.hash = item.Hash()
			if thisrepo.remoteBranch.exits {
				fmt.Fprintf(os.Stderr, "repo: %s ", thisrepo.Name())
				fmt.Fprintf(os.Stderr, "  ERROR: remote branch exists more than once\n")
				*hasErr = true
			}
			thisrepo.remoteBranch.exits = true
			comment, err := thisrepo.gitRepo.CommitObject(item.Hash())
			if err != nil {
				fmt.Fprintf(os.Stderr, "repo: %s ", thisrepo.Name())
				fmt.Fprintf(os.Stderr, "  ERROR: getting comment err:%v\n", err)
			} else {
				thisrepo.remoteBranch.comment = comment.Message
			}
		}
	}
}

var checkLocal *bool

func init() {
	rootCmd.AddCommand(pullreqCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pullreqCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	checkLocal = pullreqCmd.Flags().BoolP("check-local", "l", false, `Check the local repo. 
See if head specified branch's hash == remote branch. 
i.e. check that all changes have been pushed`)
}
