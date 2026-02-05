package x

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type CookieJar struct {
	Values  map[string]string
	Cookies []*http.Cookie
}

type cookieEditorCookie struct {
	Name           string      `json:"name"`
	Value          string      `json:"value"`
	Domain         string      `json:"domain"`
	Path           string      `json:"path"`
	Secure         bool        `json:"secure"`
	HTTPOnly       bool        `json:"httpOnly"`
	Session        bool        `json:"session"`
	ExpirationDate interface{} `json:"expirationDate"`
	SameSite       *string     `json:"sameSite"`
	PartitionKey   interface{} `json:"partitionKey"`
	StoreID        interface{} `json:"storeId"`
}

func LoadCookies(filePath string) (*CookieJar, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	trimmed := bytes.TrimPrefix(data, []byte("\xEF\xBB\xBF"))
	trimmed = bytes.TrimLeft(trimmed, " \t\r\n")
	if len(trimmed) > 0 {
		switch trimmed[0] {
		case '[', '{':
			jar, err := ParseCookieEditorJSON(trimmed)
			if err != nil {
				return nil, fmt.Errorf("parse cookies JSON: %w", err)
			}
			if len(jar.Values) == 0 {
				return nil, fmt.Errorf("no cookies parsed from %s", filePath)
			}
			return jar, nil
		}
	}

	jar, err := ParseNetscapeCookies(data)
	if err != nil {
		return nil, fmt.Errorf("parse cookies Netscape format: %w", err)
	}
	if len(jar.Values) == 0 {
		return nil, fmt.Errorf("no cookies parsed from %s", filePath)
	}
	return jar, nil
}

func ParseNetscapeCookies(data []byte) (*CookieJar, error) {
	jar := &CookieJar{Values: map[string]string{}}

	s := bufio.NewScanner(bytes.NewReader(data))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}

		httpOnly := false
		if strings.HasPrefix(line, "#HttpOnly_") {
			httpOnly = true
			line = strings.TrimPrefix(line, "#HttpOnly_")
		} else if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 7 {
			continue
		}

		domain := parts[0]
		path := parts[2]
		secure := strings.EqualFold(parts[3], "TRUE")
		expiryRaw := parts[4]
		name := parts[5]
		value := parts[6]
		if name == "" {
			continue
		}

		cookie := &http.Cookie{
			Name:     name,
			Value:    value,
			Domain:   domain,
			Path:     path,
			Secure:   secure,
			HttpOnly: httpOnly,
		}

		if exp, err := strconv.ParseInt(expiryRaw, 10, 64); err == nil && exp > 0 {
			cookie.Expires = time.Unix(exp, 0).UTC()
		}

		jar.Values[name] = value
		jar.Cookies = append(jar.Cookies, cookie)
	}

	if err := s.Err(); err != nil {
		return nil, err
	}
	return jar, nil
}

func ParseCookieEditorJSON(data []byte) (*CookieJar, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var raw interface{}
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}

	var entries []cookieEditorCookie
	switch v := raw.(type) {
	case []interface{}:
		buf, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(buf, &entries); err != nil {
			return nil, err
		}
	case map[string]interface{}:
		buf, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		var single cookieEditorCookie
		if err := json.Unmarshal(buf, &single); err != nil {
			return nil, err
		}
		entries = []cookieEditorCookie{single}
	default:
		return nil, fmt.Errorf("unexpected JSON type")
	}

	jar := &CookieJar{Values: map[string]string{}}
	for _, entry := range entries {
		if entry.Name == "" {
			continue
		}

		cookie := &http.Cookie{
			Name:     entry.Name,
			Value:    entry.Value,
			Domain:   entry.Domain,
			Path:     entry.Path,
			Secure:   entry.Secure,
			HttpOnly: entry.HTTPOnly,
		}

		if !entry.Session {
			exp, ok, err := toUnixSeconds(entry.ExpirationDate)
			if err != nil {
				return nil, fmt.Errorf("cookie %q has invalid expirationDate", entry.Name)
			}
			if ok {
				cookie.Expires = time.Unix(exp, 0).UTC()
			}
		}

		jar.Values[entry.Name] = entry.Value
		jar.Cookies = append(jar.Cookies, cookie)
	}

	return jar, nil
}

func toUnixSeconds(v interface{}) (int64, bool, error) {
	if v == nil {
		return 0, false, nil
	}

	switch n := v.(type) {
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0, false, err
		}
		return int64(f), true, nil
	case float64:
		return int64(n), true, nil
	case float32:
		return int64(n), true, nil
	case int64:
		return n, true, nil
	case int:
		return int64(n), true, nil
	default:
		return 0, false, fmt.Errorf("unsupported type")
	}
}

func (j *CookieJar) Get(name string) string {
	if j == nil {
		return ""
	}
	return j.Values[name]
}

func (j *CookieJar) CookieHeader() string {
	if j == nil || len(j.Values) == 0 {
		return ""
	}

	names := make([]string, 0, len(j.Values))
	for k := range j.Values {
		names = append(names, k)
	}
	sort.Strings(names)

	parts := make([]string, 0, len(names))
	for _, k := range names {
		v := j.Values[k]
		if v == "" {
			continue
		}
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, "; ")
}
