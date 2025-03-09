@echo off
echo Repository Heatmap �r���h�X�N���v�g
echo ==============================

echo Windows AMD64�����r���h��...
set GOOS=windows
set GOARCH=amd64
go build -o repository-heatmap-windows-amd64.exe ./cmd/repository-heatmap/main.go
if %ERRORLEVEL% neq 0 (
    echo Windows AMD64�����r���h�Ɏ��s���܂����B
    exit /b %ERRORLEVEL%
)
echo Windows AMD64�����r���h����: repository-heatmap-windows-amd64.exe

echo Windows ARM64�����r���h��...
set GOOS=windows
set GOARCH=arm64
go build -o repository-heatmap-windows-arm64.exe ./cmd/repository-heatmap/main.go
if %ERRORLEVEL% neq 0 (
    echo Windows ARM64�����r���h�Ɏ��s���܂����B
    exit /b %ERRORLEVEL%
)
echo Windows ARM64�����r���h����: repository-heatmap-windows-arm64.exe

echo Linux AMD64�����r���h��...
set GOOS=linux
set GOARCH=amd64
go build -o repository-heatmap-linux-amd64 ./cmd/repository-heatmap/main.go
if %ERRORLEVEL% neq 0 (
    echo Linux AMD64�����r���h�Ɏ��s���܂����B
    exit /b %ERRORLEVEL%
)
echo Linux AMD64�����r���h����: repository-heatmap-linux-amd64

echo Linux ARM64�����r���h��...
set GOOS=linux
set GOARCH=arm64
go build -o repository-heatmap-linux-arm64 ./cmd/repository-heatmap/main.go
if %ERRORLEVEL% neq 0 (
    echo Linux ARM64�����r���h�Ɏ��s���܂����B
    exit /b %ERRORLEVEL%
)
echo Linux ARM64�����r���h����: repository-heatmap-linux-arm64

echo macOS ARM64�����r���h��...
set GOOS=darwin
set GOARCH=arm64
go build -o repository-heatmap-darwin-arm64 ./cmd/repository-heatmap/main.go
if %ERRORLEVEL% neq 0 (
    echo macOS ARM64�����r���h�Ɏ��s���܂����B
    exit /b %ERRORLEVEL%
)
echo macOS ARM64�����r���h����: repository-heatmap-darwin-arm64

echo ���ׂẴr���h���������܂����B
echo �������ꂽ�o�C�i��:
echo - repository-heatmap-windows-amd64.exe
echo - repository-heatmap-windows-arm64.exe
echo - repository-heatmap-linux-amd64
echo - repository-heatmap-linux-arm64
echo - repository-heatmap-darwin-arm64 