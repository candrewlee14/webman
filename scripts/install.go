package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mholt/archiver/v3"
)

const latestURL = "https://api.github.com/repos/candrewlee14/webman/releases/latest"

func main() {
	var ext string
	switch runtime.GOOS {
	case "darwin", "linux":
		ext = "tar.gz"
	case "windows":
		ext = "zip"
	default:
		fmt.Println("this ARCH isn't supported by this script, please build from source")
		return
	}

	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch64"
	default:
		fmt.Println("this OS isn't supported by this script, please build from source")
		return
	}

	resp, err := http.Get(latestURL)
	if err != nil {
		fmt.Printf("could not get latest release: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Printf("could not decode latest release: %v\n", err)
		return
	}

	tmp, err := os.MkdirTemp(os.TempDir(), "webman-install")
	if err != nil {
		fmt.Printf("could not create temp dir: %v\n", err)
		return
	}
	defer func() {
		if err := os.RemoveAll(tmp); err != nil {
			fmt.Printf("could not remove temp dir: %v\n", err)
		}
	}()

	downloadURL := fmt.Sprintf("https://github.com/candrewlee14/webman/releases/download/%s/webman_%s_%s_%s.%s", release.Tag, strings.TrimPrefix(release.Tag, "v"), runtime.GOOS, arch, ext)
	dl, err := http.Get(downloadURL)
	if err != nil {
		fmt.Printf("could not download latest release: %v\n", err)
		return
	}
	defer dl.Body.Close()

	tmpArchivePath := filepath.Join(tmp, fmt.Sprintf("webman.%s", ext))
	tmpArchive, err := os.Create(tmpArchivePath)
	if err != nil {
		fmt.Printf("could not create temp archive: %v\n", err)
		return
	}
	if _, err := io.Copy(tmpArchive, dl.Body); err != nil {
		fmt.Printf("could not copy download into temp archive: %v\n", err)
		return
	}

	if err := archiver.Unarchive(tmpArchivePath, tmp); err != nil {
		fmt.Printf("could not unarchive latest release: %v\n", err)
		return
	}

	var binExt string
	if runtime.GOOS == "windows" {
		binExt = ".exe"
	}
	binPath := filepath.Join(tmp, "webman"+binExt)

	cmd := exec.Command(binPath, "add", "webman", "--switch")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("could not run webman: %v\n", err)
		return
	}
}

type GitHubRelease struct {
	Tag string `json:"tag_name"`
}
