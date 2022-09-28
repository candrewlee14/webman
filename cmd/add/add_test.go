package add

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/candrewlee14/webman/utils"

	"github.com/matryer/is"
)

func TestAdd(t *testing.T) {
	if _, ok := os.LookupEnv("WEBMAN_INTEGRATION"); !ok {
		t.Skip("skipping integration test")
	}

	assert := is.New(t)

	tmp := t.TempDir()
	utils.Init(tmp)
	os.Args = []string{"webman", "jq", "rg"}

	err := AddCmd.Execute()
	assert.NoErr(err) // Command should execute

	_, err = os.Stat(filepath.Join(utils.WebmanBinDir, "jq"))
	assert.NoErr(err) // jq binary should exist

	_, err = os.Stat(filepath.Join(utils.WebmanBinDir, "rg"))
	assert.NoErr(err) // rg binary should exist

	_, err = os.Stat(filepath.Join(utils.WebmanBinDir, "bat"))
	assert.True(errors.Is(err, fs.ErrNotExist)) // bat binary should not exist
}
