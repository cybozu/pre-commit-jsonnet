package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/cybozu/pre-commit-jsonnet/testutil"
)

var reAnsiEscape = regexp.MustCompile(`(?:\x1B[@-Z\\-_]|[\x80-\x9A\x9C-\x9F]|(?:\x1B\[|\x9B)[0-?]*[ -/]*[@-~])`)

func normalizeDiff(diff string) string {
	reSpaces := regexp.MustCompile(`\s+`)

	normDiff := strings.TrimSpace(diff)
	normDiff = reSpaces.ReplaceAllString(normDiff, " ")
	normDiff = reAnsiEscape.ReplaceAllString(normDiff, "")
	return strings.ToLower(normDiff)
}

func TestHasTestOpt(t *testing.T) {
	params := []struct {
		opts []string
		want bool
	}{
		{
			opts: []string{"--test", "hoge"},
			want: true,
		},
		{
			opts: []string{},
			want: false,
		},
		{
			opts: []string{"test", "hoge"},
			want: false,
		},
	}

	for _, param := range params {
		got := hasTestOpt(param.opts)
		if got != param.want {
			t.Errorf("args='%q', want=%v, got=%v", param.opts, param.want, got)
		}
	}
}

func TestDiffJsonnetFmt(t *testing.T) {
	tempDir, teardown := testutil.PrepareTestDir(t)
	defer teardown()

	jsonnetFile := testutil.CreateJsonnet(t, tempDir, "valid.jsonnet", testutil.ValidJsonnetBody)
	malformedJsonnetFile := testutil.CreateJsonnet(t, tempDir, "malformed.jsonnet", testutil.MalformedJsonnetBody)

	params := []struct {
		f        string
		wantDiff bool
	}{
		{
			f:        malformedJsonnetFile,
			wantDiff: true,
		},
		{
			f:        jsonnetFile,
			wantDiff: false,
		},
	}

	for _, param := range params {
		diff, err := diffJsonnetFmt(param.f)
		if err != nil {
			t.Errorf("param=%v, err=%s\n", param, err)
			continue
		}

		if param.wantDiff && diff != nil && len(diff.text) > 0 {
			continue
		} else if !param.wantDiff && diff == nil {
			continue
		}

		t.Errorf("%v\n", diff)
	}
}

func TestExecJsonnetFmt(t *testing.T) {
	tempDir, teardown := testutil.PrepareTestDir(t)
	defer teardown()

	jsonnetFile := testutil.CreateJsonnet(t, tempDir, "valid.jsonnet", testutil.ValidJsonnetBody)
	malformedjsonnetFile := testutil.CreateJsonnet(t, tempDir, "malformed.jsonnet", testutil.MalformedJsonnetBody)
	invalidJsonnetFile := testutil.CreateJsonnet(t, tempDir, "invalid.jsonnet", testutil.InvalidJsonnetBody)

	params := []struct {
		f       string
		opts    []string
		wantErr error
	}{
		{
			f:       jsonnetFile,
			opts:    nil,
			wantErr: nil,
		},
		{
			f:    malformedjsonnetFile,
			opts: []string{"--test"},
			wantErr: &FmtError{
				exitCode: 2,
				diff: &FileDiff{
					text: `...
						metadata: {
							name: name,
							namespace: namespace,  // missing the trailing comma
						},
					}`,
					numInsert: 1,
					numDelete: 0,
				},
			},
		},
		{
			f:       testutil.CreateJsonnet(t, tempDir, "in-place.jsonnet", testutil.MalformedJsonnetBody),
			opts:    []string{"-i"},
			wantErr: nil,
		},
		{
			f:       invalidJsonnetFile,
			opts:    []string{"--test"},
			wantErr: &FmtError{exitCode: 1, diff: nil},
		},
	}

	for _, param := range params {
		err := execJsonnetFmt(param.f, param.opts)
		baseInfo := fmt.Sprintf("params=(%q, %q)", param.f, param.opts)

		if param.wantErr == nil {
			if err == nil {
				continue
			}

			t.Errorf("%s, want=%v, got=%v", baseInfo, param.wantErr, err)
			continue
		}

		var want, got *FmtError

		if !errors.As(param.wantErr, &want) {
			t.Errorf("unexpected error: %s, want=%v", baseInfo, param.wantErr)
			continue
		}
		if !errors.As(err, &got) {
			t.Errorf("unexpected error: %s, got=%v", baseInfo, err)
		}
		if got.exitCode != want.exitCode {
			t.Errorf("%s, want=%v, got=%v", baseInfo, want.exitCode, got.exitCode)
		}
		if got.diff == nil && want.diff == nil {
			continue
		}

		gotDiff := normalizeDiff(got.diff.text)
		wantDiff := normalizeDiff(want.diff.text)
		if gotDiff != wantDiff {
			t.Errorf("%s, want=%v, got=%v", baseInfo, wantDiff, gotDiff)
		}
	}
}
