package cmd

import (
	"github.com/candrewlee14/webman/cmd/add"
	"github.com/candrewlee14/webman/cmd/config"
	"github.com/candrewlee14/webman/cmd/dev"
	"github.com/candrewlee14/webman/cmd/group"
	"github.com/candrewlee14/webman/cmd/remove"
	"github.com/candrewlee14/webman/cmd/run"
	"github.com/candrewlee14/webman/cmd/search"
	switchcmd "github.com/candrewlee14/webman/cmd/switch"
	"github.com/candrewlee14/webman/cmd/version"
)

func init() {
	rootCmd.AddCommand(add.AddCmd)
	rootCmd.AddCommand(config.ConfigCmd)
	rootCmd.AddCommand(dev.DevCmd)
	rootCmd.AddCommand(remove.RemoveCmd)
	rootCmd.AddCommand(run.RunCmd)
	rootCmd.AddCommand(switchcmd.SwitchCmd)
	rootCmd.AddCommand(group.GroupCmd)
	rootCmd.AddCommand(search.SearchCmd)
	rootCmd.AddCommand(version.VersionCmd)
}
