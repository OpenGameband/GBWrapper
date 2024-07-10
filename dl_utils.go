package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"
)

type WriteCounter struct {
	Written uint64
	Total   uint64
}

func (wc *WriteCounter) Percent() float32 {
	return float32(wc.Written) / float32(wc.Total)
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Written += uint64(n)
	window.Invalidate()
	return n, nil
}

const urlFormat = "https://api.adoptium.net/v3/binary/latest/8/ga/%s/%s/jre/hotspot/normal/eclipse"

const gbSwURL = "https://files.gameband.valtek.uk/gameband_sw.zip"

func GetGBSw(path string, done chan<- bool) (*WriteCounter, error) {
	slog.Info("Downloading Gameband Software from", "url", gbSwURL)

	resp, err := http.Get(gbSwURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to download Gameband Software", "status", resp.Status)
		return nil, errors.New("Failed to download Gameband Software")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	count := WriteCounter{
		Total: uint64(resp.ContentLength),
	}
	go func() {
		defer f.Close()
		defer resp.Body.Close()
		io.Copy(f, io.TeeReader(resp.Body, &count))
		done <- true
	}()
	return &count, nil
}

func GetJava(path string, done chan<- bool) (*WriteCounter, error) {
	goos := runtime.GOOS
	if goos == "darwin" {
		goos = "mac"
	}

	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		arch = "x64"
	case "386":
		arch = "x86"
	case "arm64":
		if goos != "mac" {
			return nil, errors.New("Gameband does not support this architecture")
		}
		arch = "x64"
	default:
		return nil, errors.New("Gameband does not support this architecture")
	}

	url := fmt.Sprintf(urlFormat, goos, arch)
	slog.Info("Downloading Java from", "url", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to download Java", "status", resp.Status)
		return nil, errors.New("Failed to download Java")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	count := WriteCounter{
		Total: uint64(resp.ContentLength),
	}
	go func() {
		defer f.Close()
		defer resp.Body.Close()
		io.Copy(f, io.TeeReader(resp.Body, &count))
		done <- true
	}()
	return &count, nil
}
