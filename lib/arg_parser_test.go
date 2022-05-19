package lib

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/exp/slices"

	"github.com/cybozu-private/pre-commit-jsonnet/testutil"
)

const (
	existFile = "arg_parser.go"
)

func TestInferOpt(t *testing.T) {
	params := []struct {
		arg  string
		want bool
	}{
		{
			arg:  "-i",
			want: true,
		},
		{
			arg:  "-V",
			want: true,
		},
		{
			arg:  "--jpath",
			want: true,
		},
		{
			arg:  "i",
			want: false,
		},
		{
			arg:  "jpath",
			want: false,
		},
	}

	for _, param := range params {
		got := inferOpt(param.arg)

		if got != param.want {
			t.Errorf("arg='%q', want=%v, got=%v", param.arg, param.want, got)
		}
	}
}

func TestParseArgs(t *testing.T) {
	tempDir, teardown := testutil.PrepareTestDir(t)
	defer teardown()

	libDir := filepath.Join(tempDir, "lib")
	err := os.Mkdir(libDir, 0755)
	if err != nil && !os.IsExist(err) {
		t.Fatal(err)
	}

	jsonnetFiles := []string{
		testutil.CreateJsonnet(t, tempDir, "fpath1", testutil.ValidJsonnetBody),
		testutil.CreateJsonnet(t, tempDir, "fpath2", testutil.ValidJsonnetBody),
	}

	params := []struct {
		args      []string
		wantOpts  []string
		wantFiles []string
	}{
		{
			args:      jsonnetFiles,
			wantOpts:  []string{},
			wantFiles: jsonnetFiles,
		},
		{
			args:      append([]string{"--jpath", libDir}, jsonnetFiles...),
			wantOpts:  []string{"--jpath", libDir},
			wantFiles: jsonnetFiles,
		},
		{
			args:      append([]string{" -a ", " hoge "}, jsonnetFiles...),
			wantOpts:  []string{"-a", "hoge"},
			wantFiles: jsonnetFiles,
		},
		{
			args:      append([]string{"-i", existFile}, jsonnetFiles...),
			wantOpts:  []string{"-i"},
			wantFiles: append([]string{existFile}, jsonnetFiles...),
		},
		{
			args:      append([]string{"--test", "--string-style", "l"}, jsonnetFiles...),
			wantOpts:  []string{"--test", "--string-style", "l"},
			wantFiles: jsonnetFiles,
		},
	}

	for _, param := range params {
		opts, files := ParseArgs(param.args)

		if slices.Compare(opts, param.wantOpts) != 0 {
			t.Errorf("args='%q', wantOpts=%q, gotOpts=%q", param.args, param.wantOpts, opts)
		}
		if slices.Compare(files, param.wantFiles) != 0 {
			t.Errorf("args='%q', wantFiles=%q, gotFiles=%q", param.args, param.wantFiles, files)
		}
	}
}
