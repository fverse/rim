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
			fmt.Printf("Strap version: %v\n", "11")
		},
	}

	initCmd := cobra.Command{
		Use: "init",
		Run: Init,
	}
	cmds := &cobra.Command{Use: "strap"}
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

func Init(c *cobra.Command, args []string) {
	// create a config file the variables
	file, err := os.Create("strap.toml")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(Conf)
	if err != nil {
		panic(err)
	}
}

func Seed(c *cobra.Command, args []string) {}

func Unseed(c *cobra.Command, args []string) {}
