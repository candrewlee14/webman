package remove

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/candrewlee14/webman/cmd/add"
	"github.com/candrewlee14/webman/utils"

	"github.com/matryer/is"
)

func TestRemove(t *testing.T) {
	if _, ok := os.LookupEnv("WEBMAN_INTEGRATION"); !ok {
		t.Skip("skipping integration test")
	}

	assert := is.New(t)

	tmp := t.TempDir()
	utils.Init(tmp)
	os.Args = []string{"webman", "jq"}

	err := add.AddCmd.Execute()
	assert.NoErr(err) // Command should execute

	_, err = os.Stat(filepath.Join(utils.WebmanBinDir, "jq"))
	assert.NoErr(err) // jq binary should exist

	err = RemoveCmd.Execute()
	assert.NoErr(err) // Command should execute

	_, err = os.Stat(filepath.Join(utils.WebmanBinDir, "jq"))
	assert.True(err != nil) // jq binary should no longer exist
}
