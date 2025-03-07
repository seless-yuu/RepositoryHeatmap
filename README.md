# Repository Heatmap

Gitリポジトリの変更頻度を可視化するヒートマップ生成ツールです。ファイルごとの変更頻度や行単位での変更頻度をヒートマップとして可視化します。

## 機能

- Gitリポジトリのコミット履歴解析
- ファイルごとの変更頻度の集計
- 行単位での変更頻度の分析
- ヒートマップデータのJSON形式での出力
- SVGまたはWebP形式でのヒートマップ可視化

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

```bash
# ローカルリポジトリの解析
./repository-heatmap --repo=/path/to/your/repo

# リモートリポジトリの解析
./repository-heatmap --repo=https://github.com/username/repo.git

# 出力形式の指定
./repository-heatmap --repo=/path/to/your/repo --format=svg

# 出力ディレクトリの指定
./repository-heatmap --repo=/path/to/your/repo --output=./results
```

### オプション

| オプション | 説明 | デフォルト値 |
|------------|------|------------|
| `--repo` | 解析するGitリポジトリのパスまたはURL (必須) | - |
| `--output` | 出力ディレクトリ | `output` |
| `--format` | 出力形式 (svg または webp) | `svg` |
| `--skip-clone` | リポジトリが既にクローンされている場合はスキップ | `false` |
| `--version` | バージョン情報を表示 | `false` |

## 出力

- `<リポジトリ名>-heatmap.json`: ファイルごとの変更頻度データ
- `<リポジトリ名>-repository-heatmap.<形式>`: リポジトリ全体のヒートマップ
- `file-heatmaps/`: 各ファイルの変更頻度ヒートマップ

## 依存関係

- [go-git](https://github.com/go-git/go-git): Gitリポジトリの操作
- [go-chart](https://github.com/wcharczuk/go-chart): データの可視化

## ライセンス

MIT License
