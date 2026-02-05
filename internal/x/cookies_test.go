package x

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseCookieEditorJSON(t *testing.T) {
	const payload = `[
  {
    "name": "sessionid",
    "value": "123%3AabcDEF%3A10%3AAYj4XN8cLpcG9o0QkD8aijCTalWfV\\\"lPNF",
    "domain": ".instagram.com",
    "hostOnly": false,
    "path": "/",
    "secure": true,
    "httpOnly": true,
    "sameSite": null,
    "session": false,
    "firstPartyDomain": "",
    "partitionKey": null,
    "expirationDate": 1801846439.938,
    "storeId": null
  },
  {
    "name": "rur",
    "value": "\"RVA\\\\054617073445\\\\0541801846466:01fe8774f3e1c5f4\"",
    "domain": ".instagram.com",
    "hostOnly": false,
    "path": "/",
    "secure": true,
    "httpOnly": true,
    "sameSite": "lax",
    "session": true,
    "firstPartyDomain": "",
    "partitionKey": null,
    "storeId": null
  }
]`

	jar, err := ParseCookieEditorJSON([]byte(payload))
	if err != nil {
		t.Fatalf("ParseCookieEditorJSON() error = %v", err)
	}

	if got := jar.Get("sessionid"); !strings.Contains(got, `WfV\"lPNF`) {
		t.Fatalf("sessionid value mismatch: %q", got)
	}
	if got := jar.Get("rur"); got != `"RVA\\054617073445\\0541801846466:01fe8774f3e1c5f4"` {
		t.Fatalf("rur value mismatch: %q", got)
	}

	if len(jar.Cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(jar.Cookies))
	}

	first := jar.Cookies[0]
	if first.Domain != ".instagram.com" || first.Path != "/" || !first.Secure || !first.HttpOnly {
		t.Fatalf("first cookie attrs not preserved: %+v", first)
	}
	wantExpires := time.Unix(1801846439, 0).UTC()
	if !first.Expires.Equal(wantExpires) {
		t.Fatalf("expected expires %v, got %v", wantExpires, first.Expires)
	}

	second := jar.Cookies[1]
	if !second.Expires.IsZero() {
		t.Fatalf("session cookie should have zero expires, got %v", second.Expires)
	}
}

func TestParseNetscapeCookies(t *testing.T) {
	const payload = `# Netscape HTTP Cookie File
#HttpOnly_.x.com	TRUE	/	TRUE	1801846439	auth_token	abc%22def\\ghi
x.com	TRUE	/	TRUE	0	ct0	xyz123
`

	jar, err := ParseNetscapeCookies([]byte(payload))
	if err != nil {
		t.Fatalf("ParseNetscapeCookies() error = %v", err)
	}

	if got := jar.Get("auth_token"); got != `abc%22def\\ghi` {
		t.Fatalf("auth_token value mismatch: %q", got)
	}
	if got := jar.Get("ct0"); got != "xyz123" {
		t.Fatalf("ct0 value mismatch: %q", got)
	}

	if len(jar.Cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(jar.Cookies))
	}

	auth := jar.Cookies[0]
	if !auth.HttpOnly || !auth.Secure || auth.Domain != ".x.com" || auth.Path != "/" {
		t.Fatalf("auth cookie attrs not preserved: %+v", auth)
	}
	if auth.Expires.IsZero() {
		t.Fatalf("expected auth cookie expiry")
	}

	csrf := jar.Cookies[1]
	if !csrf.Expires.IsZero() {
		t.Fatalf("expected session cookie with no expiry, got %v", csrf.Expires)
	}
}

func TestLoadCookiesAutoDetectJSONWithBOM(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "cookies.txt")
	payload := "\xEF\xBB\xBF   [{\"name\":\"ct0\",\"value\":\"abc\",\"domain\":\".x.com\",\"path\":\"/\",\"secure\":true,\"httpOnly\":false,\"session\":true}]"
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	jar, err := LoadCookies(path)
	if err != nil {
		t.Fatalf("LoadCookies() error = %v", err)
	}
	if jar.Get("ct0") != "abc" {
		t.Fatalf("expected ct0 cookie")
	}
}
