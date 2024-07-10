package main

import (
	"log/slog"
	"os"
	"path"
)

func UpdateGBSW(gbPath string, wc *WriteCounter) error {
	if _, err := os.Stat(path.Join(gbPath, ".gbwrapper_updated")); err == nil {

	}

	done := make(chan bool)
	count, _ := GetGBSw(path.Join(gbPath, "gameband_sw.zip"), done)
	*wc = *count
	<-done
	state = 4
	slog.Info("Gameband software downloaded, extracting")
	f, err := os.Open(path.Join(gbPath, "gameband_sw.zip"))
	if err != nil {
		errText = err.Error()
		state = -1
		return err
	}
	s, _ := f.Stat()
	ExtractZip(gbPath, f, s.Size())
	BlugeonGatekeeper(path.Join(gbPath, ".lib"))
	return nil
}
