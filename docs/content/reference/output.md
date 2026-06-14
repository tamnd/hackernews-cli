---
title: "Output formats"
description: "The output contract every command shares: formats, fields, and templates."
weight: 30
---

Every command renders through one formatter, so the same flags work everywhere.
Pick a format with `-o`, or let `hn` choose: a table when writing to a terminal,
JSONL when piped.

## Formats

```bash
hn top -o table   # aligned columns for reading
hn top -o jsonl   # one JSON object per line, for piping
hn top -o json    # a single JSON array
hn top -o csv     # spreadsheet friendly
hn top -o tsv     # tab-separated
hn top -o url     # just the link for each record
hn top -o raw     # the bare field values, space-separated, no header
```

| Format | Best for |
|---|---|
| `table` | Reading on a terminal |
| `jsonl` | Piping into another tool, one object at a time |
| `json` | Loading a whole result as an array |
| `csv` / `tsv` | Spreadsheets and quick column math |
| `url` | Feeding links into other commands |
| `raw` | Bare values for `cut`, `awk`, and friends |

A note on the two JSON formats: `json` prints an array, except when there is a
single record, where it prints that one object. `jsonl` always prints one object
per line, which is what you want in a stream.

The `url` format prints each record's `url`, falling back to `hn_url` when a
record has no external link (Ask HN posts, most jobs, and comments).

## Narrowing columns

Keep only the fields you want, in the order you list them:

```bash
hn top --fields rank,score,title
hn item 48517377 --depth -1 --fields depth,by,text
```

The field names are the JSON keys you see in `jsonl` output. `--no-header` drops
the header row in `table`, `csv`, and `tsv`, which helps when a downstream tool
expects bare rows.

## Templating rows

For full control over each line, apply a Go text/template. The fields are the
**lowercase JSON keys**:

```bash
hn top --template '{{.score}}  {{.title}}'
hn top --template '{{.by}} -> {{.hn_url}}'
```

A `join` function is available for any list-valued field:

```bash
hn top --template '{{join "," .fields}}'
```

## Why auto-detection helps

Because the default adapts to the destination, the same command reads well by
hand and parses cleanly in a pipe:

```bash
hn top            # a table, because this is a terminal
hn top | wc -l    # JSONL, because this is a pipe
```

You only reach for `-o` when you want something other than that default.
