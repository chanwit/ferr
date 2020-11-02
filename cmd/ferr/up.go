package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"strings"
	"time"

	"github.com/chanwit/script"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:  "up",
	RunE: runUpCmd,
}

type UpParam struct {
	Path          string
	GitBranch     string
	GitRemoteName string
}

var (
	upClusterConfigPath string
	noBuild             bool
	detach              bool
)

func init() {
	upCmd.Flags().StringVar(&upClusterConfigPath, "path", "dev-cluster", "Path to the cluster GitOps configs")
	upCmd.Flags().BoolVar(&noBuild, "no-build", false, "Not re-building images")
	upCmd.Flags().BoolVarP(&detach, "detach", "d", false, "Detached mode: run in background")

	rootCmd.AddCommand(upCmd)
}

func isInsideGitWorkTree() bool {
	err := script.Run("git", "rev-parse", "--is-inside-work-tree")
	return err == nil
}

func isGitHubAlreadyAuth() bool {
	status := script.Var()
	err := script.Exec("gh", "auth", "status").To(status)
	return err == nil
}

// UpdateFile reads the file at path, applies the filter to it, and write the result back.
// path must contain a exactly 1 resource (YAML).
func updateFile(path string, filters ...yaml.Filter) error {
	// Read the yaml
	y, err := yaml.ReadFile(path)
	if err != nil {
		return err
	}

	// Update the yaml
	if err := y.PipeE(filters...); err != nil {
		return err
	}

	// Write the yaml
	return yaml.WriteFile(y, path)
}

func up(param *UpParam) error {
	os.MkdirAll(filepath.Join(param.Path, DefaultNamespace), 0755)

	if noBuild == false {
		if err := build(&BuildParam{
			NoCache:         false,
			ComposeFilename: composeFilename,
		}); err != nil {
			return err
		}
	}

	if err := script.Run("kompose", "convert",
		"-f", composeFilename,
		"--with-kompose-annotation=false",
		"-o", filepath.Join(param.Path, DefaultNamespace)); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(filepath.Join(param.Path, DefaultNamespace))
	if err != nil {
		return err
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), "-service.yaml") {
			filename := filepath.Join(param.Path, DefaultNamespace, f.Name())
			fmt.Println("checking ", filename, "...")
			obj, err := yaml.ReadFile(filename)
			if err != nil {
				return err
			}

			typeNode, err := obj.Pipe(yaml.Lookup("spec", "type"))
			if err != nil {
				return err
			}

			if typeNode == nil {
				fmt.Println("service type not defined. patching to be NodePort.")
				err := updateFile(
					filepath.Join(param.Path, DefaultNamespace, f.Name()),
					yaml.Lookup("spec"),
					yaml.SetField("type", yaml.NewScalarRNode("NodePort")),
				)
				if err != nil {
					return err
				}
			}
		}
	}

	script.Run(GIT, "add", filepath.Join(param.Path, DefaultNamespace))
	script.Run(GIT, "commit", "-m", "add application")
	script.Run(GIT, "push", param.GitRemoteName, param.GitBranch, "-u")

	time.Sleep(1 * time.Second)

	script.Run("flux", "reconcile", "kustomization", "flux-system", "--with-source")

	// get all images
	// load into kind
	// script.Run("kubectl", "logs", "deployment/podinfo", "-f")

	if detach == false {
		return script.Run("stern", ".*")
	}

	// check if there's a Kubernetes cluster
	// check if it's a Git directory
	// check if there's a docker-compose.yaml
	return nil
}

func runUpCmd(cmd *cobra.Command, args []string) error {
	/*
		branch := GitCurrentBranch()
		remoteName := GitRemote(branch)
		if remoteName == "" {
			remoteName = "origin"
		}

		// test flux status
		if err := initialize(&InitParam{
			Path:          upClusterConfigPath,
			GitBranch:     branch,
			GitRemoteName: remoteName,
		}); err != nil {
			return err
		}
	*/

	branch := GitCurrentBranch()
	remoteName := GitRemote(branch)
	if remoteName == "" {
		remoteName = "origin"
	}

	// TODO check if the git repo is dirty
	return up(&UpParam{
		Path:          upClusterConfigPath,
		GitBranch:     branch,
		GitRemoteName: remoteName,
	})
}
