package main

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/fatih/color"
    "webman/pkgparse"
)

var webmanStr string = `
 __          __  _                           
 \ \        / / | |                          
  \ \  /\  / /__| |__  _ __ ___   __ _ _ __  
   \ \/  \/ / _ \ '_ \| '_ ' _ \ / _' | '_ \ 
    \  /\  /  __/ |_) | | | | | | (_| | | | |
     \/  \/ \___|_.__/|_| |_| |_|\__,_|_| |_|

`

func main() {
    homeDir, err := os.UserHomeDir();
    if err != nil {
       panic(err);
    }
    webmanDir := filepath.Join(homeDir, "/.webman"); 
    if _, err := os.Stat(webmanDir); err != nil {
        if os.IsNotExist(err) {
            // dir does not exist
            color.Cyan(webmanStr);
            fmt.Println("Creating webman directory: ", color.GreenString(webmanDir));
            os.Mkdir(webmanDir, 0777)
            os.Mkdir(filepath.Join(webmanDir, "/bin"), 0777)
            os.Mkdir(filepath.Join(webmanDir, "/pkgs"), 0777)
        } else {
            // other error
            panic(err);
        }
    }
    pkgConf := pkgparse.ParsePkgConfig("go")
    latest := pkgConf.GetLatestVersion()
    url := pkgConf.GetAssetUrl(latest)
    fmt.Println(latest, url)
    
}

