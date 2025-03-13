package main

import (
	"errors"
	"os"
	"testing"

	"path/filepath"

	"github.com/cybozu/pre-commit-jsonnet/testutil"
)

func TestExecJsonnetLint(t *testing.T) {
	tempDir, teardown := testutil.PrepareTestDir(t)
	defer teardown()

	libDirPath := filepath.Join(tempDir, "lib")
	err := os.Mkdir(libDirPath, 0755)
	if err != nil && !os.IsExist(err) {
		t.Fatal(err)
	}

	jsonnetFilePath := testutil.CreateJsonnet(t, tempDir, "valid.jsonnet", testutil.ValidJsonnetBody)
	invalidJsonnetFilePath := testutil.CreateJsonnet(t, tempDir, "invalid.jsonnet", testutil.InvalidJsonnetBody)

	params := []struct {
		f       string
		opts    []string
		wantErr error
	}{
		{
			f:       jsonnetFilePath,
			opts:    nil,
			wantErr: nil,
		},
		{
			f:       jsonnetFilePath,
			opts:    []string{"--jpath", libDirPath},
			wantErr: nil,
		},
		{
			f:       invalidJsonnetFilePath,
			opts:    nil,
			wantErr: &LintError{},
		},
	}

	for _, param := range params {
		err := execJsonnetLint(param.f, param.opts)

		if err == param.wantErr {
			continue
		}

		var lintErr *LintError
		if !errors.As(err, &lintErr) {
			t.Errorf("args='%q %q', want=%q, got=%q", param.f, param.opts, param.wantErr, err)
		}
	}
}
