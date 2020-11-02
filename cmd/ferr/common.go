package main

import (
	"fmt"
	"strings"

	"github.com/chanwit/script"
	"github.com/mattn/go-shellwords"
	"github.com/mitchellh/go-homedir"
	"path/filepath"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	GIT              = "git"
	DefaultNamespace = "default"
)

func GitCurrentBranch() string {
	// Fails when not on a branch unlike: `git name-rev --name-only HEAD`
	output := script.Var()
	err := script.Exec(GIT, "symbolic-ref", "--short", "HEAD").To(output)
	if err != nil {
		return ""
	}

	return output.String()
}

func GitRemoteVerbose() string {
	output := script.Var()
	err := script.Exec(GIT, "remote", "-v").To(output)
	if err != nil {
		return ""
	}

	return output.RawString()
}

func GitRemote(branch string) string {
	output := script.Var()
	err := script.Exec(GIT, "config", "--get", fmt.Sprintf("branch.%s.remote", branch)).To(output)
	if err != nil {
		return ""
	}

	return output.String()
}

func GitRemoteFetchURL(remoteName string) string {
	output := script.Var()
	err := script.Exec(GIT, "config", "--get", fmt.Sprintf("remote.%s.url", remoteName)).To(output)
	if err != nil {
		return ""
	}
	return output.String()
}

func GitHTTPUrl(url string) string {
	return strings.Replace(url, "git@github.com:", "https://github.com/", 1)
}

func GitUrl(url string) string {
	return strings.Replace(url, "https://github.com/", "git@github.com:", 1)
}

func GitSSHUrl(url string) string {
	return strings.Replace(url, "https://github.com/", "ssh://git@github.com/", 1)
}

func readGitHubTokenAndUser() (user string, token string, err error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", "", err
	}
	return readGitHubTokenAndUserFromFile(filepath.Join(home, ".config", "gh", "hosts.yml"))
}

func readGitHubTokenAndUserFromFile(filename string) (user string, token string, err error) {
	const (
		githubHostname = "github.com"
	)

	obj, err := yaml.ReadFile(filename)
	if err != nil {
		return "", "", err
	}
	oauthTokenNode, err := obj.Pipe(yaml.Lookup(githubHostname), yaml.Get("oauth_token"))
	if err != nil {
		return "", "", err
	}
	token = yaml.GetValue(oauthTokenNode)
	userNode, err := obj.Pipe(yaml.Lookup(githubHostname), yaml.Get("user"))
	if err != nil {
		return "", "", err
	}
	user = yaml.GetValue(userNode)
	return user, token, err
}

func getImagesFromComposeFile(filename string) ([]string, error) {
	obj, err := yaml.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	services, err := obj.Pipe(yaml.Lookup("services"))
	if err != nil {
		return nil, err
	}

	images := []string{}
	shellwords.ParseEnv = true
	services.VisitFields(func(node *yaml.MapNode) error {
		image := node.Value.Field("image")
		if image != nil {
			results, err := shellwords.Parse(yaml.GetValue(image.Value))
			if err != nil {
				return err
			}
			images = append(images, strings.Join(results, ""))
		}
		return nil
	})

	return images, nil
}

func getServicesFromComposeFile(filename string) ([]string, error) {
	obj, err := yaml.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	services, err := obj.Pipe(yaml.Lookup("services"))
	if err != nil {
		return nil, err
	}

	serviceNames := []string{}
	services.VisitFields(func(node *yaml.MapNode) error {
		service := yaml.GetValue(node.Key)
		serviceNames = append(serviceNames, service)
		return nil
	})

	return serviceNames, nil
}

func detectDockerComposeFile() {

}

func detectDockerfile() {

}
