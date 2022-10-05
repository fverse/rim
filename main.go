package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	versionCmd := cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("cheeder version: %v\n", VERSION)
		},
	}

	initCmd := cobra.Command{
		Use: "seeder:create",
		Run: Create,
	}
	cmds := &cobra.Command{Use: "cheeder"}
	cmds.AddCommand(&versionCmd, &initCmd)
	cmds.Execute()
}

func Create(c *cobra.Command, args []string) {
	var dir string
	switch len(args) {
	case 0:
		dir = "."
	case 1:
		dir = args[0]
	default:
		{
			c.Help()
			os.Exit(1)
		}
	}
	fmt.Printf("dir: %v\n", dir)
	// TODO: write seeder file
}

func Seed(c *cobra.Command, args []string) {}

func Unseed(c *cobra.Command, args []string) {}
