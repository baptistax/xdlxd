# xdlxd

A X (Twitter) media downloader written in Go.

> ⚠️ Use responsibly. Download only content you have the right to access and comply with X's rules/ToS.

## Cookies workflow (required)

`xdl` reads a file named exactly `cookies.txt`.

Supported formats in `cookies.txt`:
- Netscape cookies.txt format
- Cookie-Editor JSON export format

No other cookie export formats are supported.

### Steps

1. Log in to X with an active browser session.
2. Install the **Cookie-Editor** browser extension (**only Cookie-Editor has been tested**).
3. Export cookies in **Netscape** format **or** **JSON** format.
4. Save the exported content into a file named exactly `cookies.txt`.
5. Place `cookies.txt` next to the `xdl` executable.
6. Run:
   - `xdl <username>`

> 🔒 Security warning: never share `cookies.txt`. Treat cookies like credentials/passwords.

## Run prebuilt binary

Linux:

```bash
chmod +x ./xdl_linux_amd64
./xdl_linux_amd64 <username>
```

Windows (PowerShell / CMD):

```powershell
./xdl_windows_amd64.exe <username>
```

Downloads are saved under `out/<target>/`.

## Build from source

Requirements:
- Go 1.22+

Build:

```bash
git clone https://github.com/baptistax/xdlxd
cd xdlxd
go build -o xdl ./cmd/xdl
./xdl <username>
```

Dev (no build):

```bash
go run ./cmd/xdl <username>
```

## Troubleshooting

### `cookies.txt not found`

- Ensure the file is named exactly `cookies.txt`.
- Place it next to the executable.
- If not found there, the tool also checks the current working directory.

### Auth errors / empty results

- Export a fresh cookie file from an active logged-in session.
- Ensure required X cookies (such as `auth_token`, `ct0`) are present.
