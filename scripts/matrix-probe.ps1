param([string]$Pkg = './cmd/nps')
$ErrorActionPreference='Continue'
$env:CGO_ENABLED='0'
$matrix = @(
  @{os='linux';   arch='amd64'},
  @{os='linux';   arch='386'},
  @{os='linux';   arch='arm';  arm='5'},
  @{os='linux';   arch='arm';  arm='6'},
  @{os='linux';   arch='arm';  arm='7'},
  @{os='linux';   arch='arm64'},
  @{os='linux';   arch='ppc64le'},
  @{os='linux';   arch='riscv64'},
  @{os='linux';   arch='loong64'},
  @{os='freebsd'; arch='amd64'},
  @{os='freebsd'; arch='arm64'},
  @{os='windows'; arch='386'},
  @{os='windows'; arch='amd64'},
  @{os='windows'; arch='arm64'},
  @{os='darwin';  arch='amd64'},
  @{os='darwin';  arch='arm64'}
)
$probe = Join-Path $env:TEMP 'nps_probe.bin'
$logName = 'matrix.' + ($Pkg -replace '[^a-zA-Z0-9]', '_') + '.log'
$logPath = Join-Path (Get-Location) $logName
Remove-Item $logPath -ErrorAction SilentlyContinue
function Emit($s) { $s | Tee-Object -FilePath $logPath -Append }
foreach ($t in $matrix) {
  $env:GOOS=$t.os; $env:GOARCH=$t.arch
  if ($t.arm) { $env:GOARM=$t.arm } else { Remove-Item Env:GOARM -ErrorAction SilentlyContinue }
  $tag = "$($t.os)/$($t.arch)$(if($t.arm){"v$($t.arm)"})"
  $err = & go build -o $probe $Pkg 2>&1 | Out-String
  if ($LASTEXITCODE -eq 0) {
    Emit ("OK   {0}" -f $tag)
  } else {
    $line = ($err -split "`n" | Where-Object { $_ -match 'error|undefined|cannot|not supported|no such' } | Select-Object -First 1).Trim()
    if (-not $line) { $line = ($err -split "`n" | Where-Object { $_.Trim() } | Select-Object -First 1).Trim() }
    Emit ("FAIL {0} :: {1}" -f $tag, $line)
    Add-Content $logPath ("        FULL_ERR_BEGIN`n" + $err + "`n        FULL_ERR_END")
  }
}
Emit '---DONE---'
Remove-Item Env:GOOS,Env:GOARCH,Env:GOARM -ErrorAction SilentlyContinue
