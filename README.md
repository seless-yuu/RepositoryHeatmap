# Repository Heatmap

Gitリポジトリの変更頻度を可視化するヒートマップ生成ツールです。ファイルごとの変更頻度や行単位での変更頻度をSVGヒートマップとして可視化します。

## 機能

- Gitリポジトリのコミット履歴解析
- ファイルごとの変更頻度の集計
- 行単位での変更頻度の分析
- ヒートマップデータのJSON形式での出力
- SVG形式でのリポジトリヒートマップと個別ファイルヒートマップの生成
- ファイルパターンとファイル種別によるフィルタリング機能

## インストール

### バイナリのダウンロード

[リリースページ](https://github.com/your-username/repositoryheatmap/releases)から最新のバイナリをダウンロードできます。

### ソースからビルド

```bash
# リポジトリのクローン
git clone https://github.com/your-username/repositoryheatmap.git
cd repositoryheatmap

# 依存関係のインストール
go mod download

# ビルド
go build -o repository-heatmap ./cmd/repository-heatmap
```

## 使い方

コマンドは2つのサブコマンドに分かれています：`analyze`（解析）と`visualize`（可視化）です。

### 解析コマンド

```bash
# ローカルリポジトリの解析
./repository-heatmap analyze --repo=/path/to/your/repo --output=./results

# リモートリポジトリの解析
./repository-heatmap analyze --repo=https://github.com/username/repo.git --output=./results

# 特定のファイル種別のみを解析（例：Goファイル）
./repository-heatmap analyze --repo=/path/to/your/repo --file-type=go --output=./results

# 特定のファイルパターンのみを解析
./repository-heatmap analyze --repo=/path/to/your/repo --file-pattern="*.go" --output=./results

# 日付でフィルタリング
./repository-heatmap analyze --repo=/path/to/your/repo --since=2023-01-01 --output=./results

# 並列ワーカー数の指定（マルチスレッド解析）
./repository-heatmap analyze --repo=/path/to/your/repo --workers=4 --output=./results
```

### 可視化コマンド

```bash
# 解析済みデータの可視化（SVG形式）
./repository-heatmap visualize --output=./results

# リポジトリパスを指定して個別ファイルの内容も表示
./repository-heatmap visualize --output=./results --repo=/path/to/your/repo

# 表示するファイル数を指定
./repository-heatmap visualize --output=./results --max-files=200 --repo=/path/to/your/repo
```

### analyzeコマンドのオプション

| オプション | 説明 | デフォルト値 |
|------------|------|------------|
| `--repo` | 解析するGitリポジトリのパスまたはURL (必須) | - |
| `--output` | 出力ディレクトリ | `output` |
| `--file-type` | 解析対象のファイル種別（例：go, js, py） | - |
| `--file-pattern` | 解析対象のファイルパターン（例：*.go） | - |
| `--since` | 指定日付以降のコミットのみを解析 (YYYY-MM-DD形式) | - |
| `--workers` | 並列処理に使用するワーカー数（0はCPU数に自動設定） | `0` |
| `--skip-clone` | リポジトリが既にクローンされている場合はスキップ | `false` |
| `--debug` | デバッグログをファイルに出力 | `false` |

### visualizeコマンドのオプション

| オプション | 説明 | デフォルト値 |
|------------|------|------------|
| `--output` | ヒートマップ出力ディレクトリ | `output` |
| `--format` | 出力形式 (svgまたはwebp) | `svg` |
| `--repo` | ファイル内容表示のためのリポジトリパス（個別ファイルSVGに必要） | - |
| `--max-files` | ヒートマップに表示する最大ファイル数 | `100` |
| `--input` | 入力JSONファイルのパス（指定しない場合は自動検出） | - |
| `--debug` | デバッグログをファイルに出力 | `false` |

## 出力

- `<リポジトリ名>-heatmap.json`: ファイルごとの変更頻度データ
- `<リポジトリ名>-repository-heatmap.svg`: リポジトリ全体のSVGヒートマップ
- `file-heatmaps/`: 各ファイルの変更頻度ヒートマップ（SVG形式）

## 特徴

- **SVGリンク機能**: リポジトリ全体のヒートマップから個別ファイルのヒートマップへのリンク
- **行単位の色分け**: 変更頻度に応じた行ごとの色分け表示
- **フィルタリング機能**: ファイル種別、パターン、日付によるフィルタリング
- **マルチスレッド処理**: 大規模リポジトリの高速解析

## 依存関係

- [go-git](https://github.com/go-git/go-git): Gitリポジトリの操作

## ライセンス

MIT License
