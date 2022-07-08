package _115

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	neturl "net/url"
	"path"
	"strings"
	"time"

	"github.com/bluele/gcache"
	"github.com/gaoyb7/115drive-webdav/common"
	"github.com/sirupsen/logrus"
)

var (
	defaultClient *DriveClient
)

type DriveClient struct {
	HttpClient   *http.Client
	cookieJar    *cookiejar.Jar
	cache        gcache.Cache
	reserveProxy *httputil.ReverseProxy
}

func MustInit115DriveClient(uid string, cid string, seid string) {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	defaultClient = &DriveClient{
		HttpClient: &http.Client{Jar: cookieJar},
		cookieJar:  cookieJar,
		cache:      gcache.New(10000).LFU().Build(),
	}
	defaultClient.reserveProxy = &httputil.ReverseProxy{
		Transport: defaultClient.HttpClient.Transport,
		Director: func(req *http.Request) {
			targetURL := req.URL.Query().Get("target")
			u, _ := url.Parse(targetURL)
			req.URL = u
			req.Host = u.Host
			req.Header.Set("Referer", "https://115.com/")
			req.Header.Set("User-Agent", UserAgent)
			req.Header.Set("Host", u.Host)
		},
	}

	// TODO: login check
	defaultClient.ImportCredential(uid, cid, seid)
}

func Get115DriveClient() *DriveClient {
	return defaultClient
}

func (c *DriveClient) GetFiles(path string) ([]FileInfo, error) {
	path = slashClean(path)
	cacheKey := fmt.Sprintf("files:%s", path)
	if value, err := c.cache.Get(cacheKey); err == nil {
		return value.([]FileInfo), nil
	}

	cid := "0"
	// TODO: get all pages
	resp, err := APIGetFiles(c.HttpClient, cid, 1000, 0)
	if err != nil {
		return nil, err
	}
	paths := splitPath(path)
	for idx := 0; idx < len(paths); idx++ {
		found := false
		for _, fileInfo := range resp.Data {
			if fileInfo.Name == paths[idx] {
				if !fileInfo.IsDir() {
					logrus.Errorf("not dir")
					return nil, common.ErrNotFound
				}
				found = true
				cid = fileInfo.CategoryID.String()
				break
			}
		}
		if !found {
			return nil, common.ErrNotFound
		}
		resp, err = APIGetFiles(c.HttpClient, cid, 1000, 0)
		if err != nil {
			return nil, err
		}
	}

	if err := c.cache.SetWithExpire(cacheKey, resp.Data, time.Minute*2); err != nil {
		logrus.WithError(err).Errorf("call c.cache.SetWithExpire fail, path: %s", path)
	}

	return resp.Data, nil
}

func (c *DriveClient) GetFile(filePath string) (*FileInfo, error) {
	filePath = slashClean(filePath)
	if filePath == "/" || len(filePath) == 0 {
		return &FileInfo{CategoryID: "0"}, nil
	}

	filePath = strings.TrimRight(filePath, "/")
	dir, fileName := path.Split(filePath)

	files, err := c.GetFiles(dir)
	if err != nil {
		return nil, err
	}
	for _, fileInfo := range files {
		if fileInfo.Name == fileName {
			return &fileInfo, nil
		}
	}

	return nil, common.ErrNotFound
}

func (c *DriveClient) GetURL(pickCode string) (string, error) {
	cacheKey := fmt.Sprintf("url:%s", pickCode)
	if value, err := c.cache.Get(cacheKey); err == nil {
		return value.(string), nil
	}

	info, err := APIGetDownloadURL(c.HttpClient, pickCode)
	if err != nil {
		return "", err
	}

	if err := c.cache.SetWithExpire(cacheKey, info.URL.URL, time.Minute*2); err != nil {
		logrus.WithError(err).Errorf("call c.cache.SetWithExpire fail, url: %s", info.URL.URL)
	}

	return info.URL.URL, nil
}

func (c *DriveClient) Proxy(w http.ResponseWriter, r *http.Request) {
	c.reserveProxy.ServeHTTP(w, r)
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

func slashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}

func splitPath(name string) []string {
	name = strings.Trim(name, "/")
	if len(name) == 0 {
		return nil
	}

	return strings.Split(name, "/")
}
