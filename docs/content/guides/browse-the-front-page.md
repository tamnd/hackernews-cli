---
title: "Browse the front page"
description: "Read the story lists and turn them into a link queue."
weight: 10
---

`hn` mirrors the lists you see on the site. Each one takes `-n` to set how many
stories to resolve.

```bash
hn top          # the front page (default 30)
hn best         # highest-rated recent stories
hn new          # newest submissions
hn ask          # Ask HN
hn show         # Show HN
hn jobs         # YC job posts
```

## Show only what you care about

`--fields` trims the columns to the ones you want, in the order you list them:

```bash
hn top -n10 --fields rank,score,comments,title
```

## Make a reading queue

The `url` format prints just the link for each story, which is exactly what a
queue is:

```bash
hn top -n30 -o url > reading.txt
```

Self-posts (Ask HN and most jobs) have no external link, so their `url` falls
back to the Hacker News discussion page. To always get the discussion link
instead of the article, ask for the `hn_url` field:

```bash
hn top -n30 --fields hn_url -o raw
```

## Sort and slice locally

Because every row is structured, you can post-process with ordinary tools:

```bash
hn top -n100 -o jsonl | jq -r 'select(.score > 500) | "\(.score)\t\(.title)"'
```
