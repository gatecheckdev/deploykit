package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/gatecheckdev/configkit"
	"github.com/gatecheckdev/deploykit/pkg/deploy"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	Version          = "[Not Provided]"
	BuildDate        = "[Not Provided]"
	GitCommit        = "[Not Provided]"
	DefaultMsgFormat = "deploykit: push IMAGE to SERVICE in SDIR"
)

var deployKitCmd = &cobra.Command{
	Use:   "deploykit",
	Short: "GitOps DeployKit - A utility for common GitOps tasks",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version and build information",
	Run:   runVersion,
}

var printConfigCmd = &cobra.Command{
	Use:   "print-config",
	Short: "output the config table in markdown format, used for documentation",
	Run:   runPrintConfig,
}

var printActionCmd = &cobra.Command{
	Use:   "print-action",
	Short: "output the GitHub Action for this CLI",
	Run:   runPrintAction,
}

var kustomizeCmd = &cobra.Command{
	Use:   "kustomize",
	Short: "update a manifest repository with a new image",
	Long: `This command wraps the Kustomize CLI 'edit set image' commmand

1. The target manifest repository is cloned or fetched if the dir flag is used
2. The kustomize set image 'service'='image name' command is run
3. The change is committed to the repository
4. The rebase, push loop with exponential back-off is started until successful`,
	RunE: runDeployKustomize,
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy using one of the supported methods/tools",
}

func NewDeployKitCmd() *cobra.Command {
	deployKitCmd.SilenceUsage = true

	RuntimeMetaConfig.Directory.SetupCobra(kustomizeCmd)
	RuntimeMetaConfig.Repository.SetupCobra(kustomizeCmd)
	RuntimeMetaConfig.Service.SetupCobra(kustomizeCmd)
	RuntimeMetaConfig.Image.SetupCobra(kustomizeCmd)
	RuntimeMetaConfig.Message.SetupCobra(kustomizeCmd)
	RuntimeMetaConfig.ServiceDirectory.SetupCobra(kustomizeCmd)
	RuntimeMetaConfig.SkipPush.SetupCobra(kustomizeCmd)

	deployCmd.AddCommand(kustomizeCmd)
	deployKitCmd.AddCommand(versionCmd, printConfigCmd, printActionCmd, deployCmd)

	return deployKitCmd
}

func runVersion(cmd *cobra.Command, _ []string) {
	_, _ = fmt.Fprintf(cmd.OutOrStdout(),
		`
Version:       %s
Git Commit:    %s
Build Date:    %s
Platform:      %s/%s
Go Version:    %s
Compiler:      %s
`,
		Version,
		GitCommit,
		BuildDate,
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
		runtime.Compiler,
	)
}

func renderMessage(s string) string {
	s = strings.ReplaceAll(s, "IMAGE", RuntimeMetaConfig.Image.Value().(string))
	s = strings.ReplaceAll(s, "SERVICE", RuntimeMetaConfig.Service.Value().(string))
	s = strings.ReplaceAll(s, "SDIR", RuntimeMetaConfig.ServiceDirectory.Value().(string))
	return s
}

func runDeployKustomize(cmd *cobra.Command, _ []string) error {
	var err error

	directory := RuntimeMetaConfig.Directory.Value().(string)
	repository := RuntimeMetaConfig.Repository.Value().(string)
	message := RuntimeMetaConfig.Message.Value().(string)
	message = renderMessage(message)
	serviceDirectory := RuntimeMetaConfig.ServiceDirectory.Value().(string)
	service := RuntimeMetaConfig.Service.Value().(string)
	image := RuntimeMetaConfig.Image.Value().(string)
	skipPush := RuntimeMetaConfig.SkipPush.Value().(bool)
	backoffMethod := RuntimeMetaConfig.BackoffMethod.Value().(string)
	retryAttempts := RuntimeMetaConfig.Attempts.Value().(int)

	shell := deploy.NewShell(deploy.WithWorkingDirectory(directory))

	// Determine if a repository should be cloned or if an existing repository needs to be pulled
	switch {
	case directory != "":
		err := shell.GitPullRebase()
		if err != nil {
			return err
		}
	case repository != "":
		directory, err = os.MkdirTemp("", "gdk-*")
		if err != nil {
			return err
		}
		err := shell.GitClone(repository, directory)
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
	err = shell.GitCommitAll(message)
	if skipPush {
		return nil
	}

	var backoffFunction deploy.BackoffFunc
	switch strings.TrimSpace(strings.ToLower(backoffMethod)) {
	case "exponential":
		backoffFunction = deploy.ExponentialBackoff(2)
	case "random":
		backoffFunction = deploy.RandomBackoff(5)
	default:
		slog.Warn("unsupported backoff method, defaulting to exponential", "method", backoffMethod)
	}

	return deploy.RebasePushLoop(shell, retryAttempts, time.Second, backoffFunction)
}

func runPrintConfig(cmd *cobra.Command, _ []string) {
	table := tablewriter.NewWriter(cmd.OutOrStderr())

	table.SetHeader([]string{"Name", "Field Type", "Default", "flag_name", "Env Variable Key", "Required"})
	configFields := configkit.AllMetaFields(RuntimeMetaConfig)
	for _, field := range configFields {
		required, _ := field.Metadata["required"]

		row := []string{
			field.FieldName,
			field.Metadata["field_type"],
			fmt.Sprintf("%v", field.DefaultValue),
			field.Metadata["flag_name"],
			field.EnvKey,
			required,
		}
		table.Append(row)
	}
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	// Markdown Format
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.Render()
}

func runPrintAction(cmd *cobra.Command, _ []string) {
	action := deploy.GitHubAction{
		Name:        "GitOps Deploykit",
		Description: "GitOps Style Manifest update with Kustomize",
		Inputs:      map[string]deploy.GitHubActionInput{},
		Runs: deploy.GitHubActionRuns{
			Using: "docker",
			Image: "Dockerfile",
			Env:   map[string]string{},
		},
	}

	for _, field := range configkit.AllMetaFields(RuntimeMetaConfig) {
		required := false
		if isRequired, ok := field.Metadata["required"]; ok {
			if isRequired == "Y" || isRequired == "Y*" {
				required = true
			}
		}
		input := deploy.GitHubActionInput{
			Description: field.Metadata["flag_usage"],
			Default:     fmt.Sprintf("%v", field.DefaultValue),
			Required:    required,
		}
		inputName := field.Metadata["action_input_name"]
		action.Inputs[inputName] = input
		action.Runs.Env[field.EnvKey] = fmt.Sprintf("${{ inputs.%s }}", inputName)
	}

	err := yaml.NewEncoder(cmd.OutOrStdout()).Encode(&action)
	if err != nil {
		slog.Error("failed encoding", "error", err)
	}
}
