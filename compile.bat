go build -o "C:/Windows/goact.exe" .

set GOARCH=amd64
set GOOS=linux
go build -o "build/linux_amd64" .

@REM set GOARCH=amd64
@REM set GOOS=darwin
@REM go tool dist install -v pkg/runtime
@REM go install -v -a std
@REM go build -ldflags -H=windowsgui -o "jvc_macos.app" .

set GOARCH=amd64
set GOOS=windows
go build -o "build/windows_amd64.exe" .
