# hn

A command line for Hacker News.

`hn` is a single pure-Go binary. It reads Hacker News through the official
Firebase API and the Algolia search API, both open and key-free, shapes every
response into clean records, and prints them as a table, JSON, JSONL, CSV, TSV,
or plain URLs so the output drops straight into whatever you pipe it to.

It is a reader. There is no login, no voting, no commenting. `hn` is an
independent tool and is not affiliated with Hacker News or Y Combinator.

## Install

```bash
go install github.com/tamnd/hackernews-cli/cmd/hn@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/hackernews-cli/releases),
or run the container image:

```bash
docker run --rm ghcr.io/tamnd/hn:latest top -n5
```

## Quick start

```bash
hn top -n5                  # the 5 top stories
hn ask -n10                 # Ask HN
hn show                     # Show HN
hn new                      # newest submissions
hn jobs                     # YC job posts
hn item 48517377            # a story and its top-level comments
hn user pg                  # a profile
hn search "rust" -n10       # full-text search via Algolia
```

Output is a table when you are at a terminal and JSONL when you pipe, so
`hn top | jq` just works without any flags.

## Output formats

Pick a format with `-o`: `table`, `json`, `jsonl`, `csv`, `tsv`, `url`, or `raw`.

```bash
hn top -n3 -o table
```

```
RANK  ID        TYPE   TITLE                          SCORE  COMMENTS  BY
1     48517377  story  Noise infusion banned from...  723    455       nl
2     48518684  story  GLM 5.2 Is Out                 341    190       aloknnikhil
3     48516251  story  Every Frame Perfect            550    180       ravenical
```

```bash
hn top -n1 -o jsonl
```

```json
{"rank":1,"id":48517377,"type":"story","title":"Noise infusion banned from statistical products published by Census Bureau","url":"https://desfontain.es/blog/banning-noise.html","by":"nl","score":723,"comments":455,"time":1781358896,"date":"2026-06-13T13:54:56Z","text":"","hn_url":"https://news.ycombinator.com/item?id=48517377"}
```

Shape it further with `--fields`, `--no-header`, and `--template`:

```bash
hn top -n5 --fields rank,score,title          # only the columns you want
hn top -n20 -o url                             # just the links, for a reading queue
hn search "golang" -o csv --no-header > hits.csv
hn top -n5 --template '{{.score}}  {{.title}}' # any Go text/template, per record
```

## Commands

| Command | What it returns |
| --- | --- |
| `hn top` / `best` / `new` | front-page, best, and newest story lists |
| `hn ask` / `show` / `jobs` | Ask HN, Show HN, and job posts |
| `hn item <id>` | a story plus its comment tree (`--depth`, default 1; `-1` for the full thread) |
| `hn user <name>` | a profile, with `--submissions` to resolve their posts |
| `hn search <query>` | Algolia full-text search |
| `hn updates` | the firehose of recently changed items and profiles |
| `hn maxitem` | the current maximum item id |
| `hn version` | build information |

`item` and `user` accept a bare id or name as well as a full
`news.ycombinator.com/...` URL, so you can paste a link straight in:

```bash
hn item https://news.ycombinator.com/item?id=48517377
hn user https://news.ycombinator.com/user?id=pg
```

### Search

`search` rides the Algolia index, so it carries filters the Firebase API does
not have:

```bash
hn search "postgres" --tags story --sort date     # newest first
hn search "k8s" --points 100                       # at least 100 points
hn search "rust" --tags comment --comments 5       # comments on busy threads
hn search "llm" --since 24h                         # last day only (also 7d, 90m)
```

Comment hits carry the `story_id` and `story_title` of the thread they belong
to; story hits leave those empty since they would just echo the hit's own id.

## Global flags

```
-o, --output      table|json|jsonl|csv|tsv|url|raw (auto: table on a TTY, jsonl when piped)
    --fields      comma-separated columns to include
    --no-header   omit the header row in table/csv/tsv
    --template    Go text/template applied per record
-n, --limit       limit number of records (0 = per-command default)
-j, --workers     concurrent item fetches (default 16)
    --delay       minimum spacing between requests (default 50ms)
    --timeout     per-request timeout (default 30s)
    --retries     retry attempts on 429/5xx (default 5)
    --user-agent  User-Agent sent with each request
-q, --quiet       suppress progress on stderr
```

Story lists, comment trees, and submissions are fetched concurrently and
returned in rank order. The client paces itself and retries 429/5xx responses
with a linear backoff.

## Exit codes

| Code | Meaning |
| --- | --- |
| 0 | success |
| 1 | a fetch or runtime error |
| 2 | a usage error (bad flag, bad id) |
| 3 | the request succeeded but found nothing |

Code 3 lets you tell "empty" apart from "broke" in a script.

## Development

```
cmd/hn/          thin main, wires cli.Root into fang
cli/             the cobra command tree
hackernews/      the library: HTTP client, parsing, and data models
pkg/render/      the reflection-based record renderer
docs/            the documentation site
```

```bash
make build      # ./bin/hn
make test       # go test ./...
make vet        # go vet ./...
```

## Releasing

Push a version tag and GitHub Actions runs GoReleaser, which builds the
archives, Linux packages, the multi-arch GHCR image, checksums, SBOMs, and a
cosign signature:

```bash
git tag v0.1.1
git push --tags
```

The Homebrew and Scoop steps self-disable until their tokens exist, so a release
works with no extra secrets.

## License

Apache-2.0. See [LICENSE](LICENSE).
