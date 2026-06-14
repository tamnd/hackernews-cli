---
title: "Troubleshooting"
description: "The handful of things that trip people up, and how to fix each one."
weight: 40
---

Most of these come down to network reality or how Hacker News serves its data,
not a bug.

## Exit codes

`hn` uses its exit code to tell apart success, emptiness, and failure, which
matters in scripts:

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | A fetch or runtime error |
| `2` | A usage error (bad flag, malformed id) |
| `3` | The request succeeded but found nothing |

Code `3` is the one people miss: an empty result is not an error, but it is also
not a silent success, so a pipeline can react to "nothing found" on its own.

## Requests start failing or returning 429

Hacker News rate-limits like any public site. `hn` already paces requests and
retries the transient failures, but a hard limit still means backing off. Raise
the spacing between requests with `--delay` (for example `--delay 500ms`), lower
the concurrency with `-j` (for example `-j 4`), and try again. A burst of 429 or
5xx responses is the site asking you to slow down, not a defect.

## Nothing is found for something you expected

`search` only returns what Algolia has indexed, and very new items can take a
little time to appear there. If a recent post is missing, give it a few minutes,
broaden the query, or check the live lists (`hn new`) instead. Confirm the id or
username is spelled the way the site uses it; `hn item` and `hn user` accept a
pasted URL, which avoids typos.

## The binary is not on your PATH

`go install` puts the binary in `$(go env GOPATH)/bin` (usually `~/go/bin`), and
a release archive leaves it wherever you unpacked it. If your shell cannot find
`hn`, add that directory to your `PATH`, or install the release archive into
`/usr/local/bin` as shown in [installation](/getting-started/installation/).

## macOS will not run the downloaded binary

A binary pulled from the internet carries a quarantine flag. Clear it once:

```bash
xattr -d com.apple.quarantine /usr/local/bin/hn
```

## Seeing what hn is doing

`hn` writes progress to stderr as it works (which story it is resolving, how many
submissions it is fetching). Those lines never touch stdout, so they do not
pollute a pipe. If you want them gone entirely, add `-q`. To see only the
progress and discard the data, redirect the other way: `hn top >/dev/null`.
