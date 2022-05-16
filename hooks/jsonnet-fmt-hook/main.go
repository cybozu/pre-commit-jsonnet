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

const jsonnetFmtCmd = "jsonnetfmt"
const (
	ansiReset = "\x1b[0m"
	ansiRed   = "\x1b[31m"
	ansiGreen = "\x1b[32m"
)

type CmdResult struct {
	err error
}

type FileDiff struct {
	text      string
	numInsert int
	numDelete int
}

type FmtError struct {
	path     string
	args     []string
	diff     *FileDiff
	exitCode int
	stderr   string
}

func (e *FmtError) Error() string {
	stderr := strings.TrimSpace(e.stderr)

	if stderr != "" || e.diff == nil {
		return fmt.Sprintf("path=%s exit-code=%d\n%s", e.path, e.exitCode, stderr)
	}

	insert := fmt.Sprintf("%s+%d%s", ansiGreen, e.diff.numInsert, ansiReset)
	delete := fmt.Sprintf("%s-%d%s", ansiRed, e.diff.numDelete, ansiReset)

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

func summarizeDiff(diffs []diffmatchpatch.Diff) *FileDiff {
	var builder strings.Builder
	var err error
	var endWithLineBreak bool
	var fileDiff FileDiff

	writeString := func(builder *strings.Builder, s string) {
		_, err = builder.WriteString(s)
		if err != nil {
			log.Fatalln(err)
		}
	}
	splitLines := func(text string) []string {
		return strings.Split(strings.TrimSpace(strings.ReplaceAll(text, "\r\n", "\n")), "\n")
	}

	const maxShowLine = 3
	const omissionText = "...\n"
	lastDiffIdx := len(diffs) - 1

	for i, diff := range diffs {
		text := diff.Text
		lines := splitLines(text)

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			writeString(&builder, ansiGreen+text+ansiReset)
			fileDiff.numInsert += len(text)
		case diffmatchpatch.DiffDelete:
			writeString(&builder, ansiRed+text+ansiReset)
			fileDiff.numDelete += len(text)
		case diffmatchpatch.DiffEqual:
			var showHead, showTail bool
			endWithLineBreak = text[len(text)-1:] == "\n"
			writeLineCount := min(maxShowLine, len(lines))

			if i == 0 {
				showTail = true
			} else if i == lastDiffIdx {
				showHead = true
			} else {
				showTail = true
				showHead = true
			}

			if showHead && showTail && len(lines) <= writeLineCount*2 {
				writeString(&builder, text)
				continue
			}

			if (showHead || showTail) && len(lines) <= writeLineCount {
				writeString(&builder, text)
				continue
			}

			if showHead {
				partText := strings.Join(lines[:writeLineCount], "\n")
				writeString(&builder, partText)
				if partText[len(partText)-1:] != "\n" {
					writeString(&builder, "\n")
				}
				writeString(&builder, omissionText)
			}
			if showTail {
				if !showHead && !endWithLineBreak {
					writeString(&builder, omissionText)
				}
				writeString(&builder, strings.Join(lines[len(lines)-writeLineCount:], "\n"))
			}
		}
	}

	fileDiff.text = builder.String()
	return &fileDiff
}

func diffJsonnetFmt(f string) (*FileDiff, error) {
	var stdout, stderr bytes.Buffer

	cmd := execabs.Command(jsonnetFmtCmd, f)
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

	return summarizeDiff(diffs), err
}

func execJsonnetFmt(f string, opts []string) error {
	var stdout, stderr bytes.Buffer

	args := make([]string, 0, len(opts)+2)
	args = append(args, opts...)
	args = append(args, "--")
	args = append(args, f)

	cmd := execabs.Command(jsonnetFmtCmd, args...)
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
		diff, err := diffJsonnetFmt(f)
		if err != nil {
			return fmtError
		}
		fmtError.diff = diff
	}

	return fmtError
}

func main() {
	opts, files := lib.ParseArgs(os.Args[1:])

	if _, err := execabs.LookPath(jsonnetFmtCmd); err != nil {
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
