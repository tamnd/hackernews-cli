---
title: "hn"
description: "A command line for Hacker News."
heroTitle: "Hacker News, from the command line"
heroLead: "A command line for Hacker News. One pure-Go binary, no API key, output that pipes into the rest of your tools."
heroPrimaryURL: "/getting-started/quick-start/"
heroPrimaryText: "Get started"
---

`hn` reads Hacker News through the official Firebase API and the Algolia search
API, both open and key-free, and prints clean records you can read at a terminal
or pipe into the next tool.

```bash
hn top -n5                 # the five top stories
hn item 48517377 --depth 2 # a thread, two levels deep
hn search "rust" --since 24h
hn top -o url | head       # just the links
```

Output is a table when you are at a terminal and JSONL when you pipe, so
`hn top | jq` works with no flags.

## Where to go next

- New here? Read the [introduction](/getting-started/introduction/), then the
  [quick start](/getting-started/quick-start/).
- Installing? See [installation](/getting-started/installation/) for prebuilt
  binaries, packages, and one-line installers.
- Doing a specific job? The [guides](/guides/) are task-oriented walkthroughs.
- Need every flag? The [CLI reference](/reference/cli/) is the full surface.
