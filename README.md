# xdl

A minimal, clean Go CLI to download media from X (Twitter), using the same layout philosophy as `idl`.

It currently supports:

- Public tweet media download (no cookies required) via the syndication endpoint
- (Optional/experimental) authenticated profile media timelines (requires `cookies.txt`; may break as X changes often)

> ⚠️ Use responsibly. Download only content you have the right to access and comply with X's rules/ToS.

---

## Quick start (recommended): download the binary

This is the simplest way to use `xdl`: download the prebuilt binary from the GitHub Releases page.

### 1) Download

Download the asset for your platform:

- Linux (x86_64): `xdl_linux_amd64`
- Windows (x86_64): `xdl_windows_amd64.exe`

### 2) Optional: put `cookies.txt` next to the binary (authenticated flows only)

`xdl` will look for a Netscape format cookies export named `cookies.txt` in the same folder as the executable.

Folder example:

    xdl/
      xdl_linux_amd64        # or xdl_windows_amd64.exe
      cookies.txt            # optional (only needed for authenticated flows)

### 3) Run

Public tweet download (no cookies required):

Linux:

    chmod +x ./xdl_linux_amd64
    ./xdl_linux_amd64 <tweet_url_or_id>

Windows (PowerShell / CMD):

    ./xdl_windows_amd64.exe <tweet_url_or_id>

Authenticated profile download (cookies required):

Linux:

    chmod +x ./xdl_linux_amd64
    ./xdl_linux_amd64 <username>

Windows:

    ./xdl_windows_amd64.exe <username>

Downloads are saved under `out/<target>/`.

---

## Build from source (manual compilation)

### Requirements

- Go 1.22+
- (Optional) `cookies.txt` in Netscape format for authenticated flows

### Clone and build

    git clone <YOUR_REPO_URL>
    cd xdl
    go build -o xdl ./cmd/xdl
    ./xdl <tweet_url_or_id>

### Dev option: go run

    go run ./cmd/xdl <tweet_url_or_id>

---

## Cookies file (cookies.txt)

`xdl` expects a Netscape cookies.txt export.

Typical ways to obtain it:

- Log in to X in your browser.
- Export cookies for `x.com` / `twitter.com` using a cookie export extension/tool.
- Save/export as Netscape format and name it `cookies.txt`.

Notes:

- The file can include `#HttpOnly_` lines (supported).
- Comments starting with `#` are ignored.
- **Never commit `cookies.txt`** (this repo's `.gitignore` ignores it).

---

## Output structure

By default, downloads are stored in `out/`:

    out/
      <target>/
        <timestamp>_<media_id>.jpg
        <timestamp>_<media_id>.mp4
        <timestamp>_<media_id>_01.jpg
        ...

---

## Troubleshooting

### "cookies.txt not found"

- For tweet URLs/IDs, cookies are **not** required.
- For `<username>` mode, ensure `cookies.txt` is in the same folder as the binary (or run from that folder).

### Empty results / errors fetching profile

- Cookies may be expired. Export a fresh `cookies.txt`.
- X internal endpoints change often; the authenticated adapter is intentionally kept modular.

---

## Maintainers: release workflow (optional)

This repository can be set up to publish Linux/Windows binaries to GitHub Releases using GoReleaser. See `.goreleaser.yaml` and `.github/workflows/release.yml`.
