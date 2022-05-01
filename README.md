# webman

A cross-platform package manager for the web!

Add, remove, and manage different versions of software.

Package recipes are located at https://github.com/candrewlee14/webman-pkgs.
Packages recipes are simple YAML files that make it easy to submit a new package.
Package recipes parse version numbers from the web, so you'll always have the most up-to-date software available!

Windows, Linux, and MacOS are all supported.

## Examples

### Add Software

`webman add go` will install the latest version of `go`

`webman add zig@0.9.1` will install a specific version (`0.9.1`) of `zig`

`webman add rg lsd zig node go rg@12.0.0` will install each of the package versions listed

### Remove Software

`webman remove go` will allow you to select an installed version of the `go` package to uninstall

### Switch to Other Version of Software

`webman switch go` will allow you to select an installed version of the `go` package to switch to use.
If `rg --version` previously showed `13.0.0`, try running `webman switch rg` and selecting version `12.0.0` (after it has been installed).
Running `rg --version` again will say `12.0.0`. 

Webman does version management :) 
