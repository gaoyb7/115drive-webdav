package _115

import (
	"context"
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
	"golang.org/x/time/rate"
)

var (
	defaultClient *DriveClient
)

type DriveClient struct {
	HttpClient   *http.Client
	cookieJar    *cookiejar.Jar
	cache        gcache.Cache
	reserveProxy *httputil.ReverseProxy
	limiter      *rate.Limiter
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
		limiter:    rate.NewLimiter(5, 10),
	}
	defaultClient.reserveProxy = &httputil.ReverseProxy{
		Transport: defaultClient.HttpClient.Transport,
		Director: func(req *http.Request) {
			req.Header.Set("Referer", "https://115.com/")
			req.Header.Set("User-Agent", UserAgent)
			req.Header.Set("Host", req.Host)
		},
	}

	defaultClient.ImportCredential(uid, cid, seid)

	// login check
	userID, err := APILoginCheck(defaultClient.HttpClient)
	if err != nil {
		panic(err)
	}
	if userID <= 0 {
		panic("115 drive login fail")
	}
	logrus.Infof("115 drive login succ, user_id: %d", userID)
}

func (c *DriveClient) GetFiles(dir string) ([]drive.File, error) {
	dir = slashClean(dir)
	cacheKey := fmt.Sprintf("files:%s", dir)
	if value, err := c.cache.Get(cacheKey); err == nil {
		return value.([]drive.File), nil
	}

	c.limiter.Wait(context.Background())
	getDirIDResp, err := APIGetDirID(c.HttpClient, dir)
	if err != nil {
		return nil, err
	}
	cid := getDirIDResp.CategoryID.String()

	pageSize := int64(1000)
	offset := int64(0)
	files := make([]drive.File, 0)
	for {
		resp, err := APIGetFiles(c.HttpClient, cid, pageSize, offset)
		if err != nil {
			return nil, err
		}

		for idx := range resp.Data {
			files = append(files, &resp.Data[idx])
		}

		offset = resp.Offset + pageSize
		if offset >= resp.Count {
			break
		}
	}
	if err := c.cache.SetWithExpire(cacheKey, files, time.Minute*2); err != nil {
		logrus.WithError(err).Errorf("call c.cache.SetWithExpire fail, dir: %s", dir)
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

func (c *DriveClient) RemoveFile(filePath string) error {
	fi, err := c.GetFile(filePath)
	if err != nil {
		return err
	}
	fid := fi.(*FileInfo).FileID.String()
	if fi.IsDir() {
		fid = fi.(*FileInfo).CategoryID.String()
	}

	pid := fi.(*FileInfo).ParentID.String()
	if err != nil {
		return err
	}

	resp, err := APIDeleteFile(c.HttpClient, fid, pid)
	if err != nil {
		return err
	}

	if !resp.State {
		return fmt.Errorf("remove file fail, state is false")
	}

	filePath = slashClean(filePath)
	filePath = strings.TrimRight(filePath, "/")
	dir, _ := path.Split(filePath)
	c.flushDir(dir)

	return nil
}

func (c *DriveClient) NewDir(dir string) error {
	getDirIDResp, err := APIGetDirID(c.HttpClient, dir)
	if err == nil {
		return nil
	}

	ret := strings.Split(dir, "/")
	var path, cname string
	for i := 0; i < len(ret); i++ {
		if i+1 >= len(ret) {
			cname = ret[i]
		} else {
			path += fmt.Sprintf("%s/", ret[i])
		}
	}

	getDirIDResp, err = APIGetDirID(c.HttpClient, path)
	if err != nil {
		return err
	}
	pid := getDirIDResp.CategoryID.String()
	resp, err := APINewDir(c.HttpClient, pid, cname)
	if err != nil {
		return err
	}

	if !resp.State {
		return fmt.Errorf("new dir fail, state is false")
	}

	c.flushDir(path)

	return nil
}

func (c *DriveClient) MoveFile(srcPath string, dstPath string) error {
	logrus.Infof("move file, src: %s, dst: %s", srcPath, dstPath)

	fi, err := c.GetFile(srcPath)
	if err != nil {
		return err
	}
	fid := fi.(*FileInfo).FileID.String()
	if fi.IsDir() {
		fid = fi.(*FileInfo).CategoryID.String()
	}

	srcPath = slashClean(srcPath)
	if srcPath == "/" || len(srcPath) == 0 {
		logrus.Warnf("invalid src_path: %s", srcPath)
		return nil
	}
	srcPath = strings.TrimRight(srcPath, "/")
	srcDir, srcFileName := path.Split(srcPath)

	dstPath = slashClean(dstPath)
	if dstPath == "/" || len(dstPath) == 0 {
		logrus.Warnf("invalid src_path: %s", srcPath)
		return nil
	}
	dstPath = strings.TrimRight(dstPath, "/")
	dstDir, dstFileName := path.Split(dstPath)

	dstDirFi, err := c.GetFile(dstDir)
	if err != nil {
		return err
	}
	if !dstDirFi.IsDir() {
		logrus.Errorf("dst dir not exists")
		return nil
	}

	if srcDir == dstDir {
		resp, err := APIRenameFile(c.HttpClient, fid, dstFileName)
		if err != nil {
			return err
		}
		if !resp.State {
			return fmt.Errorf("rename file fail, state is false")
		}
		c.flushDir(srcDir)
	} else {
		if srcFileName == dstFileName {
			resp, err := APIMoveFile(c.HttpClient, fid, dstDirFi.(*FileInfo).CategoryID.String())
			if err != nil {
				return err
			}
			if !resp.State {
				return fmt.Errorf("move file fail, state is false")
			}
			c.flushDir(srcDir)
			c.flushDir(dstDir)
		} else {
			logrus.Errorf("invalid dst filename")
			return nil
		}
	}

	return nil
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

func (c *DriveClient) flushDir(dir string) {
	dir = slashClean(dir)
	dir = strings.TrimRight(dir, "/")
	if len(dir) == 0 {
		dir = "/"
	}
	c.cache.Remove(fmt.Sprintf("files:%s", dir))
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
