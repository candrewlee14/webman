//go:build windows
// +build windows

package link

// If symlink creation fails, fallbacks to `.bat` script.
func symlink(old string, new string) (bool, error) {
	err := os.Symlink(old, new)
	if err == nil {
		return true, nil
	}

	f, err := os.Create(new + ".bat")
	if err != nil {
		return false, err
	}
	defer f.Close()
	_, err = f.WriteString(
		fmt.Sprintf("@echo off\n%s", old) + ` %*`,
	)
	if err != nil {
		return false, err
	}

}
