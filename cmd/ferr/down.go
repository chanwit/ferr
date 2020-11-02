package main

import (
	"github.com/chanwit/script"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var downCmd = &cobra.Command{
	Use:  "down",
	RunE: runDownCmd,
}

type DownParam struct {
	Path          string
	GitRemoteName string
	GitBranch     string
}

var (
	downClusterConfigPath string
)

func init() {
	downCmd.Flags().StringVar(&downClusterConfigPath, "path", "dev-cluster", "Path to the cluster GitOps configs")

	rootCmd.AddCommand(downCmd)
}

func down(param *DownParam) error {

	tmpDir, err := ioutil.TempDir("/tmp", "ferr-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	if err := script.Run("kompose", "convert",
		"-f", composeFilename,
		"--with-kompose-annotation=false",
		"-o", tmpDir); err != nil {
		return err
	}

	filesToDelete, err := ioutil.ReadDir(tmpDir)
	if err != nil {
		return err
	}

	// use these file names to delete
	for _, f := range filesToDelete {
		script.Run(GIT, "rm", filepath.Join(param.Path, DefaultNamespace, f.Name()))
	}

	script.Run(GIT, "commit", "-m", "remove application")
	script.Run(GIT, "push", param.GitRemoteName, param.GitBranch, "-u")

	time.Sleep(1 * time.Second)

	return script.Run("flux", "reconcile", "kustomization", "flux-system", "--with-source")
}

func runDownCmd(cmd *cobra.Command, args []string) error {
	branch := GitCurrentBranch()
	remoteName := GitRemote(branch)
	if remoteName == "" {
		remoteName = "origin"
	}

	return down(&DownParam{
		Path:          downClusterConfigPath,
		GitBranch:     branch,
		GitRemoteName: remoteName,
	})
}
