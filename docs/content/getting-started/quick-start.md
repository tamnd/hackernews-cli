---
title: "Quick start"
description: "Run your first hn commands."
weight: 30
---

Once `hn` is on your `PATH`, start with the front page:

```bash
hn top -n5
```

```
RANK  ID        TYPE   TITLE                          SCORE  COMMENTS  BY
1     48517377  story  Noise infusion banned from...  723    455       nl
2     48518684  story  GLM 5.2 Is Out                 341    190       aloknnikhil
3     48516251  story  Every Frame Perfect            550    180       ravenical
...
```

## Read a thread

Pass a story id (or paste its URL) to `item`. `--depth` controls how far down
the comment tree it walks:

```bash
hn item 48517377 --depth 1       # the story and its top-level comments
hn item 48517377 --depth -1      # the whole thread
```

## Search

```bash
hn search "rust" -n10
hn search "postgres" --sort date         # newest first
hn search "llm" --since 24h --points 100 # last day, at least 100 points
```

## Pipe it anywhere

At a terminal you get a table. In a pipe you get JSONL, so `jq` and friends just
work with no flags:

```bash
hn top -n20 | jq -r '.title'        # titles only
hn top -n50 -o url > reading.txt    # a link queue
hn ask -n30 --fields score,title -o csv > ask.csv
```

## Next

- The [guides](/guides/) walk through common jobs end to end.
- The [CLI reference](/reference/cli/) lists every command and flag.
- [Output formats](/reference/output/) covers `-o`, `--fields`, and templates.
