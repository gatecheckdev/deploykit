package main

import (
	"errors"
	"io"
	"log/slog"
	"math"
	"math/rand/v2"
	"os/exec"
	"time"
)

var randomNumberGenerator = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))

type Shell struct {
	stdout io.Writer
	stderr io.Writer
	dir    string
}

type shellOption func(s *Shell)

func WithWorkingDirectory(dir string) shellOption {
	return func(s *Shell) {
		s.dir = dir
	}
}

func NewShell(options ...shellOption) *Shell {
	shell := &Shell{
		stdout: globalDefaultStdout,
		stderr: globalDefaultStderr,
	}

	for _, optionFunc := range options {
		optionFunc(shell)
	}

	return shell
}

func (s *Shell) SetDir(dir string) {
	s.dir = dir
}

func (s *Shell) KustomizeEdit(arg string) error {
	return s.Run("kustomize", "edit", "set", "image", arg)
}

func (s *Shell) gitPullRebase() error {
	return s.Run("git", "pull", "--rebase")
}

func (s *Shell) gitCommitAll(msg string) error {
	return s.Run("git", "commit", "--all", "--message", msg)
}

func (s *Shell) gitPush() error {
	return s.Run("git", "push")
}

func (s *Shell) gitClone(repo string, dst string) error {
	return s.Run("git", "clone", "repo", dst)
}

func (s *Shell) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = s.stdout
	cmd.Stderr = s.stderr
	cmd.Dir = s.dir
	slog.Info("run", "command", cmd.String(), "working_directory", cmd.Dir)
	return cmd.Run()
}

type rebaser interface {
	gitPullRebase() error
}

type pusher interface {
	gitPush() error
}

type rebaserPusher interface {
	rebaser
	pusher
}

func rebasePushLoop(shell rebaserPusher, retryAttempts int, timeCoefficient time.Duration, backoff backoffFunc) error {
	slog.Info("git rebase and push loop", "retry_attempts", retryAttempts, "time_coefficient", timeCoefficient)
	if retryAttempts < 1 {
		slog.Warn("setting retry attempts to default of 1")
		retryAttempts = 1
	}

	attempted := 0

	sleepDurations := make([]int, retryAttempts, retryAttempts)
	backoff(sleepDurations)
	// Rebase / Push Loop
	for _, sleepDuration := range sleepDurations {
		attempted++
		err := shell.gitPullRebase()
		if err != nil {
			slog.Error("pull with rebase failure", "attempted", attempted)
			return err
		}

		err = shell.gitPush()

		// If the push worked
		if err == nil {
			slog.Info("success", "attempted", attempted)
			return nil
		}

		retryAfter := timeCoefficient * time.Duration(sleepDuration)
		slog.Warn("push attempt failed", "retry_after", retryAfter, "fail_error", err)
		time.Sleep(retryAfter)
	}
	return errors.New("all push attempts failed")
}

type backoffFunc func(sleepDurations []int)

func exponentialBackoff(base int) backoffFunc {
	return func(sleepDurations []int) {
		for i := range sleepDurations {
			sleepDurations[i] = int(math.Pow(float64(base), float64(i)))
		}
	}
}

func randomBackoff(max int) backoffFunc {
	return func(sleepDurations []int) {
		for i := range sleepDurations {
			sleepDurations[i] = randomNumberGenerator.IntN(max) + 1
		}
	}
}
