package main

import (
	"os/exec"
	"runtime"
)

// I really wish macos wouldn't be such a pain

func BlugeonGatekeeper(path string) {
	if runtime.GOOS != "darwin" {
		return
	}
	cmd := exec.Command("/usr/bin/xattr", "-r", "-d", "com.apple.quarantine", path)
	cmd.Dir = path
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
