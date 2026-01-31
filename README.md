# xdlxd

A X (Twitter) media downloader written in Go.

> ⚠️ Use responsibly. Download only content you have the right to access and comply with X's rules/ToS.

---

## Quick start (recommended): use the prebuilt binary

Download the latest release binary for your platform from GitHub Releases:

- Linux (x86_64): `xdl_linux_amd64`
- Windows (x86_64): `xdl_windows_amd64.exe`

Repo: https://github.com/baptistax/xdlxd

### 1) Put `cookies.txt` next to the binary (required)

`xdl` **requires** a Netscape cookies export named `cookies.txt` in the same folder as the executable.

## Cookies

This project uses cookies in the standard Netscape format (`cookies.txt`).

All authentication tests are performed using cookies exported with the
**Cookie-Editor** browser extension.

Recommended workflow:

1. Install Cookie-Editor in your browser
2. Login to the target website
3. Export cookies in Netscape format
4. Save as `cookies.txt`
5. Pass the file to the tool

Other formats (such as JSON exports) are not supported.

Folder example:

    xdl/
      xdl_linux_amd64        # or xdl_windows_amd64.exe
      cookies.txt            # REQUIRED

### 2) Run

Linux:

    chmod +x ./xdl_linux_amd64
    ./xdl_linux_amd64 username or tweet

Windows (PowerShell / CMD):

    ./xdl_windows_amd64.exe username or tweet

Downloads are saved under `out/<target>/`.

---

## Build from source (manual compilation)

### Requirements

- Go 1.22+
- `cookies.txt` in Netscape format (**required**)

### Build

    git clone https://github.com/baptistax/xdlxd
    cd xdlxd
    go build -o xdl ./cmd/xdl
    ./xdl <tweet_url_or_id_or_username>

Dev (no build):

    go run ./cmd/xdl <tweet_url_or_id_or_username>

---

## cookies.txt (required)

`xdl` expects a **Netscape cookies.txt export** for `x.com` / `twitter.com`.

Typical flow:

1. Log in to X in your browser.
2. Export cookies using a cookie export tool/extension.
3. Export in Netscape format and save as `cookies.txt`.



## Output

By default, files go to:

    out/
      <target>/
        ...

## Troubleshooting

### “cookies.txt not found”

- Ensure `cookies.txt` is in the same directory as the binary (and you run the command from that directory).

### Empty results / errors

- Cookies may be expired — export a fresh `cookies.txt`.
- Confirm the cookies include `x.com` / `twitter.com` entries.

---

## Maintainers: releases via GoReleaser

This repository can publish Linux/Windows binaries to GitHub Releases using:

- `.goreleaser.yaml`
- `.github/workflows/release.yml`
