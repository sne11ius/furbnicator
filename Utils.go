package main

import (
	"os"
	"os/exec"
	"runtime"
)

// From https://stackoverflow.com/a/39324149
func LaunchUrl(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// From https://stackoverflow.com/a/39324149
func CloneUrl(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "git"}
	case "darwin":
		cmd = "git"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "git"
	}
	args = append(args, "clone", url)
	proc := exec.Command(cmd, args...)
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	return proc.Run()
}
