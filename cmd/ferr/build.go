package main

import (
	"fmt"
	"github.com/chanwit/script"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:  "build",
	RunE: runBuildCmd,
}

func init() {
	rootCmd.AddCommand(buildCmd)
}

type BuildParam struct {
	NoCache         bool
	ComposeFilename string
}

func boolToString(val bool) string {
	if val {
		return "true"
	} else {
		return "false"
	}
}

func boolFlag(flag string, val bool) string {
	if val {
		return flag
	}
	return ""
}

func build(param *BuildParam) error {
	if err := script.Run(
		"docker-compose",
		"-f", composeFilename,
		"pull"); err != nil {
		return err
	}

	argsForBuild := []string{"-f", composeFilename, "build"}
	if param.NoCache {
		argsForBuild = append(argsForBuild, "--no-cache")
	}
	if err := script.Run(
		"docker-compose", argsForBuild...); err != nil {
		return err
	}

	images, err := getImagesFromComposeFile(param.ComposeFilename)
	if err != nil {
		return err
	}

	for _, image := range images {
		err := script.Run("kind", "load", "docker-image", image)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func runBuildCmd(cmd *cobra.Command, args []string) error {
	return build(&BuildParam{
		NoCache:         true,
		ComposeFilename: composeFilename,
	})
}
