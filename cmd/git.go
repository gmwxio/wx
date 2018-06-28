package cmd

import (
	"fmt"

	"github.com/Masterminds/vcs"
	"github.com/spf13/cobra"
)

// gitCmd represents the git command
var gitCmd = &cobra.Command{
	Use:                "git <cmd + options>",
	Short:              "git of a repo named in .wx.proj.yaml",
	Long:               `parallel runs the git cmd provided.`,
	Args:               gitObj.PositionalArgs,
	DisableFlagParsing: true,
	// TraverseChildren: true,
	PreRun: gitObj.PreRun,
	Run:    gitObj.Run,
}

type gitType struct {
}

var gitObj = &gitType{}

func (s *gitType) PositionalArgs(cmd *cobra.Command, args []string) error {
	fmt.Printf("PositionalArgs %v\n", args)
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}
	return nil
}

func (s *gitType) PreRun(cmd *cobra.Command, args []string) {

}

func (s *gitType) Run(cmd *cobra.Command, args []string) {
	fmt.Printf("git called %v\n", args)
	// var mutex sync.Mutex
	pexec(cmd, args, func(repo *Repo) {
		gitRepo, err := vcs.NewGitRepo(repo.Address(), repo.Path)
		if err != nil {
			repo.out = fmt.Sprintf(" error : name: %v address: %v path: %v err: %v\n", repo.Name(), repo.Address(), repo.Path, err)
			return
		}

		// fmt.Printf("%v\n", args)
		out, err := gitRepo.RunFromDir("git", args...)
		if err != nil {
			repo.out = fmt.Sprintf("error getting git %v %v\n", repo.Name(), err)
			return
		}
		repo.out = string(out)
		// mutex.Lock()
		// fmt.Printf("repo %v------------\n%v", repo.Name(), string(out))
		// mutex.Unlock()
	})
	printOutput()
	// if *groupOutput {
	// 	omap := make(map[string][]string)
	// 	omapO := make([]string, 0)
	// 	for _, r := range repos.Repos {
	// 		if _, ex := omap[r.out]; !ex {
	// 			omapO = append(omapO, r.out)
	// 		}
	// 		omap[r.out] = append(omap[r.out], r.Name())
	// 	}
	// 	for _, k := range omapO {
	// 		fmt.Printf("\n----\n%v%v\n", k, omap[k])
	// 	}
	// } else {
	// 	for _, r := range repos.Repos {
	// 		fmt.Printf("--------- %s ---------:\n%s", r.Name(), r.out)
	// 	}
	// }
}

func init() {
	rootCmd.AddCommand(gitCmd)
	// groupOutput = gitCmd.Flags().BoolP("group-output", "g", false,
	// 	`Group the output and show all the repos with the same result`)

}
