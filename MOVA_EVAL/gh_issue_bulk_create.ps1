Param(
  [Parameter(Mandatory=$false)][string]$FilePath = "mova_issues.jsonl",
  [Parameter(Mandatory=$true)][string]$Repo
)
Get-Content -Path $FilePath | ForEach-Object {
  $obj = $_ | ConvertFrom-Json
  $labels = ($obj.labels) -join ","
  gh issue create --repo $Repo --title $obj.title --body $obj.body --label $labels | Out-Null
}
Write-Host "Done."
