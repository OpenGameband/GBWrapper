package main

import (
	"fmt"
	"gioui.org/io/system"
	"gioui.org/unit"
	"image/color"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"
)

var jreDirRegex = regexp.MustCompile("jdk8u([0-9]+)-[a-z]([0-9]+)-jre")
var state = 1
var errText = ""
var window *app.Window

func main() {

	exe, err := os.Executable()
	if err != nil {
		slog.Error("Failed to get executable path", "err", err)
	}
	gbPath := path.Dir(exe)

	if d, err := os.Stat(path.Join(gbPath, ".gbJava", runtime.GOOS+"-"+runtime.GOARCH)); err == nil {
		if d.IsDir() {
			slog.Info("Java already installed, launching Gameband Launcher")
			var cmd *exec.Cmd
			switch runtime.GOOS {
			case "windows":
				cmd = exec.Command(path.Join(gbPath, ".gbJava", runtime.GOOS+"-"+runtime.GOARCH, "bin", "java.exe"), "-jar", path.Join(gbPath, ".lib/gblauncher.jar"))
			case "darwin":
				cmd = exec.Command(path.Join(gbPath, ".gbJava", runtime.GOOS+"-"+runtime.GOARCH, "Contents", "Home", "bin", "java"), "-jar", path.Join(gbPath, ".lib/gblauncher.jar"))
			case "linux":
				cmd = exec.Command(path.Join(gbPath, ".gbJava", runtime.GOOS+"-"+runtime.GOARCH, "bin", "java"), "-jar", path.Join(gbPath, ".lib/gblauncher.jar"))
			}
			cmd.Dir = gbPath
			err = cmd.Run()
			if err != nil {
				panic(err)
			}
			return
		}
	}

	done := make(chan bool)
	count, err := GetJava(path.Join(gbPath, "java.zip"), done)

	go func() {
		<-done
		state = 2
		window.Invalidate()
		slog.Info("Java downloaded, extracting")
		f, err := os.Open(path.Join(gbPath, "java.zip"))
		if err != nil {
			errText = err.Error()
			state = -1
			window.Invalidate()
			return
		}
		s, _ := f.Stat()

		if runtime.GOOS == "windows" {
			ExtractZip(path.Join(gbPath, ".gbJava", runtime.GOOS+"-"+runtime.GOARCH), f, s.Size())
		} else {
			ExtractTarGz(path.Join(gbPath, ".gbJava", runtime.GOOS+"-"+runtime.GOARCH), f)
		}

		err = UpdateGBSW(gbPath, count)
		if err != nil {
			return
		}

		os.Remove(path.Join(gbPath, "java.zip"))
		os.Remove(path.Join(gbPath, "gameband_sw.zip"))

		state = 0
		window.Perform(system.ActionClose)
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command(path.Join(gbPath, ".gbJava", runtime.GOOS+"-"+runtime.GOARCH, "bin", "java.exe"), "-jar", path.Join(gbPath, ".lib/gblauncher.jar"))
		case "darwin":
			cmd = exec.Command(path.Join(gbPath, ".gbJava", runtime.GOOS+"-"+runtime.GOARCH, "Contents", "Home", "bin", "java"), "-jar", path.Join(gbPath, ".lib/gblauncher.jar"))
		case "linux":
			cmd = exec.Command(path.Join(gbPath, ".gbJava", runtime.GOOS+"-"+runtime.GOARCH, "bin", "java"), "-jar", path.Join(gbPath, ".lib/gblauncher.jar"))
		}
		err = cmd.Run()
		if err != nil {
			panic(err)
		}

	}()

	app.Title("Gameband Updater")
	go func() {
		window = new(app.Window)
		err := run(window, count)
		if err != nil {
			log.Fatal(err)
		}
		if state == -1 {
			os.Exit(0)
		}

	}()
	app.Main()

}

func run(window *app.Window, counter *WriteCounter) error {
	theme := material.NewTheme()
	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			// This graphics context is used for managing the rendering state.
			gtx := app.NewContext(&ops, e)
			if state == 1 || state == 3 {
				if counter == nil {
					continue
				}
				// Define an large label with an appropriate text:
				var titleText string
				if state == 1 {
					titleText = "Downloading Java..."
				} else {
					titleText = "Downloading Gameband Software..."
				}
				title := material.H1(theme, titleText)

				// Change the color of the label.
				maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
				title.Color = maroon

				// Change the position of the label.
				title.Alignment = text.Middle

				// Draw the label to the graphics context.
				title.Layout(gtx)

				label := material.Label(theme, unit.Sp(10), fmt.Sprintf("%d / %d bytes", counter.Written, counter.Total))
				label.Layout(gtx)

				progress := material.ProgressBar(theme, counter.Percent())
				progress.Layout(gtx)
			} else if state == 2 || state == 4 {
				var titleText string
				if state == 1 {
					titleText = "Extracting Java"
				} else {
					titleText = "Extracting Gameband Software"
				}
				title := material.H1(theme, titleText)
				// Change the position of the label.
				title.Alignment = text.Middle

				// Change the color of the label.
				maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
				title.Color = maroon

				title.Layout(gtx)

				loader := material.Loader(theme)
				loader.Layout(gtx)
			} else if state == 0 {
				return nil
			} else if state == -1 {
				title := material.H1(theme, "Error")
				// Change the position of the label.
				title.Alignment = text.Middle

				// Change the color of the label.
				maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
				title.Color = maroon

				title.Layout(gtx)
				material.Body1(theme, errText).Layout(gtx)
			}

			// Pass the drawing operations to the GPU.
			e.Frame(gtx.Ops)
		}
	}
}
