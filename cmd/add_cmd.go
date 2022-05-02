package cmd

import (
	"webman/cmd/add"
	"webman/cmd/check"
	"webman/cmd/remove"
	switchcmd "webman/cmd/switch"
)

func init() {
	rootCmd.AddCommand(add.AddCmd)
	rootCmd.AddCommand(check.CheckCmd)
	rootCmd.AddCommand(remove.RemoveCmd)
	rootCmd.AddCommand(switchcmd.SwitchCmd)
}
