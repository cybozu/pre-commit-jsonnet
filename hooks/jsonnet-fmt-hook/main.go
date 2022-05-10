package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"io/ioutil"

	"golang.org/x/sys/execabs"

	"github.com/cybozu-private/pre-commit-jsonnet/lib"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const JSONNET_FMT_CMD = "jsonnetfmt"
const (
	ANSI_RESET = "\x1b[0m"
	ANSI_RED   = "\x1b[31m"
	ANSI_GREEN = "\x1b[32m"
)

type CmdResult struct {
	err error
}

type Diff struct {
	text      string
	numInsert int
	numDelete int
}

type FmtError struct {
	path     string
	args     []string
	diff     *Diff
	exitCode int
	stderr   string
}

func (e *FmtError) Error() string {
	stderr := strings.TrimSpace(e.stderr)

	if stderr != "" || e.diff == nil {
		return fmt.Sprintf("path=%s exit-code=%d\n%s", e.path, e.exitCode, stderr)
	}

	insert := fmt.Sprintf("%s+%d%s", ANSI_GREEN, e.diff.numInsert, ANSI_RESET)
	delete := fmt.Sprintf("%s-%d%s", ANSI_RED, e.diff.numDelete, ANSI_RESET)

	return fmt.Sprintf("path=%s (%s %s), exit-code=%d\n%s\n", e.path, insert, delete, e.exitCode, strings.TrimSpace(e.diff.text))
}

func hasTestOpt(opts []string) bool {
	for _, v := range opts {
		if v == "--test" {
			return true
		}
	}

	return false
}

func toSummary(diffs []diffmatchpatch.Diff) *Diff {
	var builder strings.Builder
	var err error
	var lastLineBreak bool
	var fileDiff Diff

	writeString := func(builder *strings.Builder, s string) {
		_, err = builder.WriteString(s)
		if err != nil {
			log.Fatalln(err)
		}
	}
	min := func(x, y int) int {
		if x < y {
			return x
		}
		return y
	}
	splitLines := func(text string) []string {
		return strings.Split(strings.TrimSpace(strings.ReplaceAll(text, "\r\n", "\n")), "\n")
	}

	for i, diff := range diffs {
		text := diff.Text
		lines := splitLines(text)

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			writeString(&builder, ANSI_GREEN+text+ANSI_RESET)
			fileDiff.numInsert += len(text)
		case diffmatchpatch.DiffDelete:
			writeString(&builder, ANSI_RED+text+ANSI_RESET)
			fileDiff.numDelete += len(text)
		case diffmatchpatch.DiffEqual:
			var showHead, showTail bool
			lastLineBreak = text[len(text)-1:] == "\n"
			partLines := min(3, len(lines))

			if i == 0 {
				showTail = true
			} else if i == (len(diffs) - 1) {
				showHead = true
			} else {
				showTail = true
				showHead = true
			}

			if showHead && showTail && len(lines) <= partLines*2 {
				writeString(&builder, text)
				continue
			}

			if (showHead || showTail) && len(lines) <= partLines {
				writeString(&builder, text)
				continue
			}

			if showHead {
				partText := strings.Join(lines[:partLines], "\n")
				writeString(&builder, partText)
				if partText[len(partText)-1:] != "\n" {
					writeString(&builder, "\n")
				}
				writeString(&builder, "..."+"\n")
			}
			if showTail {
				if !showHead && !lastLineBreak {
					writeString(&builder, "..."+"\n")
				}
				writeString(&builder, strings.Join(lines[len(lines)-partLines:], "\n"))
			}
		}
	}

	fileDiff.text = builder.String()
	return &fileDiff
}

func diffJsonnetFmt(f string) (*Diff, error) {
	var stdout, stderr bytes.Buffer

	cmd := execabs.Command(JSONNET_FMT_CMD, f)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		fmtError := &FmtError{
			path:     f,
			exitCode: cmd.ProcessState.ExitCode(),
			stderr:   stderr.String(),
		}

		return nil, fmtError
	}

	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	before := string(b)
	after := stdout.String()

	if before == after {
		return nil, nil
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(after, before, false)

	return toSummary(diffs), err
}

func execJsonnetFmt(f string, opts []string) error {
	var stdout, stderr bytes.Buffer

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
		path:     f,
		args:     args,
		exitCode: cmd.ProcessState.ExitCode(),
		stderr:   stderr.String(),
	}

	if hasTestOpt(opts) {
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

	if _, err := execabs.LookPath(JSONNET_FMT_CMD); err != nil {
		log.Fatalln(err)
	}

	results := make(chan CmdResult, len(files))

	for _, f := range files {
		go func(f string) {
			err := execJsonnetFmt(f, opts)
			results <- CmdResult{err: err}
		}(f)
	}

	errExit := false
	for i := 0; i < len(files); i++ {
		result := <-results
		if result.err == nil {
			continue
		}

		fmt.Fprintf(os.Stderr, "[ERROR] malformed jsonnet found: %s\n", result.err)
		errExit = true
	}

	if errExit {
		os.Exit(1)
	}
}
