package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"io/ioutil"

	"golang.org/x/sys/execabs"

	"github.com/cybozu-private/pre-commit-jsonnet/lib"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const JSONNET_FMT_CMD = "jsonnetfmt"

type FmtError struct {
	args     []string
	diff     string
	exitCode int
	stderr   string
}

func (e *FmtError) Error() string {
	if e.stderr != "" {
		return fmt.Sprintf("exit-code=%d, args=%v\n%s", e.exitCode, e.args, e.stderr)
	}

	return fmt.Sprintf("exit-code=%d, args=%v\n%s\n", e.exitCode, e.args, e.diff)
}

func hasTestOpt(opts []string) bool {
	for _, v := range opts {
		if v == "--test" {
			return true
		}
	}

	return false
}

func diffJsonnetFmt(f string) (string, error) {
	var stdout, stderr bytes.Buffer

	cmd := execabs.Command(JSONNET_FMT_CMD, f)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		fmtError := &FmtError{
			exitCode: cmd.ProcessState.ExitCode(),
			stderr:   stderr.String(),
		}

		return "", fmtError
	}

	b, err := ioutil.ReadFile(f)
	if err != nil {
		return "", err
	}
	before := string(b)
	after := stdout.String()

	if before == after {
		return "", nil
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(after, before, false)

	return dmp.DiffPrettyText(diffs), err
}

func execJsonnetFmt(f string, opts []string) error {
	var stdout, stderr bytes.Buffer

	isTest := hasTestOpt(opts)

	args := []string{}
	args = append(args, opts...)
	args = append(args, "--")
	args = append(args, f)

	cmd := execabs.Command(JSONNET_FMT_CMD, args...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err == nil {
		return nil
	}

	fmtError := &FmtError{
		args:     args,
		exitCode: cmd.ProcessState.ExitCode(),
		stderr:   stderr.String(),
	}

	if isTest {
		diff, diffErr := diffJsonnetFmt(f)
		if diffErr != nil {
			return fmtError
		}
		fmtError.diff = diff
	}

	return fmtError
}

func main() {
	opts, files := lib.ParseArgs(os.Args[1:])
	errExit := false

	if _, err := execabs.LookPath(JSONNET_FMT_CMD); err != nil {
		log.Fatalln(err)
	}

	for _, f := range files {
		err := execJsonnetFmt(f, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] malformed jsonnet found: %s: %s\n", f, err)
			errExit = true
		}
	}

	if errExit {
		os.Exit(1)
	}
}
