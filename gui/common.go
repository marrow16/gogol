package gui

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

var modKeyName = func() string {
	if runtime.GOOS == "darwin" {
		return "⌥" //"⌘"
	}
	return "Ctrl+"
}()
var altKeyName = func() string {
	if runtime.GOOS == "darwin" {
		return "⌥" //"⌘"
	}
	return "Alt+"
}()

func commas(s string) string {
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func saveFile(path string, allowOverwrite bool) (result *os.File, err error) {
	if path, err = resolveSavePath(path); err != nil {
		return
	}
	if allowOverwrite {
		if result, err = os.Create(path); err != nil {
			err = errors.New("Unable to create file")
		}
	} else if result, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		if errors.Is(err, fs.ErrExist) {
			err = errors.New("File already exists")
		} else {
			err = errors.New("Unable to create file")
		}
	}
	return result, err
}

func resolvePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("Invalid user home directory")
	}
	dir := filepath.Join(home, "Documents", "GoGoL")
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return "", errors.New("Invalid user home directory")
	}
	return filepath.Join(dir, path), nil
}

func resolveSavePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("Invalid user home directory")
	}
	dir := filepath.Join(home, "Documents", "GoGoL")
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return "", errors.New("Invalid user home directory")
	}
	result := filepath.Join(dir, path)
	dir = filepath.Dir(result)
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return "", errors.New("Invalid path")
	}
	return result, nil
}
