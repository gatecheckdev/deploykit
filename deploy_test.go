package main

import (
	"errors"
	"log/slog"
	"math/rand/v2"
	"os"
	"testing"
	"time"

	"github.com/lmittmann/tint"
)

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
	})))
	randomNumberGenerator = rand.New(rand.NewPCG(0, 0))
	m.Run()
}

func TestExponentialBackoff_0(t *testing.T) {
	testCase := make([]int, 0, 0)
	want := []int{}
	exponentialBackoff(2)(testCase)

	for i := range want {
		if want[i] != testCase[i] {
			t.Logf("%v does not match %v", want, testCase)
			t.Fatalf("want: %d got: %d", want[i], testCase[i])
		}
	}
}

func TestExponentialBackoff_10(t *testing.T) {
	testCase := make([]int, 10, 10)
	want := []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512}
	exponentialBackoff(2)(testCase)

	for i := range want {
		if want[i] != testCase[i] {
			t.Logf("%v does not match %v", want, testCase)
			t.Fatalf("want: %d got: %d", want[i], testCase[i])
		}
	}
}

func TestRandomBackoff_0(t *testing.T) {
	testCase := make([]int, 0, 0)
	want := []int{}
	randomBackoff(10)(testCase)

	for i := range want {
		if want[i] != testCase[i] {
			t.Logf("%v does not match %v", want, testCase)
			t.Fatalf("want: %d got: %d", want[i], testCase[i])
		}
	}
}

func TestRandomBackoff_10(t *testing.T) {
	testCase := make([]int, 10, 10)
	randomBackoff(10)(testCase)

	for i := range testCase {
		if testCase[i] == 0 {
			t.Logf("no values should be 0: %v", testCase)
		}
	}
}

type mockRebaserPusher struct {
	pushFunc func() error
	pullFunc func() error
}

func (m *mockRebaserPusher) gitPush() error {
	return m.pushFunc()
}

func (m *mockRebaserPusher) gitPullRebase() error {
	return m.pullFunc()
}

func TestRebasePushLoop_Success(t *testing.T) {
	shell := &mockRebaserPusher{
		pushFunc: func() error {
			return nil
		},
		pullFunc: func() error {
			return nil
		},
	}
	err := rebasePushLoop(shell, 1, time.Second, exponentialBackoff(2))

	if err != nil {
		t.Fatal(err)
	}
}

func TestRebasePushLoop_FailureExponentialBackoff(t *testing.T) {
	// Test is expected to fail because of the number of retry attempt
	failCounter := 0
	shell := &mockRebaserPusher{
		pushFunc: func() error {
			if failCounter == 10 {
				return nil
			}
			failCounter++
			return errors.New("mock error")
		},
		pullFunc: func() error {
			return nil
		},
	}
	err := rebasePushLoop(shell, 5, time.Nanosecond, exponentialBackoff(2))

	if err == nil {
		t.Fatal("expected an error, number of attempts is 5 and push would only success at 10")
	}
}

func TestRebasePushLoop_SuccessExponentialBackOff(t *testing.T) {
	// Test is expected to fail because of the number of retry attempt
	failCounter := 0
	shell := &mockRebaserPusher{
		pushFunc: func() error {
			if failCounter == 10 {
				return nil
			}
			failCounter++
			return errors.New("mock error")
		},
		pullFunc: func() error {
			return nil
		},
	}
	err := rebasePushLoop(shell, 50, time.Nanosecond, exponentialBackoff(2))

	if err != nil {
		t.Fatal(err)
	}
}

func TestRebasePushLoop_FailureRandomBackoff(t *testing.T) {
	// Test is expected to fail because of the number of retry attempt
	failCounter := 0
	shell := &mockRebaserPusher{
		pushFunc: func() error {
			if failCounter == 10 {
				return nil
			}
			failCounter++
			return errors.New("mock error")
		},
		pullFunc: func() error {
			return nil
		},
	}
	err := rebasePushLoop(shell, 5, time.Nanosecond, randomBackoff(100))

	if err == nil {
		t.Fatal("expected an error, number of attempts is 5 and push would only success at 10")
	}
}

func TestRebasePushLoop_SuccessRandomBackoff(t *testing.T) {
	failCounter := 0
	shell := &mockRebaserPusher{
		pushFunc: func() error {
			if failCounter == 10 {
				return nil
			}
			failCounter++
			return errors.New("mock error")
		},
		pullFunc: func() error {
			return nil
		},
	}
	err := rebasePushLoop(shell, 50, time.Nanosecond, randomBackoff(100))

	if err != nil {
		t.Fatal(err)
	}
}
