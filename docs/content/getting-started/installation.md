---
title: "Installation"
description: "Install hn from a prebuilt binary, a package, go install, Docker, or source."
weight: 20
---

`hn` is a single static binary with no runtime dependencies. Pick whichever
method fits your machine.

The newest version is always on the
[releases page](https://github.com/tamnd/hackernews-cli/releases/latest). The
direct links below point at **v0.1.1**.

## Download a prebuilt binary

Every release ships archives for Linux, macOS, Windows, and FreeBSD on common
architectures.

| Platform | Architecture | Download |
|---|---|---|
| macOS | Apple Silicon (arm64) | [hn_0.1.1_darwin_arm64.tar.gz](https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/hn_0.1.1_darwin_arm64.tar.gz) |
| macOS | Intel (amd64) | [hn_0.1.1_darwin_amd64.tar.gz](https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/hn_0.1.1_darwin_amd64.tar.gz) |
| Linux | x86-64 (amd64) | [hn_0.1.1_linux_amd64.tar.gz](https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/hn_0.1.1_linux_amd64.tar.gz) |
| Linux | arm64 | [hn_0.1.1_linux_arm64.tar.gz](https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/hn_0.1.1_linux_arm64.tar.gz) |
| Linux | armv7 | [hn_0.1.1_linux_armv7.tar.gz](https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/hn_0.1.1_linux_armv7.tar.gz) |
| Linux | 386 | [hn_0.1.1_linux_386.tar.gz](https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/hn_0.1.1_linux_386.tar.gz) |
| Windows | x86-64 (amd64) | [hn_0.1.1_windows_amd64.zip](https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/hn_0.1.1_windows_amd64.zip) |
| Windows | arm64 | [hn_0.1.1_windows_arm64.zip](https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/hn_0.1.1_windows_arm64.zip) |
| FreeBSD | x86-64 (amd64) | [hn_0.1.1_freebsd_amd64.tar.gz](https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/hn_0.1.1_freebsd_amd64.tar.gz) |
| FreeBSD | arm64 | [hn_0.1.1_freebsd_arm64.tar.gz](https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/hn_0.1.1_freebsd_arm64.tar.gz) |

### Install it (macOS and Linux)

Unpack the archive and move `hn` onto your `PATH`:

```bash
VERSION=0.1.1
OS=linux        # or darwin
ARCH=amd64      # or arm64, armv7, 386
curl -fsSL -O "https://github.com/tamnd/hackernews-cli/releases/download/v${VERSION}/hn_${VERSION}_${OS}_${ARCH}.tar.gz"
tar xzf "hn_${VERSION}_${OS}_${ARCH}.tar.gz"
sudo install -m 0755 hn /usr/local/bin/hn
hn version
```

On macOS, if Gatekeeper blocks the unsigned binary, clear the quarantine
attribute once with `xattr -d com.apple.quarantine /usr/local/bin/hn`.

### Install it (Windows)

Unzip the archive and move `hn.exe` somewhere on your `PATH` (for example a
folder you add under `%LOCALAPPDATA%`), then run `hn version` in a new terminal.

## Linux packages

Native packages put `hn` in `/usr/bin` and let your package manager track it.

```bash
# Debian / Ubuntu (amd64; also arm64, armhf, i386)
curl -fsSL -O https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/hn_0.1.1_amd64.deb
sudo dpkg -i hn_0.1.1_amd64.deb

# Fedora / RHEL / openSUSE (x86_64; also aarch64, armv7hl, i386)
sudo rpm -i https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/hn-0.1.1-1.x86_64.rpm

# Alpine (x86_64; also aarch64, armv7, x86)
curl -fsSL -O https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/hn_0.1.1_x86_64.apk
sudo apk add --allow-untrusted hn_0.1.1_x86_64.apk
```

## With Go

```bash
go install github.com/tamnd/hackernews-cli/cmd/hn@latest
```

That puts `hn` in `$(go env GOPATH)/bin`, which is `~/go/bin` unless you moved
it. Make sure that directory is on your `PATH`.

## Container image

The multi-arch image is published to the GitHub Container Registry:

```bash
docker run --rm ghcr.io/tamnd/hn:latest top -n5
```

Pin a version with `ghcr.io/tamnd/hn:0.1.1`.

## From source

```bash
git clone https://github.com/tamnd/hackernews-cli
cd hackernews-cli
make build        # produces ./bin/hn
./bin/hn version
```

## Verify a download

Each release includes a `checksums.txt`, signed with keyless
[cosign](https://docs.sigstore.dev/). Verify the archive you grabbed:

```bash
# checksum
curl -fsSL -O https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/checksums.txt
sha256sum -c --ignore-missing checksums.txt

# signature (optional, needs the cosign CLI)
curl -fsSL -O https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/checksums.txt.sig
curl -fsSL -O https://github.com/tamnd/hackernews-cli/releases/download/v0.1.1/checksums.txt.pem
cosign verify-blob \
  --certificate checksums.txt.pem \
  --signature checksums.txt.sig \
  --certificate-identity-regexp 'https://github.com/tamnd/hackernews-cli' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  checksums.txt
```

## Checking the install

```bash
hn version
```

prints the version, commit, build date, platform, and Go version, then exits.
