---
title: "Pipe hn into scripts"
description: "Use JSONL output with jq and shell tools."
weight: 50
---

Every command emits structured records, and in a pipe the default format is
JSONL: one JSON object per line. That makes `hn` a clean source for `jq` and
ordinary shell tools.

## One object per line

```bash
hn top -n10 | jq -r '.title'
hn top -n50 | jq -r 'select(.score > 300) | .url'
```

Because the default adapts to the destination, you do not pass `-o` for this. At
a terminal the same command prints a table; in a pipe it prints JSONL.

## A whole result as one array

When a tool wants a single JSON document rather than a stream, use `-o json`:

```bash
hn search "rust" -n100 -o json | jq 'length'
```

## Templates for custom lines

For full control over each line without `jq`, apply a Go text/template. The
fields are the lowercase JSON keys:

```bash
hn top -n10 --template '{{.score}}  {{.title}}'
hn top -n10 --template '{{.by}} -> {{.hn_url}}'
```

## CSV for spreadsheets

```bash
hn ask -n100 --fields score,comments,title -o csv > ask.csv
hn ask -n100 --fields score,comments,title -o csv --no-header >> running.csv
```

## Exit codes in scripts

`hn` distinguishes "found nothing" from "broke":

```bash
if hn search "a-term-with-no-hits" -q >/dev/null; then
  echo "had results"
else
  case $? in
    3) echo "no results" ;;
    *) echo "request failed" ;;
  esac
fi
```

See [troubleshooting](/reference/troubleshooting/) for the full exit-code table.
