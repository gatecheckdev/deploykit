package cmd

import (
	"strconv"

	"github.com/gatecheckdev/configkit"
	"github.com/spf13/cobra"
)

type metaConfig struct {
	Directory        configkit.MetaField
	Repository       configkit.MetaField
	Service          configkit.MetaField
	Image            configkit.MetaField
	Message          configkit.MetaField
	ServiceDirectory configkit.MetaField
	SkipPush         configkit.MetaField
	Attempts         configkit.MetaField
	BackoffMethod    configkit.MetaField
}

type CLIConfig struct {
	Directory        string
	Repository       string
	Service          string
	Image            string
	Message          string
	ServiceDirectory string
	SkipPush         bool
	Attempts         int
	BackoffMethod    string
}

var RuntimeMetaConfig = metaConfig{
	Directory: configkit.MetaField{
		FlagValueP:   new(string),
		FieldName:    "Directory",
		DefaultValue: "",
		EnvKey:       "DK_DIRECTORY",
		Metadata: map[string]string{
			"flag_name":         "directory",
			"flag_usage":        "The directory of an existing repository",
			"field_type":        "string",
			"required":          "Y*",
			"action_input_name": "directory",
		},
		CobraSetupFunc: func(f configkit.MetaField, cmd *cobra.Command) {
			cmd.Flags().StringVarP(
				f.FlagValueP.(*string),
				f.Metadata["flag_name"],
				"d",
				f.DefaultValue.(string),
				f.Metadata["flag_usage"],
			)
		},
	},

	Repository: configkit.MetaField{
		FlagValueP:   new(string),
		FieldName:    "Repository",
		DefaultValue: "",
		EnvKey:       "DK_REPOSITORY",
		Metadata: map[string]string{
			"flag_name":         "repository",
			"flag_usage":        "The directory of an existing repository",
			"field_type":        "string",
			"required":          "Y*",
			"action_input_name": "repository",
		},
		CobraSetupFunc: func(f configkit.MetaField, cmd *cobra.Command) {
			cmd.Flags().StringVarP(
				f.FlagValueP.(*string),
				f.Metadata["flag_name"],
				"r",
				f.DefaultValue.(string),
				f.Metadata["flag_usage"],
			)
		},
	},
	Service: configkit.MetaField{
		FlagValueP:   new(string),
		FieldName:    "Service",
		DefaultValue: "",
		EnvKey:       "DK_SERVICE",
		Metadata: map[string]string{
			"flag_name":         "service",
			"flag_usage":        "the destination service for the kustomize command",
			"field_type":        "string",
			"required":          "Y",
			"action_input_name": "service",
		},
		CobraSetupFunc: func(f configkit.MetaField, cmd *cobra.Command) {
			cmd.Flags().StringVarP(
				f.FlagValueP.(*string),
				f.Metadata["flag_name"],
				"s",
				f.DefaultValue.(string),
				f.Metadata["flag_usage"],
			)
		},
	},
	Image: configkit.MetaField{
		FlagValueP:   new(string),
		FieldName:    "Image",
		DefaultValue: "",
		EnvKey:       "DK_IMAGE",
		Metadata: map[string]string{
			"flag_name":         "image",
			"flag_usage":        "The container image name to use in the kustomize command",
			"field_type":        "string",
			"required":          "Y",
			"action_input_name": "image",
		},
		CobraSetupFunc: func(f configkit.MetaField, cmd *cobra.Command) {
			cmd.Flags().StringVarP(
				f.FlagValueP.(*string),
				f.Metadata["flag_name"],
				"i",
				f.DefaultValue.(string),
				f.Metadata["flag_usage"],
			)
		},
	},
	Message: configkit.MetaField{
		FlagValueP:   new(string),
		FieldName:    "Message",
		DefaultValue: DefaultMsgFormat,
		EnvKey:       "DK_MESSAGE",
		Metadata: map[string]string{
			"flag_name":         "message",
			"flag_usage":        "override the default git commit message",
			"field_type":        "string",
			"action_input_name": "message",
		},
		CobraSetupFunc: func(f configkit.MetaField, cmd *cobra.Command) {
			cmd.Flags().StringVarP(
				f.FlagValueP.(*string),
				f.Metadata["flag_name"],
				"m",
				f.DefaultValue.(string),
				f.Metadata["flag_usage"],
			)
		},
	},
	ServiceDirectory: configkit.MetaField{
		FlagValueP:   new(string),
		FieldName:    "ServiceDirectory",
		DefaultValue: "",
		EnvKey:       "DK_SERVICE_DIRECTORY",
		Metadata: map[string]string{
			"flag_name":         "service-directory",
			"flag_usage":        "The sub-directory (or environment) where the target kustomization.yaml file is located",
			"field_type":        "string",
			"action_input_name": "service_directory",
		},
		CobraSetupFunc: func(f configkit.MetaField, cmd *cobra.Command) {
			cmd.Flags().StringVarP(
				f.FlagValueP.(*string),
				f.Metadata["flag_name"],
				"e",
				f.DefaultValue.(string),
				f.Metadata["flag_usage"],
			)
		},
	},
	SkipPush: configkit.MetaField{
		FlagValueP:   new(bool),
		FieldName:    "SkipPush",
		EnvKey:       "DK_SKIP_PUSH",
		DefaultValue: false,
		EnvToValueFunc: func(s string) any {
			b, _ := strconv.ParseBool(s)
			return b
		},
		Metadata: map[string]string{
			"flag_name":         "skip-push",
			"flag_usage":        "Do the update but do not push commit",
			"field_type":        "bool",
			"action_input_name": "skip_push",
		},
		CobraSetupFunc: func(f configkit.MetaField, cmd *cobra.Command) {
			cmd.Flags().BoolVar(
				f.FlagValueP.(*bool),
				f.Metadata["flag_name"],
				f.DefaultValue.(bool),
				f.Metadata["flag_usage"],
			)
		},
	},
	Attempts: configkit.MetaField{
		FlagValueP:   new(int),
		FieldName:    "Attempts",
		EnvKey:       "DK_ATTEMPTS",
		DefaultValue: 3,
		EnvToValueFunc: func(s string) any {
			v, _ := strconv.Atoi(s)
			return v
		},
		Metadata: map[string]string{
			"flag_name":         "attempts",
			"flag_usage":        "Number of git push retry attempts",
			"field_type":        "int",
			"action_input_name": "attempts",
		},
		CobraSetupFunc: func(f configkit.MetaField, cmd *cobra.Command) {
			cmd.Flags().IntVar(
				f.FlagValueP.(*int),
				f.Metadata["flag_name"],
				f.DefaultValue.(int),
				f.Metadata["flag_usage"],
			)
		},
	},
	BackoffMethod: configkit.MetaField{
		FlagValueP:   new(string),
		FieldName:    "BackoffMethod",
		EnvKey:       "DK_BACKOFF_METHOD",
		DefaultValue: "random",
		Metadata: map[string]string{
			"flag_name":         "backoff-method",
			"flag_usage":        "the algorithm used to determine how long to wait before retry [exponential|random]",
			"field_type":        "string",
			"action_input_name": "backoff_method",
		},
		CobraSetupFunc: func(f configkit.MetaField, cmd *cobra.Command) {
			cmd.Flags().StringVar(
				f.FlagValueP.(*string),
				f.Metadata["flag_name"],
				f.DefaultValue.(string),
				f.Metadata["flag_usage"],
			)
		},
	},
}
