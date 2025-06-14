<p align="center">
  <img src="/assets/webmanLogo.svg" width=260/>
 </p>
<h3 align="center">
  A cross-platform package manager for the web!
</h3>
<hr/>

![schema-linter](https://github.com/candrewlee14/webman-pkgs/actions/workflows/schema-linter.yml/badge.svg)
![bintest](https://github.com/candrewlee14/webman-pkgs/actions/workflows/bintest.yml/badge.svg)
![report-card](https://goreportcard.com/badge/github.com/candrewlee14/webman)

> [!WARNING]  
> This has been part of my workflow for a couple years now, but I'd recommend using [mise](https://mise.jdx.dev/) instead going forward.
> That will have better community support.

Add, remove, and manage different versions of web-distributed software binaries.
No elevated permissions required!

> Warning: This repo is still under development and has not stabilized.
There may be frequent breaking changes until a 1.x release.

Package recipes are located in the [webman-pkgs](https://github.com/candrewlee14/webman-pkgs) repo.
Recipes are simple YAML files that make it easy to submit a new package.
Webman locates version numbers online and installs packages from the web, so you'll always have the most up-to-date software available!

Windows (Powershell), Linux, and MacOS are supported!

<img alt="webman help example" src="/assets/tapes/webman.gif" width=600/>

### Installation

#### MacOS, Linux, Git Bash, WSL, etc.
```bash
curl https://raw.githubusercontent.com/candrewlee14/webman/main/scripts/install.sh | sh
```

#### Windows Powershell
Webman requires the ability to create symlinks!
Make sure to [enable developer mode](https://learn.microsoft.com/en-us/windows/apps/get-started/enable-your-device-for-development) so that admin privileges aren't required.
```powershell
Invoke-Expression (New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/candrewlee14/webman/main/scripts/install.ps1')
```

> NOTE: Never blindly run a shell script from the internet. Please check the source [shell](scripts/install.sh) or [powershell](scripts/install.ps1) file.
Alternatively, download the [latest release](/releases/latest) manually.


# Philosophy

I wanted a cross-platform package manager like [webi](https://github.com/webinstall/webi-installers) with no dependencies, a nice CLI, and a simple package configuration format.
I wanted a generalized version of [nvm](https://github.com/nvm-sh/nvm), [nvm-windows](https://github.com/coreybutler/nvm-windows), and [gvm](https://github.com/moovweb/gvm) for easily switching between package versions.
I wanted an easy way to install groups of packages, like the tools in [modern-unix](https://github.com/ibraheemdev/modern-unix).

That's why I built `webman`.

All of `webman`'s resources are located in `~/.webman`.
The only directory that needs to go on your system PATH is `~/.webman/bin`.
Simply remove the `~/.webman` directory and all of webman's packages and resources will be removed.

Security is an important priority to me here.
Package recipes cannot specify commands to be run, only endpoints to access.
Everything is implemented in Go.

# Examples

Below are examples of adding, removing, and switching with webman.

## Add Software

`webman add go` will install the latest version of Go.

`webman add zig@0.9.1` will install a specific version (`0.9.1`) of Zig.

`webman add rg lsd zig node go rg@12.0.0` will install each of the package versions listed.

`webman group add modern-unix` will allow checkbox selections for adding packages in the `modern-unix` group.

<img alt="webman add example" src="/assets/addNodeZigGoRg.gif" width=600/>

## Run Software

`webman run go` will run the in-use version of Go (if installed).

`webman run zig@0.9.1 --version` will run a specific version (`0.9.1`) of Zig with the argument `--version`.

`webman run node:npm --version` will run `npm --version` using the in-use version of node.

## Remove Software

`webman remove go` will allow you to select an installed version of the Go package to uninstall/

`webman group remove modern-unix` will allow checkbox selections for removing packages in the `modern-unix` group.

<img alt="webman remove example" src="/assets/removeNode.gif" width=600/>

## Switch to Other Versions of Software

`webman switch go` will allow you to select an installed version of the `go` package to switch to use.
If `rg --version` previously showed `13.0.0`, try running `webman switch rg` and selecting version `12.0.0` (after it has been installed).
Running `rg --version` again will say `12.0.0`.

Webman does version management.

<img alt="webman switch example" src="/assets/switchRg.gif" width=600/>

## Check Packages & Test Locally

You can create new package recipes by adding a simple recipe file in a cloned [webman-pkgs](https://github.com/candrewlee14/webman-pkgs) directory. Check if it is in a valid format with `webman dev check [WEBMAN-PKGS-DIR]`.

Next, `webman dev bintest [NEW-PKG] -l [WEBMAN-PKGs-DIR]` will do a cross-platform installation test on a package.

The package recipe format was built around making it easy to contribute new packages to webman, so if you're missing a package, go ahead and create it!

## Disable output color and ANSI escape codes

Set `NO_COLOR` environment variable to hava a raw console output.

# Setup

Run the script above or download the binary for your OS and architecture [here](/releases/latest).

Alternatively, if you have Go installed, run:

```bash
go install https://github.com/candrewlee14/webman@latest
```

Next, add `~/.webman/bin` to your system PATH.
If you are on Windows, use `%USERPROFILE%` instead of `~`.

Now you're ready to use webman! Hope you enjoy :)

# Updating

You can update webman at any time using `webman add webman --switch`.
