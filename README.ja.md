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
# ローカルリポジトリの解析（長いオプション名）
./repository-heatmap analyze --repo=/path/to/your/repo --output=./results

# ローカルリポジトリの解析（短いオプション名）
./repository-heatmap analyze -r /path/to/your/repo -o ./results

# リモートリポジトリの解析
./repository-heatmap analyze --repo=https://github.com/username/repo.git --output=./results
./repository-heatmap analyze -r https://github.com/username/repo.git -o ./results

# 特定のファイル種別のみを解析（例：Goファイル）
./repository-heatmap analyze --repo=/path/to/your/repo --file-type=go --output=./results
./repository-heatmap analyze -r /path/to/your/repo -t go -o ./results

# 特定のファイルパターンのみを解析
./repository-heatmap analyze --repo=/path/to/your/repo --file-pattern="*.go" --output=./results
./repository-heatmap analyze -r /path/to/your/repo -p "*.go" -o ./results

# 日付でフィルタリング
./repository-heatmap analyze --repo=/path/to/your/repo --since=2023-01-01 --output=./results
./repository-heatmap analyze -r /path/to/your/repo --since=2023-01-01 -o ./results

# 並列ワーカー数の指定（マルチスレッド解析）
./repository-heatmap analyze --repo=/path/to/your/repo --workers=4 --output=./results
./repository-heatmap analyze -r /path/to/your/repo -w 4 -o ./results
```

### 可視化コマンド

```bash
# 解析済みデータの可視化（SVG形式）
./repository-heatmap visualize --output=./results
./repository-heatmap visualize -o ./results

# リポジトリパスを指定して個別ファイルの内容も表示
./repository-heatmap visualize --output=./results --repo=/path/to/your/repo
./repository-heatmap visualize -o ./results -r /path/to/your/repo

# 表示するファイル数を指定
./repository-heatmap visualize --output=./results --max-files=200 --repo=/path/to/your/repo
./repository-heatmap visualize -o ./results -m 200 -r /path/to/your/repo

# 入力ファイルを明示的に指定
./repository-heatmap visualize --input=./results/repo-heatmap.json --output=./results
./repository-heatmap visualize -i ./results/repo-heatmap.json -o ./results
```

### analyzeコマンドのオプション

| オプション | 短いオプション | 説明 | デフォルト値 |
|------------|----------------|------|------------|
| `--repo` | `-r` | 解析するGitリポジトリのパスまたはURL (必須) | - |
| `--output` | `-o` | 出力ディレクトリ | `output` |
| `--file-type` | `-t` | 解析対象のファイル種別（例：go, js, py） | - |
| `--file-pattern` | `-p` | 解析対象のファイルパターン（例：*.go） | - |
| `--since` | - | 指定日付以降のコミットのみを解析 (YYYY-MM-DD形式) | - |
| `--workers` | `-w` | 並列処理に使用するワーカー数（0はCPU数に自動設定） | `0` |
| `--skip-clone` | `-s` | リポジトリが既にクローンされている場合はスキップ | `false` |
| `--debug` | `-d` | デバッグログをファイルに出力 | `false` |
| `--help` | `-h` | ヘルプを表示 | `false` |

### visualizeコマンドのオプション

| オプション | 短いオプション | 説明 | デフォルト値 |
|------------|----------------|------|------------|
| `--output` | `-o` | ヒートマップ出力ディレクトリ | `output` |
| `--format` | `-f` | 出力形式 (svgまたはwebp) | `svg` |
| `--repo` | `-r` | ファイル内容表示のためのリポジトリパス（個別ファイルSVGに必要） | - |
| `--max-files` | `-m` | ヒートマップに表示する最大ファイル数 | `100` |
| `--input` | `-i` | 入力JSONファイルのパス（指定しない場合は自動検出） | - |
| `--debug` | `-d` | デバッグログをファイルに出力 | `false` |
| `--help` | `-h` | ヘルプを表示 | `false` |

### グローバルオプション

以下のオプションはサブコマンドなしで使用できます：

| オプション | 短いオプション | 説明 |
|------------|----------------|------|
| `--version` | `-v` | バージョン情報を表示 |
| `--help` | `-h` | ヘルプを表示 |

## 出力

- `<リポジトリ名>-heatmap.json`: ファイルごとの変更頻度データ
- `<リポジトリ名>-repository-heatmap.svg`: リポジトリ全体のSVGヒートマップ
- `file-heatmaps/`: 各ファイルの変更頻度ヒートマップ（SVG形式）

## 特徴

- **SVGリンク機能**: リポジトリ全体のヒートマップから個別ファイルのヒートマップへのリンク
- **行単位の色分け**: 変更頻度に応じた行ごとの色分け表示
- **フィルタリング機能**: ファイル種別、パターン、日付によるフィルタリング
- **マルチスレッド処理**: 大規模リポジトリの高速解析
- **標準CLI慣習**: 長いオプション（`--option`）と短いオプション（`-o`）をサポート

## 依存関係

- [go-git](https://github.com/go-git/go-git): Gitリポジトリの操作
- [pflag](https://github.com/spf13/pflag): POSIX準拠のコマンドラインフラグ処理

## ライセンス

MIT License
