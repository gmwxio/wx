package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var githubChmodCmd = &cobra.Command{
	Use:   "githubChmod name",
	Short: "Sets the perms on a github repo",
	Long: `Team, Master permission, etc.
dryrun NOT Implemented`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("no repo provided")
		}
		name = args[0]
		return nil
	},
	Run: runChmod,
}

func init() {
	rootCmd.AddCommand(githubChmodCmd)
}

func runChmod(cmd *cobra.Command, args []string) {
	var err error
	oauth := viper.GetString("GITHUB_API_TOKEN")
	if oauth == "" {
		fmt.Fprintf(os.Stderr, "now git access token found\n")
		os.Exit(1)
	}
	name := args[0]
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: oauth},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

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
