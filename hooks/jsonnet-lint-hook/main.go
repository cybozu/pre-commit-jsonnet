package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"golang.org/x/sys/execabs"

	"github.com/cybozu-private/pre-commit-jsonnet/lib"
)

const JSONNET_LINT_CMD = "jsonnet-lint"

type LintError struct {
	args   []string
	stderr string
}

func (e *LintError) Error() string {
	return fmt.Sprintf("args=%v\n%s", e.args, e.stderr)
}

func execJsonnetLint(f string, opts []string) error {
	var args []string
	var stderr bytes.Buffer

	args = append(args, opts...)
	args = append(args, "--")
	args = append(args, f)
	cmd := execabs.Command(JSONNET_LINT_CMD, args...)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return &LintError{args: args, stderr: stderr.String()}
	}

	return nil
}

func main() {
	opts, files := lib.ParseArgs(os.Args[1:])
	errExit := false

	if _, err := execabs.LookPath(JSONNET_LINT_CMD); err != nil {
		log.Fatalln(err)
	}

	for _, f := range files {
		err := execJsonnetLint(f, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] lint failed: %s\n", err)
			errExit = true
		}
	}

	if errExit {
		os.Exit(1)
	}
}
