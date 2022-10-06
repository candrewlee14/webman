package add

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/candrewlee14/webman/config"
	"github.com/candrewlee14/webman/utils"

	"github.com/matryer/is"
)

func TestGroupAdd(t *testing.T) {
	if _, ok := os.LookupEnv("WEBMAN_INTEGRATION"); !ok {
		t.Skip("skipping integration test")
	}

	assert := is.New(t)

	tmp := t.TempDir()
	utils.Init(tmp)
	cfg, err := config.Load()
	assert.NoErr(err) // Should load config
	err = cfg.PkgRepos[0].RefreshRecipes()
	assert.NoErr(err) // Should refresh recipes

	fi, err := os.Create(filepath.Join(utils.WebmanRecipeDir, "webman", "groups", "test.webman-group.yml"))
	assert.NoErr(err) // Should create test file
	_, err = fi.Write(testGroupYAML)
	assert.NoErr(err)        // Should write test file
	assert.NoErr(fi.Close()) // Should close test file

	os.Args = []string{"webman", "test", "--all"}

	err = AddCmd.Execute()
	assert.NoErr(err) // Command should execute

	_, err = os.Stat(filepath.Join(utils.WebmanBinDir, "jq"))
	assert.NoErr(err) // jq binary should exist

	_, err = os.Stat(filepath.Join(utils.WebmanBinDir, "rg"))
	assert.NoErr(err) // rg binary should exist

	_, err = os.Stat(filepath.Join(utils.WebmanBinDir, "bat"))
	assert.True(errors.Is(err, fs.ErrNotExist)) // bat binary should not exist
}

var testGroupYAML = []byte(`packages:
    - jq
    - rg
`)
