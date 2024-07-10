package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
)

func ExtractZip(gbPath string, f io.ReaderAt, size int64) {
	r, err := zip.NewReader(f, size)
	if err != nil {
		errText = err.Error()
		state = -1
		return
	}
	for _, f := range r.File {
		adjustedFile := jreDirRegex.ReplaceAllString(f.Name, "")
		if adjustedFile == "" {
			continue
		}
		f.FileInfo().Mode().
		fname := path.Base(f.FileInfo().Name())
		if fname == "Gameband.app" || fname == "Gameband.exe" || fname == "Gameband.bat" {
			continue
		}
		if f.FileInfo().IsDir() {
			fmt.Println("Creating directory: ", path.Join(gbPath, adjustedFile))
			os.MkdirAll(path.Join(gbPath, adjustedFile), 0755)
			continue
		}

		fmt.Println("Extracting file: ", path.Join(gbPath, adjustedFile))
		rc, err := f.Open()
		if err != nil {
			errText = err.Error()
			state = -1
			return
		}
		outFile, err := os.OpenFile(path.Join(gbPath, adjustedFile), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			errText = err.Error()
			state = -1
			return
		}
		io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		fmt.Println("Done")
	}
}

func ExtractTarGz(gbPath string, f io.Reader) {
	gz, err := gzip.NewReader(f)
	if err != nil {
		errText = err.Error()
		state = -1
		return
	}
	r := tar.NewReader(gz)
	for true {
		header, err := r.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}
		adjustedFile := path.Join(gbPath, jreDirRegex.ReplaceAllString(header.Name, ""))

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(adjustedFile, 0755); err != nil {
				log.Fatalf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.OpenFile(adjustedFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				log.Fatalf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, r); err != nil {
				log.Fatalf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
			outFile.Close()

		default:
			log.Fatalf(
				"ExtractTarGz: uknown type: %s in %s",
				header.Typeflag,
				header.Name)
		}

	}
}
