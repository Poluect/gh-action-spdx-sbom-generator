package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {

	// get input variables from action
	inputVersion := os.Getenv("INPUT_VERSION")
	inputCommands := os.Getenv("INPUT_COMMANDS")

	// build artifact url
	githubPackageURL := "https://github.com/spdx/spdx-sbom-generator/releases/download/v"
	artifactName := "-linux-386.tar.gz"
	url := fmt.Sprintf("%s%s%s%s%s", githubPackageURL, inputVersion, "/spdx-sbom-generator-v", inputVersion, artifactName)

	fmt.Println("--- VERSION ---", inputVersion)
	fmt.Println("--- COMMAND ---", inputCommands)
	fmt.Println("--- URL ---", url)

	//download artifact file
	err := DownloadFile(url)
	if err != nil {
		panic(err)
	}

	// execute spdx-sbom-generator cli
	cmd := exec.Command("./spdx-sbom-generator", strings.Fields(inputCommands)...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("An error occurred during spdx-sbom-generator operation: %v, stderr: (%s)", err, stderr.String())
		os.Exit(1)
		return
	}

	fmt.Println("--- OUTPUT ---")
	fmt.Printf("%s\n%s\n", out.String(), stderr.String())
}

// DownloadFile will download a url and store it in local current directory
func DownloadFile(url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// return if nothing to return
		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}

		// current target dir where the file should be downloaded
		path, err := os.Getwd()
		if err != nil {
			return err
		}
		target := filepath.Join(path, header.Name)

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()
		}
	}
}
