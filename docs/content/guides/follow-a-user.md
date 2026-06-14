---
title: "Follow a user"
description: "Look up a profile and list what someone has submitted."
weight: 40
---

`hn user` fetches a profile by name or pasted URL:

```bash
hn user pg
hn user https://news.ycombinator.com/user?id=pg
```

The record carries the karma, the account creation date, the about text (tags
stripped), and a count of how much the account has submitted.

## List their submissions

Add `--submissions` to resolve the account's posts into story records. `-n`
caps how many it fetches (default 20), newest first:

```bash
hn user pg --submissions -n10
hn user pg --submissions -n10 --fields score,title,hn_url
```

The first record is still the profile; the rows after it are the submissions, so
in a pipe you can split them apart by shape:

```bash
hn user pg --submissions -n50 -o jsonl |
  jq -r 'select(.title) | "\(.score)\t\(.title)"'
```
