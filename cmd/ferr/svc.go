package main

import (
	"fmt"
	"github.com/chanwit/script"
	"github.com/spf13/cobra"
	"strings"

	"text/tabwriter"
)

var svcCmd = &cobra.Command{
	Use:     "service",
	Aliases: []string{"svc", "services", "ps"},
	RunE:    runSvcCmd,
}

func init() {
	rootCmd.AddCommand(svcCmd)
}

func runSvcCmd(cmd *cobra.Command, args []string) error {
	ip := script.Var()
	if err := script.Exec("kubectl", "get", "nodes", "-o",
		`jsonpath={.items[0].status.addresses[?(@.type=="InternalIP")].address}`).To(ip); err != nil {
		return err
	}

	nodeIP := ip.String()
	services, err := getServicesFromComposeFile(composeFilename)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(script.Stdout(), 0, 2, 3, ' ', 0)
	fmt.Fprintf(w, "SERVICE\tEND-POINT\n")
	for _, s := range services {
		result := script.Var()
		err := script.Exec("kubectl", "get", "service", s, "--output", "jsonpath={.spec.ports[*].nodePort}").To(result)
		if err == nil {
			nodePorts := strings.Split(result.String(), " ")
			for i, n := range nodePorts {
				if i == 0 {
					fmt.Fprintf(w, "%s\thttp://%s:%s\n", s, nodeIP, n)
				} else {
					fmt.Fprintf(w, "%s\thttp://%s:%s\n", "", nodeIP, n)
				}
			}
		}
	}
	w.Flush()
	return nil
}
