---
title: "Introduction"
description: "What hn is and how it is put together."
weight: 10
---

A command line for Hacker News.

`hn` is a single binary. It reads Hacker News over plain HTTPS, shapes the
responses into clean records, and gets out of your way. There is nothing to sign
up for and nothing to run alongside it.

## The two APIs

Hacker News exposes two open, key-free APIs, and `hn` uses both:

- The official **Firebase** endpoint (`hacker-news.firebaseio.com/v0`) for live
  data: the story lists, single items and their comment trees, user profiles,
  the maximum item id, and the recent-changes feed.
- The **Algolia** endpoint (`hn.algolia.com/api/v1`) for full-text search across
  stories and comments, with filters Firebase does not offer (by tag, by date,
  by points, by comment count).

You never choose between them. Each command talks to whichever API serves it.

## How it is built

- A **library package** (`hackernews`) holds the HTTP client and the typed data
  models. It paces requests, sets an honest User-Agent, retries the transient
  429 and 5xx responses any public site throws under load, and fetches story
  lists and comment trees concurrently.
- A **command tree** (`cli`) wraps the library in subcommands that all share the
  same output formats and flags.
- One **`cmd/hn`** entry point ties them together.

## Scope

`hn` is a read-only client. It reads what Hacker News already serves publicly
and shapes it for you. There is no login, no voting, and no commenting: the
write surface needs a session and is out of scope. That narrow focus keeps `hn`
a single small binary with no database, no daemon, and no setup.

Next: [install it](/getting-started/installation/), then take the
[quick start](/getting-started/quick-start/).
