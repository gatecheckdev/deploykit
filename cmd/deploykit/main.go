package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/gatecheckdev/deploykit/cmd"
	"github.com/lmittmann/tint"
)

var (
	cliVersion = "[Not Provided]"
	buildDate  = "[Not Provided]"
	gitCommit  = "[Not Provided]"
)

var exitOK = 0
var exitErr = 1

func main() {
	os.Exit(run())
}

func run() int {
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
	})))
	cmd.Version = cliVersion
	cmd.BuildDate = buildDate
	cmd.GitCommit = gitCommit
	cmd := cmd.NewDeployKitCmd()

	if err := cmd.Execute(); err != nil {
		return exitErr
	}

	return exitOK
}
