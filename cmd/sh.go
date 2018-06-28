package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// shCmd represents the sh command
var shCmd = &cobra.Command{
	Use:                "sh <cmd + options>",
	Short:              "sh of a repo named in .wx.proj.yaml",
	Long:               `parallel runs the sh cmd provided.`,
	Args:               shObj.PositionalArgs,
	DisableFlagParsing: true,
	Run:                shObj.Run,
}

type shType struct {
}

var shObj = &shType{}

func (s *shType) PositionalArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}
	return nil
}

func (s *shType) Run(cmd *cobra.Command, args []string) {
	fmt.Printf("sh called %v\n", args)
	pexec(cmd, args, func(repo *Repo) {
		shArgs := []string{"-c", strings.Join(args, " ")}
		c := exec.Command("sh", shArgs...)
		c.Dir = repo.Path
		out, err := c.CombinedOutput()
		if err != nil {
			repo.out = fmt.Sprintf("error getting sh %v %v\n", repo.Name(), err)
			return
		}
		repo.out = string(out)
	})
	printOutput()
}

func init() {
	rootCmd.AddCommand(shCmd)
}
