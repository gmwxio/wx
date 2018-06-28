package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	name string
	desc string
)

// createRepoCmd represents the createRepo command
var createRepoCmd = &cobra.Command{
	Use:   "createRepo name descriptio",
	Short: "Create a wxio repo with default config",
	Long: `Team, Master permission, etc.
dryrun NOT Implemented`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("no repo and description provided")
		}
		name = args[0]
		desc = strings.Join(args[1:], " ")
		return nil
	},
	Run: runCR,
}

func init() {
	rootCmd.AddCommand(createRepoCmd)
}

func runCR(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: oauth},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	repo := &github.Repository{
		Name:        String(name),
		Description: String(desc),
		Private:     Bool(true),
	}
	r, _, err := client.Repositories.Create(ctx, "wxio", repo)
	if err != nil {
		log.Fatalf("1. err %v\n", err)
	}
	fmt.Printf("created. %v\n", r.SSHURL)

	content := &github.RepositoryContentFileOptions{
		Message: String("Create README.md"),
		Branch:  String("master"),
		Content: []byte(fmt.Sprintf("# %s\n%s\n", name, desc)),
	}
	_, _, err = client.Repositories.CreateFile(ctx, "wxio", name, "README.md", content)
	if err != nil {
		log.Fatalf("Create file error %v\n", err)
	}
	fmt.Printf("Readme created.\n")

	pr := &github.ProtectionRequest{
		RequiredStatusChecks: &github.RequiredStatusChecks{
			Contexts: []string{},
		},
		RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
			DismissalRestrictionsRequest: &github.DismissalRestrictionsRequest{
			// Teams: []string{},
			// Users: []string{},
			},
		},
		Restrictions: &github.BranchRestrictionsRequest{
			Teams: []string{},
			Users: []string{},
		},
		EnforceAdmins: false,
	}
	_, _, err = client.Repositories.UpdateBranchProtection(ctx, "wxio", name, "master", pr)
	if err != nil {
		log.Fatalf("UpdateBranchProtection error %v\n", err)
	}
	fmt.Printf("UpdateBranchProtection.\n")

	_, _, err = AddRepoTeam(ctx, client, 2594881, "wxio", name)
	if err != nil {
		log.Fatalf("AddRepoTeam error %v\n", err)
	}
	fmt.Printf("AddRepoTeam.\n")
}

func AddRepoTeam(ctx context.Context, client *github.Client, team int, org string, repo string) (map[string]interface{}, *github.Response, error) {
	u := fmt.Sprintf("teams/%d/repos/%s/%s", team, org, repo)
	p := Perms{
		Permission: "push",
	}
	req, err := client.NewRequest("PUT", u, p)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", "application/vnd.github.hellcat-preview+json")

	t := make(map[string]interface{})
	resp, err := client.Do(ctx, req, &t)
	if err != nil {
		return nil, resp, err
	}

	return t, resp, nil
}

type Perms struct {
	Permission string `json:"permission"`
}
