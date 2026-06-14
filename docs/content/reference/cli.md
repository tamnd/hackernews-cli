---
title: "CLI"
description: "Every command and subcommand, with the flags that matter."
weight: 10
---

```
hn <command> [arguments] [flags]
```

Run `hn <command> --help` for the exact flag list on any command.

## Commands

| Command | What it does |
|---|---|
| `hn top` | Top stories (front page) |
| `hn best` | Highest-rated recent stories |
| `hn new` | Newest submissions |
| `hn ask` | Ask HN posts |
| `hn show` | Show HN posts |
| `hn jobs` | YC job listings |
| `hn item <id>` | One item and its comment tree |
| `hn user <name>` | A user profile, optionally with submissions |
| `hn search <query>` | Full-text search via Algolia |
| `hn updates` | Recently changed items and profiles |
| `hn maxitem` | The current maximum item id |
| `hn version` | Print version information |

The story-list commands (`top`, `best`, `new`, `ask`, `show`, `jobs`) default to
30 records; raise or lower it with `-n`.

## Command-specific flags

### `item <id>`

Accepts a bare id or a `news.ycombinator.com/item?id=...` URL.

| Flag | Default | Meaning |
|---|---|---|
| `--depth` | `1` | Comment-tree depth. `0` = item only, `-1` = the full tree |

### `user <name>`

Accepts a bare username or a `news.ycombinator.com/user?id=...` URL.

| Flag | Default | Meaning |
|---|---|---|
| `--submissions` | off | Also resolve and list the user's submissions (capped by `-n`, default 20) |

### `search <query>`

| Flag | Default | Meaning |
|---|---|---|
| `--tags` | `story` | What to search: `story`, `comment`, `ask_hn`, `show_hn`, `front_page`, `job` |
| `--sort` | `relevance` | `relevance` or `date` (newest first) |
| `--since` | none | Only results newer than a duration: `90m`, `24h`, `7d` |
| `--points` | `0` | Minimum points |
| `--comments` | `0` | Minimum comment count |

### `version`

| Flag | Default | Meaning |
|---|---|---|
| `--short` | off | Print just the version number |

## Global flags

These work on every command. See [output formats](/reference/output/) and
[configuration](/reference/configuration/) for detail.

| Flag | Default | Meaning |
|---|---|---|
| `-o`, `--output` | `auto` | `table`, `json`, `jsonl`, `csv`, `tsv`, `url`, `raw` |
| `--fields` | all | Comma-separated columns to include |
| `--no-header` | off | Omit the header row in `table`/`csv`/`tsv` |
| `--template` | none | Go text/template applied per record |
| `-n`, `--limit` | per command | Limit the number of records |
| `-j`, `--workers` | `16` | Concurrent item fetches |
| `--delay` | `50ms` | Minimum spacing between requests |
| `--timeout` | `30s` | Per-request timeout |
| `--retries` | `5` | Retry attempts on 429/5xx |
| `--user-agent` | `hn/dev ...` | User-Agent sent with each request |
| `-q`, `--quiet` | off | Suppress progress on stderr |
