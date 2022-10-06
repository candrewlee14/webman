package remove

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/candrewlee14/webman/cmd/group/add"
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

	err = add.AddCmd.Execute()
	assert.NoErr(err) // Command should execute

	_, err = os.Stat(filepath.Join(utils.WebmanBinDir, "jq"))
	assert.NoErr(err) // jq binary should exist

	_, err = os.Stat(filepath.Join(utils.WebmanBinDir, "rg"))
	assert.NoErr(err) // rg binary should exist

	err = RemoveCmd.Execute()
	assert.NoErr(err) // Command should execute

	_, err = os.Stat(filepath.Join(utils.WebmanBinDir, "jq"))
	assert.True(errors.Is(err, fs.ErrNotExist)) // jq binary should no longer exist

	_, err = os.Stat(filepath.Join(utils.WebmanBinDir, "rg"))
	assert.True(errors.Is(err, fs.ErrNotExist)) // rg binary should no longer exist

	_, err = os.Stat(filepath.Join(utils.WebmanPkgDir, "jq"))
	assert.True(errors.Is(err, fs.ErrNotExist)) // jq pkg should no longer exist

	_, err = os.Stat(filepath.Join(utils.WebmanPkgDir, "rg"))
	assert.True(errors.Is(err, fs.ErrNotExist)) // rg pkg should no longer exist
}

var testGroupYAML = []byte(`packages:
    - jq
    - rg
`)
