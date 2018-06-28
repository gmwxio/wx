package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

func init() {
	dumpCmd := &cobra.Command{
		Use:   "dump",
		Short: "Print the .wx.yaml file in yaml format",
		Long:  `Of no real use now. This is the start of a utility to create and modify the config file`,
		Run: func(cmd *cobra.Command, args []string) {
			out, err := yaml.Marshal(&repos)
			if err != nil {
				log.Fatalf("marshall: %v", err)
				os.Exit(1)
			}
			fmt.Println(string(out))
		},
	}
	rootCmd.AddCommand(dumpCmd)
}
