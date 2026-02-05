package x

import (
	"fmt"
)

func (c *Client) InitAuthenticatedSession() error {
	jar, err := LoadCookies(c.cfg.CookiesPath)
	if err != nil {
		return err
	}

	authToken := jar.Get("auth_token")
	ct0 := jar.Get("ct0")
	if authToken == "" || ct0 == "" {
		return fmt.Errorf("missing required cookies: auth_token and/or ct0")
	}

	c.cookies = jar
	c.ct0 = ct0
	c.cookieHeader = jar.CookieHeader()
	c.queryIDs = LoadQueryIDsOrDefault()

	return nil
}
