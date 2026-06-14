---
title: "Configuration"
description: "Global flags that tune how hn talks to the network."
weight: 20
---

`hn` needs no configuration to run: there is no config file, no environment
variable, and no key. Everything is a flag, and the defaults are tuned for
everyday use. Reach for these when you want to be gentler on the API or are on a
slow link.

## Networking flags

| Flag | Default | Meaning |
|---|---|---|
| `-j`, `--workers` | `16` | How many items to fetch at once. Story lists and comment trees fan out across this many requests. Lower it to be gentler on the API. |
| `--delay` | `50ms` | Minimum spacing between requests, applied across all workers. Raise it (for example `--delay 250ms`) if you start seeing 429s. |
| `--timeout` | `30s` | Per-request timeout. |
| `--retries` | `5` | How many times to retry a request that returns 429 or 5xx, with a linear backoff between attempts. |
| `--user-agent` | `hn/dev (+https://github.com/tamnd/hackernews-cli)` | The User-Agent header sent with every request. |

## Behaviour flags

| Flag | Default | Meaning |
|---|---|---|
| `-q`, `--quiet` | off | Suppress the progress lines `hn` writes to stderr, leaving stdout untouched. |
| `-h`, `--help` | | Help for the command. |
| `-v`, `--version` | | Print the version and exit. |

For the output flags (`-o`, `--fields`, `--no-header`, `--template`, `-n`), see
[output formats](/reference/output/).

## Environment variables

None. `hn` reads no environment variables of its own. Everything is configured
through flags.
