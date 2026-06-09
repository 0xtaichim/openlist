# openlist

A simple and efficient Go-based list management tool.

## Features
- **Project Structure**: Clean and modularized Go project design.
- **CLI Support**: Built with command-line interactions in mind.
- **Configurable**: Easy to manage settings through the `config` module.

## Project Structure
```text
.
├── client/     # Client implementation
├── cmd/        # CLI command definitions
├── config/     # Configuration management
├── model/      # Data types and models
├── main.go     # Application entry point
└── openlist    # Compiled binary
```

## Getting Started

### Prerequisites
- Go 1.21 or higher (recommended)

### Installation
Clone the repository:
```bash
git clone git@github.com:0xtaichim/openlist.git
cd openlist
```

### Build
To build the project from source:
```bash
go build -o openlist main.go
```

## CLI Usage

### Batch Rename

Preview a batch rename plan without changing files:
```bash
openlist fs batch-rename \
  --dir /media \
  --names "S01E01_openlist.txt,S01E02_openlist.txt" \
  --replace "openlist=agent" \
  --regex-replace "S01E0*=" \
  --insert "0=episode-" \
  --case lower \
  --dry-run
```

Execute the same kind of plan by omitting `--dry-run`. The command prints a JSON
result for every item and calls the existing OpenList rename API for changed
names.

Supported transformations:
- `--replace old=new`: literal find and replace; repeatable.
- `--regex-replace pattern=replacement`: regular expression replacement; repeatable.
- `--insert index=text`: insert at a 0-based rune index; repeatable.
- `--delete letters,digits,special,serial,punctuation,unit,arrow,drawing`: remove character classes.
- `--case lower|upper|title|first`: convert case.
- `--chinese simplified|traditional`: convert common simplified/traditional Chinese characters.
- `--include-extension`: apply transformations to the extension too; by default the extension is preserved.
- `--paths /a/file.txt,/b/file.txt`: rename explicit full paths instead of names under one directory.
- `--all --dir /media`: list and rename all direct children returned by the server for that directory.

## License
MIT License (or your preferred license)

---
*Created with ✨ by Hikari Yagami*
