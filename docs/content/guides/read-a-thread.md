---
title: "Read a thread"
description: "Fetch a story and walk its comment tree to any depth."
weight: 20
---

`hn item` fetches one item and, by default, its top-level comments. It accepts a
bare id or a pasted URL:

```bash
hn item 48517377
hn item https://news.ycombinator.com/item?id=48517377
```

The first record is the story; the rest are comments, each carrying a `depth` so
you can see where it sits in the tree.

## Control the depth

`--depth` sets how far down the tree `hn` walks:

```bash
hn item 48517377 --depth 0     # the story only, no comments
hn item 48517377 --depth 1     # story plus direct replies (the default)
hn item 48517377 --depth 3     # three levels deep
hn item 48517377 --depth -1    # the entire thread
```

Comment trees are fetched concurrently, so even a full thread comes back fast.
Tune the parallelism with `-j` if you want to be gentler on the API.

## Just the comment text

The comment bodies arrive as plain text, with HTML tags stripped and entities
decoded, so they read cleanly in a terminal or a pipe:

```bash
hn item 48517377 --depth -1 --fields depth,by,text
```

```bash
# every top-level commenter on a thread
hn item 48517377 --depth 1 -o jsonl | jq -r '.by'
```
