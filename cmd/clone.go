package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/Masterminds/vcs"

	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "clone all missing repo",
	Long: `based on .wx.<project.>yaml file, clone all repos that don't exist. 
	yaml format
	defaultGitOwner: <owner>
	repos:
	- path:  required
	  name:  default is last part of path
	  owner: defaults to <defaultGitOwner>
	  url:   defaults to git@github.com:<owner>/<name>.git>
	`,
	Args: clone.PositionalArgs,
	Run:  clone.Run,
	// func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("clone called")
	// },
}

type cloneData struct {
	my string
}

var clone = &cloneData{}

func (cd *cloneData) PositionalArgs(cmd *cobra.Command, args []string) error {
	cd.my = "hey hey hey"
	return nil
}
func (cd *cloneData) Run(cmd *cobra.Command, args []string) {
	fmt.Println("foo ", cd.my)

	var wg sync.WaitGroup
	for i, _ := range repos.Repos {
		repo := repos.Repos[i]
		var err error
		if repo.gitRepo, err = git.PlainOpen(repo.Path); err == nil {
			fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
			fmt.Fprintf(os.Stderr, "  exists\n")
			continue
		}

		// vcs.Logger.SetOutput(os.Stdout)
		fmt.Printf("cloning %s %s\n", repo.Name(), repo.Address())
		gitRepo, err := vcs.NewGitRepo(repo.Address(), repo.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
			fmt.Fprintf(os.Stderr, "  clone error : %v\n", err)
			continue
		}
		wg.Add(1)
		go func(repo *Repo) {
			defer wg.Done()
			err = gitRepo.Get()
			if err != nil {
				fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
				fmt.Fprintf(os.Stderr, "  clone error : %v\n", err)
				return
			}
			fmt.Printf("cloned %v\n", repo.Name())
		}(repo)
	}
	wg.Wait()
	fmt.Printf("-------\n")

}

func init() {
	rootCmd.AddCommand(cloneCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cloneCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cloneCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
