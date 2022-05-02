package cmd

import (
	"webman/cmd/add"
	"webman/cmd/check"
	"webman/cmd/remove"
	"webman/cmd/run"
	switchcmd "webman/cmd/switch"
)

func init() {
	rootCmd.AddCommand(add.AddCmd)
	rootCmd.AddCommand(check.CheckCmd)
	rootCmd.AddCommand(remove.RemoveCmd)
	rootCmd.AddCommand(run.RunCmd)
	rootCmd.AddCommand(switchcmd.SwitchCmd)
}
