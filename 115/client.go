package _115

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/bluele/gcache"
	"github.com/gaoyb7/115drive-webdav/common"
	"github.com/gaoyb7/115drive-webdav/common/drive"
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

func Get115DriveClient() drive.DriveClient {
	return defaultClient
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
			req.Header.Set("Referer", "https://115.com/")
			req.Header.Set("User-Agent", UserAgent)
			req.Header.Set("Host", req.Host)
		},
	}

	// TODO: login check
	defaultClient.ImportCredential(uid, cid, seid)
}

func (c *DriveClient) GetFiles(filePath string) ([]drive.File, error) {
	filePath = slashClean(filePath)
	cacheKey := fmt.Sprintf("files:%s", filePath)
	if value, err := c.cache.Get(cacheKey); err == nil {
		return value.([]drive.File), nil
	}

	cid := "0"
	// TODO: get all pages
	resp, err := APIGetFiles(c.HttpClient, cid, 1000, 0)
	if err != nil {
		return nil, err
	}
	paths := splitPath(filePath)
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

	files := make([]drive.File, 0, len(resp.Data))
	for idx := range resp.Data {
		files = append(files, &resp.Data[idx])
	}
	if err := c.cache.SetWithExpire(cacheKey, files, time.Minute*2); err != nil {
		logrus.WithError(err).Errorf("call c.cache.SetWithExpire fail, file_path: %s", filePath)
	}

	return files, nil
}

func (c *DriveClient) GetFile(filePath string) (drive.File, error) {
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
	for _, file := range files {
		if file.GetName() == fileName {
			return file, nil
		}
	}

	return nil, common.ErrNotFound
}

func (c *DriveClient) GetFileURL(file drive.File) (string, error) {
	pickCode := file.(*FileInfo).PickCode
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

func (c *DriveClient) Proxy(w http.ResponseWriter, req *http.Request, targetURL string) {
	defer func() {
		if err := recover(); err != nil {
			if realErr, ok := err.(error); ok {
				if errors.Is(realErr, http.ErrAbortHandler) {
					return
				}
			}
			logrus.Errorf("panic: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	u, _ := url.Parse(targetURL)
	req.URL = u
	req.Host = u.Host
	c.reserveProxy.ServeHTTP(w, req)
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
	url := &url.URL{
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
