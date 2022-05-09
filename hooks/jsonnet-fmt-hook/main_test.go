package main

import (
	"errors"
	"testing"

	"github.com/cybozu-private/pre-commit-jsonnet/testutil"
)

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
			//t.Errorf("param=%q, wantDiff=%v, %s", paam, err)
		}

		if param.wantDiff && len(diff) > 0 {
			continue
		} else if !param.wantDiff && len(diff) == 0 {
			continue
		}

		t.Errorf("%v\n", diff)
	}
}

func TestExecJsonnetFmt(t *testing.T) {
	tempDir, teardown := testutil.PrepareTestDir(t)
	defer teardown()

	jsonnetFile := testutil.CreateJsonnet(t, tempDir, "valid.jsonnet", testutil.ValidJsonnetBody)
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
			f:       testutil.CreateJsonnet(t, tempDir, "malformed.jsonnet", testutil.MalformedJsonnetBody),
			opts:    []string{"--test"},
			wantErr: &FmtError{},
		},
		{
			f:       testutil.CreateJsonnet(t, tempDir, "in-place.jsonnet", testutil.MalformedJsonnetBody),
			opts:    []string{"-i"},
			wantErr: nil,
		},
		{
			f:       invalidJsonnetFile,
			opts:    []string{"--test"},
			wantErr: &FmtError{},
		},
	}

	for _, param := range params {
		err := execJsonnetFmt(param.f, param.opts)

		if param.wantErr == nil {
			if err == nil {
				continue
			}

			t.Errorf("args='%q %q', want=%v, got=%v", param.f, param.opts, param.wantErr, err)
		}

		if !errors.As(err, &param.wantErr) {
			t.Errorf("args='%q %q', want=%v, got=%v", param.f, param.opts, param.wantErr, err)
		}
	}
}
