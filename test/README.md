# Repository Heatmap テストディレクトリ

このディレクトリにはRepository Heatmapの開発に関連するテストコードとサンプルが含まれています。

## pflagのテストとサンプル

pflagライブラリはPOSIX準拠のコマンドラインフラグ処理を実装するためのGoライブラリです。このプロジェクトではpflagを使用してコマンドラインオプションを処理しています。

### テストファイル

- `pflag_debug.go` - pflagの問題点と修正方法を示す例
- `pflag_examples/simple.go` - シンプルなpflagの使用例
- `pflag_subcommand_example/main.go` - サブコマンドを使用したpflagの使用例

### pflag実装の改善点

元のコードでは以下の問題がありました：

1. **フラグの設定方法**：
   - `StringP()`, `BoolP()`などのメソッドを使用していたため、変数に直接値が割り当てられず、後でフラグの値を取得するために`Lookup()`や`GetBool()`などのメソッドを使用する必要がありました。
   - **修正**: `StringVarP()`, `BoolVarP()`などのメソッドを使用して、フラグを直接変数に割り当てるようにしました。

2. **ヘルプ表示の処理**：
   - ヘルプ表示のためのカスタム関数を使用していましたが、pflagの標準的な機能を活用していませんでした。
   - **修正**: `FlagSet.Usage`関数を適切に設定し、ヘルプ表示を一元化しました。

3. **フラグ解析の処理**：
   - フラグ解析後に`showHelp`変数を確認していましたが、この変数はフラグ解析時に設定されるため、解析前に明示的に`-h`や`--help`引数をチェックする必要がありました。
   - **修正**: 解析前に明示的に`-h`や`--help`引数をチェックする処理を追加しました。

### テスト実行方法

```bash
# pflag_debugの実行
cd test
go run pflag_debug.go problematic --help  # 問題のある方法のデモ
go run pflag_debug.go correct --help      # 正しい方法のデモ
go run pflag_debug.go explicit --help     # 明示的なチェックのデモ

# シンプルな例
cd pflag_examples
go run simple.go --help                   # ヘルプの表示
go run simple.go -h                       # 短縮形オプションでヘルプ表示
go run simple.go -r ./myrepo -o ./output # 短縮形オプションの使用

# サブコマンドの例
cd ../pflag_subcommand_example
go run main.go --help                     # メインヘルプの表示
go run main.go analyze --help             # analyzeコマンドのヘルプ表示
go run main.go visualize -h               # visualizeコマンドの短縮ヘルプ表示
``` 