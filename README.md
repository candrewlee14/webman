<p align="center">
  <img src="/assets/webmanLogo.svg" width=260/>
 </p>
<h3 align="center">
  A cross-platform package manager for the web!
</h3>
<hr/>

Add, remove, and manage different versions of software.
No elevated permissions required!

Package recipes are located at https://github.com/candrewlee14/webman-pkgs.
Recipes are simple YAML files that make it easy to submit a new package.
Webman locates version numbers online and installs packages from the web, so you'll always have the most up-to-date software available!

Windows, Linux, and MacOS are all supported.

# Philosophy

I wanted a cross-platform package manager with no dependencies, a nice CLI, and a simple package configuration format.
I wanted a generalized version of [nvm](https://github.com/nvm-sh/nvm), [nvm-windows](https://github.com/coreybutler/nvm-windows), and [gvm](https://github.com/moovweb/gvm) for easily switching between package versions.
I wanted an easy way to install groups of packages, like the tools in [modern-unix](https://github.com/ibraheemdev/modern-unix).

That's why I built `webman`.

All of `webman`'s resources are located in `~/.webman`.
Simply remove that directory and all of `webman`'s packages and resources will be removed.

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

You can create new package recipes by adding a simple `[PKG_NAME].yaml` file in a cloned [webman-pkgs](https://github.com/candrewlee14/webman-pkgs) directory. Check if it is in a valid format with `webman check [WEBMAN-PKGS-DIR]`.

Next, you can test installing your local recipes with the `--local-recipes` flag on the `add` command, like `webman add [PKG_NAME] -l [WEBMAN-PKGS-DIR]`.

The package recipe format was built around making it easy to contribute new packages to webman, so if you're missing a package, go ahead and create it!

# Setup

Download the binary for your OS and architecture [here](/releases/latest).

Alternatively, if you have Go installed, run:

```bash
git clone https://github.com/candrewlee14/webman.git
cd webman
go install .
```

Next, add `~/.webman/bin` to your system PATH.
If you are on Windows, use `%USERPROFILE%` instead of `~`.

Now you're ready to use webman! Hope you enjoy :)

# Updating

You can update webman at any time using `webman install webman & webman switch webman` and choosing the newest version.
