$swagger = Join-Path $env:GOPATH "bin\goctl-swagger.exe"
if (-not (Test-Path $swagger)) {
  throw "goctl-swagger is not installed. Run: go install github.com/zeromicro/goctl-swagger@latest"
}

goctl api plugin `
  --plugin "$swagger=swagger --filename swagger.json" `
  --api "core\core.api" `
  --dir "docs"
