---
title: "Search and filter"
description: "Full-text search with tags, dates, points, and comment counts."
weight: 30
---

`hn search` rides the Algolia index, which carries filters the live API does
not. The query is the one positional argument; the rest are flags.

```bash
hn search "rust"
hn search "postgres performance" -n25
```

## Pick what to search

`--tags` chooses the kind of item. The default is `story`.

```bash
hn search "rust"   --tags story        # stories (default)
hn search "rust"   --tags comment      # comments
hn search "show"   --tags show_hn      # Show HN
hn search "hiring" --tags job          # jobs
```

## Sort by date

The default sort is relevance. Switch to newest-first with `--sort date`:

```bash
hn search "llm" --sort date
```

## Filter by time, points, and comments

```bash
hn search "kubernetes" --since 24h        # last day (also 7d, 90m, and so on)
hn search "golang" --points 100           # at least 100 points
hn search "ai" --comments 50              # threads with at least 50 comments
```

`--since` understands Go durations (`90m`, `24h`) plus a `d` suffix for days
(`7d`, `30d`).

## Comment hits point back to their thread

When a hit is a comment, its record carries the `story_id` and `story_title` of
the thread it belongs to, so you can jump from a match to its context:

```bash
hn search "borrow checker" --tags comment -o jsonl |
  jq -r '"\(.story_title): \(.text)"'
```

Story hits leave those fields empty, since they would just echo the hit's own
id.
