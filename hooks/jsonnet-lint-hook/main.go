package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/sys/execabs"

	"github.com/cybozu/pre-commit-jsonnet/lib"
)

const jsonnetLintCmd = "jsonnet-lint"

type CmdResult struct {
	err error
}

type LintError struct {
	args   []string
	stderr string
}

func (e *LintError) Error() string {
	return fmt.Sprintf("args=%v\n%s", e.args, strings.TrimSpace(e.stderr))
}

func execJsonnetLint(f string, opts []string) error {
	var stderr bytes.Buffer

	args := make([]string, 0, len(opts)+2)
	args = append(args, opts...)
	args = append(args, []string{"--", f}...)
	cmd := execabs.Command(jsonnetLintCmd, args...)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return &LintError{args: args, stderr: stderr.String()}
	}

	return nil
}

func main() {
	opts, files := lib.ParseArgs(os.Args[1:])

	if _, err := execabs.LookPath(jsonnetLintCmd); err != nil {
		log.Fatalln(err)
	}

	results := make(chan CmdResult, len(files))

	for _, f := range files {
		go func(f string) {

			err := execJsonnetLint(f, opts)
			results <- CmdResult{err: err}
		}(f)
	}

	errExit := false
	for i := 0; i < len(files); i++ {
		result := <-results
		if result.err == nil {
			continue
		}

		fmt.Fprintf(os.Stderr, "[ERROR] lint failed: %s\n", result.err)
		errExit = true
	}

	if errExit {
		os.Exit(1)
	}
}
