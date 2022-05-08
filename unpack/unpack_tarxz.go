package unpack

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func UntarxzExec(src string, dir string) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("windows doesn't have support for tarballs")
	}
	cmd := exec.Command("tar", "-xf", src, "--directory="+dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
