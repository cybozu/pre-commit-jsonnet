package lib

import (
	"errors"
	"os"
	"strings"
)

func canOpt(arg string) bool {
	return strings.HasPrefix(arg, "-")
}

// ParseArgs split arguments into options and files
func ParseArgs(args []string) (opts, files []string) {
	foundOpt := false

	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		fi, err := os.Stat(arg)
		fileNotFound := errors.Is(err, os.ErrNotExist)

		if foundOpt && (fileNotFound || fi.IsDir()) {
			opts = append(opts, arg)
			foundOpt = false
			continue
		}

		foundOpt = false

		if canOpt(arg) && fileNotFound {
			foundOpt = true
			opts = append(opts, arg)
			continue
		}

		if err == nil && !fi.IsDir() {
			files = append(files, arg)
		}
	}

	return
}
