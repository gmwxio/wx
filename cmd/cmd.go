package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// cmdCmd represents the sh command
var cmdCmd = &cobra.Command{
	Use:                "cmd <cmd + options>",
	Short:              "cmd of a repo named in .wx.proj.yaml",
	Long:               `parallel runs the sh cmd provided.`,
	Args:               cmdObj.PositionalArgs,
	DisableFlagParsing: true,
	Run:                cmdObj.Run,
}

type cmdType struct {
}

var cmdObj = &cmdType{}

func (s *cmdType) PositionalArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}
	return nil
}

func (s *cmdType) Run(cmd *cobra.Command, args []string) {
	fmt.Printf("cmd called %v\n", args)
	pexec(cmd, args, func(repo *Repo) {
		shArgs := []string{"/C", strings.Join(args, " ")}
		c := exec.Command("cmd.exe", shArgs...)
		c.Dir = repo.Path
		out, err := c.CombinedOutput()
		if err != nil {
			repo.out = fmt.Sprintf("error getting cmd %v %v\n", repo.Name(), err)
			return
		}
		repo.out = string(out)
	})
	printOutput()
}

func init() {
	rootCmd.AddCommand(cmdCmd)
}
