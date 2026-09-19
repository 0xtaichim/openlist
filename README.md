# openlist

CLI for [OpenList](https://github.com/OpenListTeam/OpenList) (AList-compatible) file servers.

## Features

- Manage files remotely: list, search, mkdir, rename, copy, move, remove
- Offline download tasks
- Generate `.strm` files (local or remote) with concurrent directory walking
- XDG config (`~/.config/openlist/config.yaml`) plus `OPENLIST_*` env and flags

## Install

```bash
git clone git@github.com:0xtaichim/openlist.git
cd openlist
go build -o openlist .
```

Requires Go 1.23+.

## Configuration

Priority (low → high): defaults → config file → environment → flags.

```bash
openlist config set url http://localhost:5244
openlist config set token <your-token>
openlist config list
```

Environment:

- `OPENLIST_URL`
- `OPENLIST_TOKEN`
- `XDG_CONFIG_HOME` (config directory root)

Flags on every command: `--url` / `-u`, `--token` / `-t`.

## Examples

```bash
openlist fs list --path /
openlist fs get --path /Movies
openlist strm --src /Movies --dst /STRM --base-url https://example.com
openlist strm --src /Movies --dst ./strm --dst-type local --base-url https://example.com --dry-run
```

## License

MIT
