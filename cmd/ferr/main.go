package main

import (
	"github.com/chanwit/script"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var VERSION = "0.0.0-dev.0"

var rootCmd = &cobra.Command{
	Use:              "ferr",
	Version:          VERSION,
	SilenceUsage:     true,
	SilenceErrors:    true,
	PersistentPreRun: preRun,
}

var (
	kubeconfig      string
	timeout         time.Duration
	verbose         bool
	composeFilename string
)

func init() {
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 5*time.Minute, "timeout for this operation")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "run verbosely")
	rootCmd.PersistentFlags().StringVarP(&composeFilename, "filename", "f", "docker-compose.yaml", "compose filename")
}

func preRun(cmd *cobra.Command, args []string) {
	if verbose {
		script.Debug = true
	}
}

func main() {
	log.SetFlags(0)
	generateDocs()
	kubeconfigFlag()
	if err := rootCmd.Execute(); err != nil {
		// logger.Failuref("%v", err)
		os.Exit(1)
	}
}

func kubeconfigFlag() {
	if home, err := homedir.Dir(); err != nil {
		rootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "", filepath.Join(home, ".kube", "config"), "path to the kubeconfig file")
	} else {
		rootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "", "", "absolute path to the kubeconfig file")
	}

	if len(os.Getenv("KUBECONFIG")) > 0 {
		kubeconfig = os.Getenv("KUBECONFIG")
	}
}

func generateDocs() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "docgen" {
		rootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "", "~/.kube/config", "path to the kubeconfig file")
		rootCmd.DisableAutoGenTag = true
		err := doc.GenMarkdownTree(rootCmd, "./docs/cmd")
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}
}
