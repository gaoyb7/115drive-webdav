package _115

import (
	"net/http"
	"net/http/cookiejar"
	neturl "net/url"
	"strings"

	"github.com/bluele/gcache"
)

var (
	defaultClient *DriveClient
)

const (
	CookieDomain115   = ".115.com"
	CookieDomainAnxia = ".anxia.com"
)

type DriveClient struct {
	webClient *http.Client
	cookieJar *cookiejar.Jar
	cache     gcache.Cache
}

func MustInit115DriveClient(uid string, cid string, seid string) {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	defaultClient = &DriveClient{
		cookieJar: cookieJar,
		webClient: &http.Client{Jar: cookieJar},
		cache:     gcache.New(10000).LFU().Build(),
	}
	// TODO: login check
	defaultClient.ImportCredential(uid, cid, seid)
}

func Get115DriveClient() *DriveClient {
	return defaultClient
}

func (c *DriveClient) ImportCredential(uid string, cid string, seid string) {
	cookies := map[string]string{
		"UID":  uid,
		"CID":  cid,
		"SEID": seid,
	}
	c.importCookies(CookieDomain115, "/", cookies)
	c.importCookies(CookieDomainAnxia, "/", cookies)
}

func (c *DriveClient) importCookies(domain string, path string, cookies map[string]string) {
	url := &neturl.URL{
		Scheme: "https",
		Path:   "/",
	}
	if domain[0] == '.' {
		url.Host = "www" + domain
	} else {
		url.Host = domain
	}
	cks := make([]*http.Cookie, 0)
	for name, value := range cookies {
		cookie := &http.Cookie{
			Name:     name,
			Value:    value,
			Domain:   domain,
			Path:     path,
			HttpOnly: true,
		}
		cks = append(cks, cookie)
	}
	c.cookieJar.SetCookies(url, cks)
}

func (c *DriveClient) ExportCookies(url string) string {
	u, _ := neturl.Parse(url)
	cookies := make(map[string]string)
	for _, cookie := range c.cookieJar.Cookies(u) {
		cookies[cookie.Name] = cookie.Value
	}
	if len(cookies) > 0 {
		buf, isFirst := strings.Builder{}, true
		for ck, cv := range cookies {
			if !isFirst {
				buf.WriteString("; ")
			}
			buf.WriteString(ck)
			buf.WriteRune('=')
			buf.WriteString(cv)
			isFirst = false
		}
		return buf.String()
	}
	return ""
}

func (c *DriveClient) GetWebClient() *http.Client {
	return c.webClient
}
