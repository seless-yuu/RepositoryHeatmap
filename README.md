# Repository Heatmap

[日本語版README](README.ja.md)

A tool for visualizing the frequency of changes in Git repositories. It visualizes the frequency of changes for each file and line-by-line changes as SVG heatmaps.

## Features

- Analysis of Git repository commit history
- Aggregation of change frequency by file
- Analysis of change frequency by line
- Output of heatmap data in JSON format
- Generation of repository-wide and individual file heatmaps in SVG format
- Filtering by file pattern and file type
- Standard command line options (supporting both short and long formats)

## Recent Changes

- Improved command line processing: Implemented POSIX-compliant option processing using the pflag library
- Fixed the behavior of short options such as `-h` and `--help`
- Improved help display for subcommands (analyze, visualize)

## Installation

### Download Binary

You can download the latest binary for your platform from the [releases page](https://github.com/your-username/repositoryheatmap/releases):

- Windows AMD64: `repository-heatmap-windows-amd64.exe`
- macOS Apple Silicon: `repository-heatmap-darwin-arm64`
- Linux AMD64: `repository-heatmap-linux-amd64`

### Build from Source

```bash
# Clone the repository
git clone https://github.com/your-username/repositoryheatmap.git
cd repositoryheatmap

# Install dependencies
go mod download

# Build for your current platform
go build -o repository-heatmap ./cmd/repository-heatmap

# Cross-compile for specific platforms
# For Windows AMD64
go build -o repository-heatmap-windows-amd64.exe ./cmd/repository-heatmap/main.go

# For macOS Apple Silicon (ARM64)
GOOS=darwin GOARCH=arm64 go build -o repository-heatmap-darwin-arm64 ./cmd/repository-heatmap/main.go

# For Linux AMD64
GOOS=linux GOARCH=amd64 go build -o repository-heatmap-linux-amd64 ./cmd/repository-heatmap/main.go
```

### Build Scripts

For convenience, the repository includes build scripts that automatically compile binaries for multiple platforms:

- On Windows: Run `build.bat` to build for all supported platforms.
- On macOS/Linux: Run `./build.sh` to build for all supported platforms (make the script executable first with `chmod +x build.sh`).

These scripts will create executables for the following platforms:
- Windows AMD64: `repository-heatmap-windows-amd64.exe`
- Windows ARM64: `repository-heatmap-windows-arm64.exe`
- Linux AMD64: `repository-heatmap-linux-amd64`
- Linux ARM64: `repository-heatmap-linux-arm64`
- macOS ARM64: `repository-heatmap-darwin-arm64`

## Usage

The command is divided into two subcommands: `analyze` and `visualize`.

### Analyze Command

```bash
# Analyze a local repository (long option names)
./repository-heatmap analyze --repo=/path/to/your/repo --output=./results

# Analyze a local repository (short option names)
./repository-heatmap analyze -r /path/to/your/repo -o ./results

# Analyze a remote repository
./repository-heatmap analyze --repo=https://github.com/username/repo.git --output=./results
./repository-heatmap analyze -r https://github.com/username/repo.git -o ./results

# Analyze only specific file types (e.g., Go files)
./repository-heatmap analyze --repo=/path/to/your/repo --file-type=go --output=./results
./repository-heatmap analyze -r /path/to/your/repo -t go -o ./results

# Analyze only specific file patterns
./repository-heatmap analyze --repo=/path/to/your/repo --file-pattern="*.go" --output=./results
./repository-heatmap analyze -r /path/to/your/repo -p "*.go" -o ./results

# Filter by date
./repository-heatmap analyze --repo=/path/to/your/repo --since=2023-01-01 --output=./results
./repository-heatmap analyze -r /path/to/your/repo --since=2023-01-01 -o ./results

# Specify the number of parallel workers (multi-threaded analysis)
./repository-heatmap analyze --repo=/path/to/your/repo --workers=4 --output=./results
./repository-heatmap analyze -r /path/to/your/repo -w 4 -o ./results

# Display help
./repository-heatmap analyze --help
./repository-heatmap analyze -h
```

### Visualize Command

```bash
# Visualize analyzed data (SVG format)
./repository-heatmap visualize --output=./results
./repository-heatmap visualize -o ./results

# Specify repository path to also display file contents
./repository-heatmap visualize --output=./results --repo=/path/to/your/repo
./repository-heatmap visualize -o ./results -r /path/to/your/repo

# Specify the number of files to display
./repository-heatmap visualize --output=./results --max-files=200 --repo=/path/to/your/repo
./repository-heatmap visualize -o ./results -m 200 -r /path/to/your/repo

# Explicitly specify the input file
./repository-heatmap visualize --input=./results/repo-heatmap.json --output=./results
./repository-heatmap visualize -i ./results/repo-heatmap.json -o ./results

# Display help
./repository-heatmap visualize --help
./repository-heatmap visualize -h
```

### Analyze Command Options

| Option | Short Option | Description | Default Value |
|------------|----------------|------|------------|
| `--repo` | `-r` | Path or URL of the Git repository to analyze (required) | - |
| `--output` | `-o` | Output directory | `output` |
| `--file-type` | `-t` | File type to analyze (e.g., go, js, py) | - |
| `--file-pattern` | `-p` | File pattern to analyze (e.g., *.go) | - |
| `--since` | - | Analyze only commits after the specified date (YYYY-MM-DD format) | - |
| `--workers` | `-w` | Number of workers for parallel processing (0 automatically sets based on CPU count) | `0` |
| `--skip-clone` | `-s` | Skip cloning if the repository is already cloned | `false` |
| `--debug` | `-d` | Output debug logs to a file | `false` |
| `--help` | `-h` | Display help | `false` |

### Visualize Command Options

| Option | Short Option | Description | Default Value |
|------------|----------------|------|------------|
| `--output` | `-o` | Heatmap output directory | `output` |
| `--format` | `-f` | Output format (svg or webp) | `svg` |
| `--repo` | `-r` | Repository path for displaying file contents (needed for individual file SVGs) | - |
| `--max-files` | `-m` | Maximum number of files to display in the heatmap | `100` |
| `--input` | `-i` | Path to input JSON file (automatically detected if not specified) | - |
| `--debug` | `-d` | Output debug logs to a file | `false` |
| `--help` | `-h` | Display help | `false` |

### Global Options

The following options can be used without a subcommand:

| Option | Short Option | Description |
|------------|----------------|------|
| `--version` | `-v` | Display version information |
| `--help` | `-h` | Display help |

## Output

- `<repository_name>-heatmap.json`: Change frequency data by file
- `<repository_name>-repository-heatmap.svg`: SVG heatmap of the entire repository
- `file-heatmaps/`: Change frequency heatmaps for each file (SVG format)

## Feature Highlights

- **SVG Link Functionality**: Links from the repository-wide heatmap to individual file heatmaps
- **Line-by-Line Coloring**: Display of line-by-line coloring based on change frequency
- **Filtering Functionality**: Filtering by file type, pattern, and date
- **Multi-threaded Processing**: Fast analysis of large repositories
- **Standard CLI Conventions**: Full support for both long options (`--option`) and short options (`-o`)

## Command Line Processing Improvements

- **pflag Library**: Implemented POSIX-compliant command line flag processing
- **Intuitive Interface**: Both `--help` and `-h` can be used to display help for all commands and subcommands
- **Standardized Option Processing**: Support for both long and short forms for all options
- **Clear Help Messages**: Detailed display of usage and examples for each command

## Dependencies

- [go-git](https://github.com/go-git/go-git): Git repository operations
- [pflag](https://github.com/spf13/pflag): POSIX-compliant command line flag processing

## License

MIT License
