# xdlxd

A X (Twitter) media downloader written in Go.

Warning: use responsibly. Download only content you have the right to access and comply with X's rules and Terms of Service.

## What It Does

- Download media from a single tweet URL or tweet ID
- Download media from a user profile timeline
- Show a simple terminal banner and live status output while it runs

Downloads are saved under `out/<target>/`.

## Cookies Workflow

`xdl` reads a file named exactly `cookies.txt`.

Cookie requirements:

- Single tweet mode: `cookies.txt` is optional
- Profile media mode: `cookies.txt` is required

Supported formats in `cookies.txt`:

- Netscape `cookies.txt` format
- Cookie-Editor JSON export format

No other cookie export formats are supported.

### How To Export Cookies

1. Log in to X with an active browser session.
2. Install the Cookie-Editor browser extension.
3. Export cookies in Netscape or JSON format.
4. Save the export into a file named exactly `cookies.txt`.
5. Place `cookies.txt` next to the `xdl` executable or in the current working directory.

Security note: never share `cookies.txt`. Treat cookies like credentials.

## Run Prebuilt Binary

Linux:

```bash
chmod +x ./xdl
./xdl <username>
./xdl <tweet_url_or_id>
```

Windows (PowerShell / CMD):

```powershell
./xdl.exe <username>
./xdl.exe <tweet_url_or_id>
```

## Build From Source

Requirements:

- Go 1.22+

Build:

```bash
git clone https://github.com/baptistax/xdlxd
cd xdlxd
go build -o xdl ./cmd/xdl
./xdl <username>
./xdl <tweet_url_or_id>
```

Dev (no build):

```bash
go run ./cmd/xdl <username>
go run ./cmd/xdl <tweet_url_or_id>
```

## Example Output

```text
===================================
 XDL | X/Twitter Media Downloader
===================================
[xdl] Target : someuser
[xdl] Mode   : profile media
[xdl] Output : C:\path\to\out
[xdl] Auth   : cookies loaded
[xdl] Scan complete | 20 item(s) found
[xdl] Download [========================] 20/20 | saved 18 | cached 2 | 154.2 MB
[xdl] Done -> C:\path\to\out\someuser
```

During scanning and downloads, the progress line updates in place instead of printing one line per file.

## Troubleshooting

### `cookies.txt not found`

- Ensure the file is named exactly `cookies.txt`.
- Place it next to the executable.
- If not found there, the tool also checks the current working directory.
- `cookies.txt` is only required for profile downloads.

### Auth errors or empty results

- Export a fresh cookie file from an active logged-in session.
- Ensure required X cookies such as `auth_token` and `ct0` are present.

### No media downloaded

- For tweet mode, verify that the tweet really contains media.
- For profile mode, verify that the account has media posts available to your logged-in session.
- If files already exist in `out/<target>/`, `xdl` will skip re-downloading them.
