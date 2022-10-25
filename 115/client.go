package _115

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/bluele/gcache"
	"github.com/gaoyb7/115drive-webdav/common"
	"github.com/gaoyb7/115drive-webdav/common/drive"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type DriveClient struct {
	HttpClient   *resty.Client
	cache        gcache.Cache
	reserveProxy *httputil.ReverseProxy
	limiter      *rate.Limiter
}

func MustNew115DriveClient(uid string, cid string, seid string) *DriveClient {
	httpClient := resty.New().SetCookie(&http.Cookie{
		Name:     "UID",
		Value:    uid,
		Domain:   "www.115.com",
		Path:     "/",
		HttpOnly: true,
	}).SetCookie(&http.Cookie{
		Name:     "CID",
		Value:    cid,
		Domain:   "www.115.com",
		Path:     "/",
		HttpOnly: true,
	}).SetCookie(&http.Cookie{
		Name:     "SEID",
		Value:    seid,
		Domain:   "www.115.com",
		Path:     "/",
		HttpOnly: true,
	}).SetHeader("User-Agent", UserAgent)

	client := &DriveClient{
		HttpClient: httpClient,
		cache:      gcache.New(10000).LFU().Build(),
		limiter:    rate.NewLimiter(5, 1),
		reserveProxy: &httputil.ReverseProxy{
			Transport: httpClient.GetClient().Transport,
			Director: func(req *http.Request) {
				req.Header.Set("Referer", "https://115.com/")
				req.Header.Set("User-Agent", UserAgent)
				req.Header.Set("Host", req.Host)
			},
		},
	}

	// login check
	userID, err := APILoginCheck(client.HttpClient)
	if err != nil || userID <= 0 {
		logrus.WithError(err).Panicf("115 drive login fail")
	}
	logrus.Infof("115 drive login succ, user_id: %d", userID)

	return client
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

func (c *DriveClient) ServeContent(w http.ResponseWriter, req *http.Request, fi drive.File) {
	fileURL, err := c.GetFileURL(fi)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	logrus.Infof("proxy open [name: %v] [url: %v] [range: %v]", fi.GetName(), fileURL, req.Header.Get("Range"))
	req.Header.Del("If-Match")
	c.Proxy(w, req, fileURL)
}

func (c *DriveClient) GetFileURL(file drive.File) (string, error) {
	pickCode := file.(*FileInfo).PickCode
	cacheKey := fmt.Sprintf("url:%s", pickCode)
	if value, err := c.cache.Get(cacheKey); err == nil {
		return value.(string), nil
	}

	c.limiter.Wait(context.Background())
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
	c.limiter.Wait(context.Background())
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

func (c *DriveClient) MakeDir(dir string) error {
	c.limiter.Wait(context.Background())
	getDirIDResp, err := APIGetDirID(c.HttpClient, dir)
	if err != nil {
		return err
	}
	cid, _ := getDirIDResp.CategoryID.Int64()
	if cid != 0 {
		logrus.WithField("dir", dir).Infof("dir exists, ignore")
		return nil
	}

	dir = slashClean(dir)
	parentDir, name := path.Split(dir)
	getDirIDResp, err = APIGetDirID(c.HttpClient, parentDir)
	if err != nil {
		return err
	}

	pid := getDirIDResp.CategoryID.String()
	if pid == "0" && parentDir != "/" {
		return nil
	}
	resp, err := APIAddDir(c.HttpClient, pid, name)
	if err != nil {
		return err
	}
	if !resp.State {
		return fmt.Errorf("new dir fail, state is false")
	}

	c.flushDir(parentDir)
	return nil
}

func (c *DriveClient) MoveFile(srcPath string, dstPath string) error {
	logrus.Infof("move file, src: %s, dst: %s", srcPath, dstPath)

	c.limiter.Wait(context.Background())
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
					logrus.WithError(realErr).Warnf("proxy abort error")
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
	c.limiter.Wait(context.Background())
	c.reserveProxy.ServeHTTP(w, req)
}

func (c *DriveClient) flushDir(dir string) {
	dir = slashClean(dir)
	dir = strings.TrimRight(dir, "/")
	if len(dir) == 0 {
		dir = "/"
	}
	c.cache.Remove(fmt.Sprintf("files:%s", dir))
}

func slashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}
