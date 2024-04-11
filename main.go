package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	os.Exit(run())
}

func run() int {
	cmd := newCommand()

	if err := cmd.Execute(); err != nil {
		return 1
	}

	return 0
}

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gtk",
		Short: "GitOps Toolkit - A utility for common GitOps tasks",
	}
	cmd.SilenceUsage = true
	return cmd
}
