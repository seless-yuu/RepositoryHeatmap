@echo off
echo Repository Heatmap ビルドスクリプト
echo ==============================

echo Windows AMD64向けビルド中...
set GOOS=windows
set GOARCH=amd64
go build -o repository-heatmap-windows-amd64.exe ./cmd/repository-heatmap/main.go
if %ERRORLEVEL% neq 0 (
    echo Windows AMD64向けビルドに失敗しました。
    exit /b %ERRORLEVEL%
)
echo Windows AMD64向けビルド完了: repository-heatmap-windows-amd64.exe

echo Windows ARM64向けビルド中...
set GOOS=windows
set GOARCH=arm64
go build -o repository-heatmap-windows-arm64.exe ./cmd/repository-heatmap/main.go
if %ERRORLEVEL% neq 0 (
    echo Windows ARM64向けビルドに失敗しました。
    exit /b %ERRORLEVEL%
)
echo Windows ARM64向けビルド完了: repository-heatmap-windows-arm64.exe

echo Linux AMD64向けビルド中...
set GOOS=linux
set GOARCH=amd64
go build -o repository-heatmap-linux-amd64 ./cmd/repository-heatmap/main.go
if %ERRORLEVEL% neq 0 (
    echo Linux AMD64向けビルドに失敗しました。
    exit /b %ERRORLEVEL%
)
echo Linux AMD64向けビルド完了: repository-heatmap-linux-amd64

echo Linux ARM64向けビルド中...
set GOOS=linux
set GOARCH=arm64
go build -o repository-heatmap-linux-arm64 ./cmd/repository-heatmap/main.go
if %ERRORLEVEL% neq 0 (
    echo Linux ARM64向けビルドに失敗しました。
    exit /b %ERRORLEVEL%
)
echo Linux ARM64向けビルド完了: repository-heatmap-linux-arm64

echo macOS ARM64向けビルド中...
set GOOS=darwin
set GOARCH=arm64
go build -o repository-heatmap-darwin-arm64 ./cmd/repository-heatmap/main.go
if %ERRORLEVEL% neq 0 (
    echo macOS ARM64向けビルドに失敗しました。
    exit /b %ERRORLEVEL%
)
echo macOS ARM64向けビルド完了: repository-heatmap-darwin-arm64

echo すべてのビルドが完了しました。
echo 生成されたバイナリ:
echo - repository-heatmap-windows-amd64.exe
echo - repository-heatmap-windows-arm64.exe
echo - repository-heatmap-linux-amd64
echo - repository-heatmap-linux-arm64
echo - repository-heatmap-darwin-arm64 