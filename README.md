```
__          __  _
\ \        / / | |
 \ \  /\  / /__| |__  _ __ ___   __ _ _ __
  \ \/  \/ / _ \ '_ \| '_ ' _ \ / _' | '_ \
   \  /\  /  __/ |_) | | | | | | (_| | | | |
    \/  \/ \___|_.__/|_| |_| |_|\__,_|_| |_|
    
```

A cross-platform package manager for the web!

Add, remove, and manage different versions of software.
No elevated permissions required!

Package recipes are located at https://github.com/candrewlee14/webman-pkgs.
Recipes are simple YAML files that make it easy to submit a new package.
Webman locates version numbers online and installs packages from the web, so you'll always have the most up-to-date software available!

Windows, Linux, and MacOS are all supported.

# Examples

Below are examples of adding, removing, and switching with webman.

## Add Software

`webman add go` will install the latest version of Go.

`webman add zig@0.9.1` will install a specific version (`0.9.1`) of Zig.

`webman add rg lsd zig node go rg@12.0.0` will install each of the package versions listed.

<img alt="webman add example" src="/assets/addNodeZigGoRg.gif" width=600/>

## Run Software 

`webman run go` will run the in-use version of Go (if installed).

`webman run zig@0.9.1 --version` will run a specific version (`0.9.1`) of Zig with the argument `--version`.

`webman run node:npm --version` will run `npm --version` using the in-use version of node.

## Remove Software

`webman remove go` will allow you to select an installed version of the Go package to uninstall/

<img alt="webman remove example" src="/assets/removeNode.gif" width=600/>

## Switch to Other Versions of Software

`webman switch go` will allow you to select an installed version of the `go` package to switch to use.
If `rg --version` previously showed `13.0.0`, try running `webman switch rg` and selecting version `12.0.0` (after it has been installed).
Running `rg --version` again will say `12.0.0`. 

Webman does version management.

<img alt="webman switch example" src="/assets/switchRg.gif" width=600/>

## Check Packages & Test Locally

You can create new package recipes by adding a simple `[PKG_NAME].yaml` file in a cloned [webman-pkgs](https://github.com/candrewlee14/webman-pkgs) directory. Check if it is in a valid format with `webman check [WEBMAN-PKGS-DIR]`. 

Next, you can test installing your local recipes with the `--local-recipes` flag on the `add` command, like `webman add [PKG_NAME] -l [WEBMAN-PKGS-DIR]`.

The package recipe format was built around making it easy to contribute new packages to webman, so if you're missing a package, go ahead and create it!

# Setup

Binary releases will be coming soon, but until then, if you have Go installed, run:

```bash
git clone https://github.com/candrewlee14/webman.git
cd webman
go install .
```

Next, add `~/.webman/bin` to your system PATH. 
If you are on Windows, use `%USERPROFILE%` instead of `~`. 

Now you're ready to use webman! Hope you enjoy :)
