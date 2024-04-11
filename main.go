package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	cliVersion     = "[Not Provided]"
	buildDate      = "[Not Provided]"
	gitCommit      = "[Not Provided]"
	gitDescription = "[Not Provided]"
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
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "version and build information",
		Run:   runVersion,
	}

	cmd.SilenceUsage = true

	cmd.AddCommand(versionCmd)
	return cmd
}

func runVersion(cmd *cobra.Command, _ []string) {
	_, _ = fmt.Fprintf(cmd.OutOrStdout(),
		`CLIVersion:     %s
GitCommit:      %s
Build Date:     %s
GitDescription: %s
Platform:       %s/%s
GoVersion:      %s
Compiler:       %s
`,
		cliVersion,
		gitCommit,
		buildDate,
		gitDescription,
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
		runtime.Compiler,
	)
}
