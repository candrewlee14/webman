/*
Copyright © 2022 Andrew Lee candrewlee14@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/candrewlee14/webman/multiline"
	"github.com/candrewlee14/webman/ui"
	"github.com/candrewlee14/webman/utils"

	"github.com/fatih/color"
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "webman",
	Short:         "A cross-platform package manager for the web",
	SilenceErrors: true,
	SilenceUsage:  true,
	Long: `
__          __  _
\ \        / / | |
 \ \  /\  / /__| |__  _ __ ___   __ _ _ __
  \ \/  \/ / _ \ '_ \| '_ ' _ \ / _' | '_ \
   \  /\  /  __/ |_) | | | | | | (_| | | | |
    \/  \/ \___|_.__/|_| |_| |_|\__,_|_| |_|

A cross-platform package manager for the web!
	- created by candrewlee14

`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	ansiOn := ui.AreAnsiCodesEnabled()
	if ansiOn {
		fmt.Printf("%s", multiline.HideCursor)
		defer fmt.Printf("%s", multiline.ShowCursor)
		cc.Init(&cc.Config{
			RootCmd:  rootCmd,
			Headings: cc.HiCyan + cc.Bold + cc.Underline,
			Commands: cc.HiYellow + cc.Bold,
			Example:  cc.Italic,
			ExecName: cc.Bold,
			Flags:    cc.Bold,
		})
	} else {
		color.NoColor = true
	}
	err := rootCmd.Execute()
	if err != nil {
		color.HiRed("%v", err)
		if ansiOn {
			fmt.Printf("%s", multiline.ShowCursor)
		}
		os.Exit(1)
	}
}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	utils.Init(homeDir)
	rootCmd.PersistentFlags().StringVarP(&utils.RecipeDirFlag, "local-recipes", "l", "", "use given local recipe directory")
}
