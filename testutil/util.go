package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

var (
	ValidJsonnetBody     = []byte("{}\n")
	InvalidJsonnetBody   = []byte("{,}\n")
	MalformedJsonnetBody = []byte(`function(name, namespace)
  {
    apiVersion: 'accurate.cybozu.com/v1',
    kind: 'SubNamespace',
    metadata: {
      name: name,
      namespace: namespace  // missing the trailing comma
    },
  }
`)
)

func PrepareTestDir(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatal(err)
	}

	return tempDir, func() {
		os.RemoveAll(tempDir)
	}
}

func CreateJsonnet(t *testing.T, dir, fname string, body []byte) string {
	file := filepath.Join(dir, fname)
	if err := os.WriteFile(file, body, 0666); err != nil {
		t.Fatal(err)
	}

	return file
}
