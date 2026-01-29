package x

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

type CookieJar struct {
	Values map[string]string
}

func LoadNetscapeCookies(filePath string) (*CookieJar, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	jar := &CookieJar{Values: map[string]string{}}

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}

		// The Netscape cookie format commonly prefixes HttpOnly cookies with "#HttpOnly_".
		// Those lines are NOT comments and must be parsed.
		if strings.HasPrefix(line, "#HttpOnly_") {
			line = strings.TrimPrefix(line, "#HttpOnly_")
		} else if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 7 {
			continue
		}

		name := parts[5]
		value := parts[6]
		if name == "" {
			continue
		}
		jar.Values[name] = value
	}

	if err := s.Err(); err != nil {
		return nil, err
	}
	if len(jar.Values) == 0 {
		return nil, fmt.Errorf("no cookies parsed from %s", filePath)
	}
	return jar, nil
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
