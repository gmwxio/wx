package cmd

import (
	"bytes"
	"fmt"
	"os"
	"github.com/wxio/wx/cmd/porcelain"

	"github.com/Masterminds/vcs"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:                "status",
	Short:              "status of a repo named in .wx.proj.yaml",
	Long:               `parallel git status.`,
	Args:               status.PositionalArgs,
	DisableFlagParsing: true,
	Run:                status.Run,
}

type statusType struct {
}

var status = &statusType{}

func (s *statusType) PositionalArgs(cmd *cobra.Command, args []string) error {
	return nil
}

func (s *statusType) Run(cmd *cobra.Command, args []string) {
	fmt.Println("status called")
	pexec(cmd, args, func(repo *Repo) {
		gitRepo, err := vcs.NewGitRepo(repo.Address(), repo.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "repo: %s ", repo.Name())
			fmt.Fprintf(os.Stderr, "  clone error : %v\n", err)
			return
		}

		// fmt.Printf("%v\n", args)
		out, err := gitRepo.RunFromDir("git", "status", "--porcelain=v2", "--branch",
			// "--ignored=matching",
			"--ignore-submodules", "--no-lock-index")
		if err != nil {
			fmt.Printf("error getting status %v %v\n", repo.Name(), err)
			return
		}
		var pi = new(porcelain.PorcInfo)
		if err := pi.ParsePorcInfo(bytes.NewBuffer(out)); err != nil {
			fmt.Printf("porcelain error %v %v\n", repo.Name(), err)
			return
		}

		repo.out = fmt.Sprintf("%v\n", pi)
		// fmt.Printf("repo %v %v\n", repo.Name(), string(out))
	})
	printOutput()
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
