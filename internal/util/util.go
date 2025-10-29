package util

import (
	"os"
	"path/filepath"
)

func Pointer[T any](value T) *T {
	return &value
}

// ExpandTilde expands ~ to home directory
func ExpandTilde(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if len(path) == 1 {
		return home
	}

	if path[1] == '/' || path[1] == filepath.Separator {
		return filepath.Join(home, path[2:])
	}

	return path
}
