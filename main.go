package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"log/slog"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

var (
	cliVersion     = "[Not Provided]"
	buildDate      = "[Not Provided]"
	gitCommit      = "[Not Provided]"
	gitDescription = "[Not Provided]"
)

var globalDefaultStdout = os.Stdout
var globalDefaultStderr = os.Stderr
var globalDefaultMsg = "gtk deploy: service %s update image to %s"

func main() {
	os.Exit(run())
}

func run() int {
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
	})))

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
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy using one of the supported methods/tools",
	}
	kustomizeCmd := &cobra.Command{
		Use:   "kustomize",
		Short: "update a manifest repository with a new image",
		Long: `This command wraps the Kustomize CLI 'edit set image' commmand

1. The target manifest repository is cloned or fetched if the dir flag is used
2. The kustomize set image 'service'='image name' command is run
3. The change is committed to the repository
4. The rebase, push loop with exponential back-off is started until successful`,
		RunE: runDeployKustomize,
	}

	cmd.SilenceUsage = true
	kustomizeCmd.Flags().String("directory", "", "The directory of an existing repository")
	kustomizeCmd.Flags().String("repository", "", "The target repository to clone using git")
	kustomizeCmd.Flags().String("service", "", "The destination service for the kustomize command")
	kustomizeCmd.Flags().String("image", "", "The container image name to use in the kustomize command")
	kustomizeCmd.Flags().String("message", "", "The commit message to use for the deployment commit")
	kustomizeCmd.Flags().String("service-directory", "", "The sub-directory (or environment) where the target kustomization.yaml file is located")
	kustomizeCmd.Flags().Bool("skip-push", false, "Do not push commit")
	kustomizeCmd.Flags().Int("attempts", 3, "Number of push retry attempts")
	kustomizeCmd.Flags().String("backoff-method", "exponential", "the algorithm used to determine how long to wait before retry [exponential|random]")

	deployCmd.AddCommand(kustomizeCmd)
	cmd.AddCommand(versionCmd, deployCmd)
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

func runDeployKustomize(cmd *cobra.Command, _ []string) error {
	var err error

	directory, _ := cmd.Flags().GetString("directory")
	repository, _ := cmd.Flags().GetString("repository")
	service, _ := cmd.Flags().GetString("service")
	image, _ := cmd.Flags().GetString("image")
	message, _ := cmd.Flags().GetString("message")
	skipPush, _ := cmd.Flags().GetBool("skip-push")
	serviceDirectory, _ := cmd.Flags().GetString("service-directory")
	retryAttempts, _ := cmd.Flags().GetInt("attempts")
	backoffMethod, _ := cmd.Flags().GetString("backoff-method")

	if message == "" {
		message = fmt.Sprintf(globalDefaultMsg, service, image)
	}

	shell := NewShell(WithWorkingDirectory(directory))

	// Determine if a repository should be cloned or if an existing repository needs to be pulled
	switch {
	case directory != "":
		err := shell.gitPullRebase()
		if err != nil {
			return err
		}
	case repository != "":
		directory, err = os.MkdirTemp("", "gtk-*")
		if err != nil {
			return err
		}
		err := shell.gitClone(repository, directory)
		if err != nil {
			return err
		}
	default:
		return errors.New("need an existing repsitory directory or repository url to clone")
	}

	shell.SetDir(path.Join(directory, serviceDirectory))
	err = shell.KustomizeEdit(fmt.Sprintf("%s=%s", service, image))
	if err != nil {
		return err
	}

	shell.SetDir(directory)
	err = shell.gitCommitAll(message)
	if skipPush {
		return nil
	}

	var backoffFunction backoffFunc
	switch strings.TrimSpace(strings.ToLower(backoffMethod)) {
	case "exponential":
		backoffFunction = exponentialBackoff(2)
	case "random":
		backoffFunction = randomBackoff(5)
	default:
		slog.Warn("unsupported backoff method, defaulting to exponential", "method", backoffMethod)
	}

	return rebasePushLoop(shell, retryAttempts, time.Second, backoffFunction)
}
