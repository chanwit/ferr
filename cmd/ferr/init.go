package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/chanwit/script"
	"github.com/cli/cli/pkg/prompt"
	"github.com/spf13/cobra"
	giturls "github.com/whilp/git-urls"
)

var initCmd = &cobra.Command{
	Use:  "init",
	RunE: runInitCmd,
}

type InitParam struct {
	Path          string
	GitBranch     string
	GitRemoteName string
}

var (
	initClusterConfigPath string
)

func init() {
	initCmd.Flags().StringVar(&initClusterConfigPath, "path", "dev-cluster", "Path to the cluster GitOps configs")

	rootCmd.AddCommand(initCmd)
}

func initialize(param *InitParam) error {
	if err := script.Run("kubectl", "version", "--short"); err != nil {
		fmt.Println("cannot connect to Kubernetes")
		createDevCluster := 0
		err := prompt.SurveyAskOne(&survey.Select{
			Message: "No Kubernetes running.\nWould you like to create a Dev cluster using Kind?",
			Options: []string{
				"No",
				"Yes",
			},
		}, &createDevCluster)
		if err != nil {
			return err
		}
		if createDevCluster == 0 {
			fmt.Println("please start your Kubernetes cluster and run `ferr up` again")
			return nil
		} else if createDevCluster == 1 {
			if err := script.Run("kind", "create", "cluster"); err != nil {
				return err
			}
		}
	}

	gitJustInited := false
	if isInsideGitWorkTree() == false {
		initGitRepo := 0
		err := prompt.SurveyAskOne(&survey.Select{
			Message: "No git repository detected.\nWould you like to initialize a Git repository here?",
			Options: []string{
				"No",
				"Yes",
			},
		}, &initGitRepo)
		if err != nil {
			return nil
		}

		if initGitRepo == 0 {
			// exit
			return nil
		} else {
			err := script.Run("git", "init")
			if err != nil {
				return err
			}
			gitJustInited = true
		}
	}

	if isGitHubAlreadyAuth() == false {
		if _, err := exec.LookPath("gh"); err != nil {
			fmt.Println("no gh cli exists")
		} else {
			authWithGH := 0
			err := prompt.SurveyAskOne(&survey.Select{
				Message: "Would you like to login GitHub with GH cli?",
				Options: []string{
					"Yes",
					"No, I'll set GITHUB_TOKEN manually",
				},
			}, &authWithGH)
			if err != nil {
				return err
			}
			if authWithGH == 0 {
				if err := script.Run("gh", "auth", "login"); err != nil {
					return nil
				}
			} else {
				fmt.Println("please set GITHUB_TOKEN environment variable and run gaw up again")
				return nil
			}
		}
	}

	if gitJustInited {
		fmt.Println("create remote of the same dir name")
		if dir, err := os.Getwd(); err == nil {
			count := 0

		recreate:
			dirName := filepath.Base(dir)
			if count > 0 {
				dirName = fmt.Sprintf("%s-%d", dirName, count)
			}
			if err := script.Run("gh", "repo", "create", dirName, "-y", "--private"); err != nil {
				count++
				goto recreate
			}
		}

		if err := script.Run(GIT, "add", "."); err != nil {
			return err
		}

		if err := script.Run(GIT, "commit", "-m", "init"); err != nil {
			return err
		}

		// re-init
		if param.GitBranch == "" {
			param.GitBranch = GitCurrentBranch()
		}
		if err := script.Run(GIT, "push", param.GitRemoteName, param.GitBranch, "-u"); err != nil {
			return err
		}
	}

	// we expect a clean working tree

	remoteUrl := GitRemoteFetchURL(param.GitRemoteName)
	u, err := giturls.Parse(remoteUrl)
	if err != nil {
		return err
	}
	parts := strings.SplitN(u.Path, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("unexpected %s", u.Path)
	}
	owner := parts[0]
	name := parts[1]
	if strings.HasSuffix(name, ".git") {
		name = strings.TrimSuffix(name, ".git")
	}

	_, token, err := readGitHubTokenAndUser()
	if err != nil {
		return err
	}

	// if flux check returns an error, try bootstrapping it
	if err := script.Run("flux", "check"); err != nil {
		if err := script.
			Export("GITHUB_TOKEN="+token).
			Exec("flux", "bootstrap", "github",
				"--branch="+param.GitBranch,
				"--owner="+owner,
				"--repository="+name,
				"--path="+param.Path).
			Tee(os.Stdout).
			Run(); err != nil {
			return err
		}
	}

	if err := script.Run(GIT, "pull", param.GitRemoteName, param.GitBranch); err != nil {
		return nil
	}

	return nil
}

func runInitCmd(cmd *cobra.Command, args []string) error {
	branch := GitCurrentBranch()
	remoteName := GitRemote(branch)
	if remoteName == "" {
		remoteName = "origin"
	}

	return initialize(&InitParam{
		Path:          initClusterConfigPath,
		GitBranch:     branch,
		GitRemoteName: remoteName,
	})
}
