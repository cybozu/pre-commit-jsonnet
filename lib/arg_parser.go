package lib

import (
	"errors"
	"os"
	"regexp"
	"strings"
)

var reOpt = regexp.MustCompile("^--[a-z][-a-z]*|^-[a-zA-Z]$")

func inferOpt(arg string) bool {
	return reOpt.MatchString(arg)
}

// ParseArgs split arguments into options and file paths
func ParseArgs(args []string) (opts, files []string) {
	foundOpt := false

	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		fi, err := os.Stat(arg)
		fileNotFound := errors.Is(err, os.ErrNotExist)

		if foundOpt && !inferOpt(arg) && (fileNotFound || fi.IsDir()) {
			opts = append(opts, arg)
			foundOpt = false
			continue
		}

		foundOpt = false

		if inferOpt(arg) && fileNotFound {
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
