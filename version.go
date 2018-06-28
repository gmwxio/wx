


package main

import (
	"fmt"
	"github.com/wxio/wx/cmd"

	"github.com/spf13/cobra"
)

var version = "master"
var commit = "HEAD"
var date = "now"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "makes coffee",
	Long: `

Version in inserted by goreleaser via go build
	ldflags: -X cmd.version={{.Version}} -X cmd.commit={{.Commit}} -X cmd.date={{.Date}}
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if *versionObj.long {
			fmt.Printf("version: %s\ncommit: %s\ndate: %s\n", version, commit, date)
			return
		}
		fmt.Printf("version: %s\n", version)
	},
}

var versionObj = struct {
	long *bool
}{}

func init() {
	cmd.RootCmd().AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	versionObj.long = versionCmd.Flags().BoolP("long", "l", false, "long")
}
