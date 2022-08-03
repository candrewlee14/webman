function abort($message) {
  Write-Error -Message $message
  exit 1
}

function message($message) {
  Write-Information -Message $message -InformationAction Continue
}

function InstallWindows {
  $WindowsArchitecture = (Get-CimInstance Win32_operatingsystem).OSArchitecture
  $arch = if ($WindowsArchitecture.Contains("ARM")) {
    if ($WindowsArchitecture.StartsWith("64")) {
      "aarch64"
    } else {
      abort("arm is not supported")
    }
  } else {
    if ($WindowsArchitecture.StartsWith("64")) {
      "x86_64"
    } else {
      abort("32-bit is not supported")
    }
  }

  message("Finding webman for windows $arch ...")
  $latest = Invoke-WebRequest "https://api.github.com/repos/candrewlee14/webman/releases/latest" | ConvertFrom-Json
  $asset = $latest.assets | Where-Object { $_.name -like "webman*windows*$arch*" }

  $tmp = "$env:TEMP/webman.zip"
  $binDir = "$env:TEMP/webman"
  message("Retrieving $($asset.browser_download_url)")
  Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $tmp
  Expand-Archive -Path $tmp -DestinationPath $binDir
  $webman = Join-Path $binDir "webman.exe"
  & $webman add webman --switch
  Remove-Item -Path $tmp -Recurse
  Remove-Item -Path $binDir -Recurse
}

function InstallOtherOS {
  message("Installing webman via bash")
  curl -s https://raw.githubusercontent.com/candrewlee14/webman/main/scripts/install.sh | bash
}

$isWin = $env:OS -like "windows*"
if ($isWin) {
  InstallWindows
} else {
  InstallOtherOS
}
