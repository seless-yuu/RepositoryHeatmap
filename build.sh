#!/bin/bash

echo "Repository Heatmap ビルドスクリプト"
echo "=============================="

# Windows AMD64向けビルド
echo "Windows AMD64向けビルド中..."
GOOS=windows GOARCH=amd64 go build -o repository-heatmap-windows-amd64.exe ./cmd/repository-heatmap/main.go
if [ $? -ne 0 ]; then
    echo "Windows AMD64向けビルドに失敗しました。"
    exit 1
fi
echo "Windows AMD64向けビルド完了: repository-heatmap-windows-amd64.exe"

# Windows ARM64向けビルド
echo "Windows ARM64向けビルド中..."
GOOS=windows GOARCH=arm64 go build -o repository-heatmap-windows-arm64.exe ./cmd/repository-heatmap/main.go
if [ $? -ne 0 ]; then
    echo "Windows ARM64向けビルドに失敗しました。"
    exit 1
fi
echo "Windows ARM64向けビルド完了: repository-heatmap-windows-arm64.exe"

# Linux AMD64向けビルド
echo "Linux AMD64向けビルド中..."
GOOS=linux GOARCH=amd64 go build -o repository-heatmap-linux-amd64 ./cmd/repository-heatmap/main.go
if [ $? -ne 0 ]; then
    echo "Linux AMD64向けビルドに失敗しました。"
    exit 1
fi
echo "Linux AMD64向けビルド完了: repository-heatmap-linux-amd64"
chmod +x repository-heatmap-linux-amd64

# Linux ARM64向けビルド
echo "Linux ARM64向けビルド中..."
GOOS=linux GOARCH=arm64 go build -o repository-heatmap-linux-arm64 ./cmd/repository-heatmap/main.go
if [ $? -ne 0 ]; then
    echo "Linux ARM64向けビルドに失敗しました。"
    exit 1
fi
echo "Linux ARM64向けビルド完了: repository-heatmap-linux-arm64"
chmod +x repository-heatmap-linux-arm64

# macOS ARM64向けビルド
echo "macOS ARM64向けビルド中..."
GOOS=darwin GOARCH=arm64 go build -o repository-heatmap-darwin-arm64 ./cmd/repository-heatmap/main.go
if [ $? -ne 0 ]; then
    echo "macOS ARM64向けビルドに失敗しました。"
    exit 1
fi
echo "macOS ARM64向けビルド完了: repository-heatmap-darwin-arm64"
chmod +x repository-heatmap-darwin-arm64

echo "すべてのビルドが完了しました。"
echo "生成されたバイナリ:"
echo "- repository-heatmap-windows-amd64.exe"
echo "- repository-heatmap-windows-arm64.exe"
echo "- repository-heatmap-linux-amd64"
echo "- repository-heatmap-linux-arm64"
echo "- repository-heatmap-darwin-arm64" 