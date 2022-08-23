package _115

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gaoyb7/115drive-webdav/common"

	"github.com/bluele/gcache"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type DriveClientProxy struct {
	httpClient *http.Client
	cookieJar  *cookiejar.Jar
	cache      gcache.Cache
	limiter    *rate.Limiter
}

func NewDriveClientProxy(uid string, cid string, seid string) *DriveClientProxy {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	client := &DriveClientProxy{
		httpClient: &http.Client{Jar: cookieJar},
		cookieJar:  cookieJar,
		cache:      gcache.New(10000).LFU().Build(),
		limiter:    rate.NewLimiter(5, 10),
	}
	client.ImportCredential(uid, cid, seid)

	userID, err := APILoginCheck(client.httpClient)
	if err != nil {
		panic(err)
	}
	if userID <= 0 {
		panic("115 drive login fail")
	}
	logrus.Infof("115 drive login succ, user_id: %d", userID)

	return client
}

func (c *DriveClientProxy) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	logrus.Infof("drive client proxy stat, name: %v", name)
	name = slashClean(name)
	if name == "/" || len(name) == 0 {
		return toWebdavFileInfo(&FileInfo{CategoryID: "0"}), nil
	}

	name = strings.TrimRight(name, "/")
	dir, fileName := path.Split(name)

	files, err := c.ReadDir(ctx, dir)
	logrus.Infof("files: %+v", files)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.Name() == fileName {
			return file, nil
		}
	}

	return nil, common.ErrNotFound
}

func (c *DriveClientProxy) ReadDir(ctx context.Context, dir string) ([]os.FileInfo, error) {
	logrus.Infof("drive client proxy read dir, dir: %v", dir)
	dir = slashClean(dir)
	cacheKey := fmt.Sprintf("files:%s", dir)
	if value, err := c.cache.Get(cacheKey); err == nil {
		return value.([]os.FileInfo), nil
	}

	c.limiter.Wait(context.Background())
	getDirIDResp, err := APIGetDirID(c.httpClient, dir)
	if err != nil {
		return nil, err
	}
	cid := getDirIDResp.CategoryID.String()

	files, err := c.ReadDirByID(ctx, cid)
	if err != nil {
		return nil, err
	}

	if err := c.cache.SetWithExpire(cacheKey, files, time.Minute*2); err != nil {
		logrus.WithError(err).Errorf("call c.cache.SetWithExpire fail, dir: %s", dir)
	}

	return files, nil
}

func (c *DriveClientProxy) ReadDirByID(ctx context.Context, cid string) ([]os.FileInfo, error) {
	logrus.Infof("drive client proxy read dir by id, cid: %v", cid)
	pageSize := int64(1000)
	offset := int64(0)
	files := make([]os.FileInfo, 0)
	for {
		resp, err := APIGetFiles(c.httpClient, cid, pageSize, offset)
		if err != nil {
			return nil, err
		}

		for idx := range resp.Data {
			files = append(files, toWebdavFileInfo(&resp.Data[idx]))
		}

		offset = resp.Offset + pageSize
		if offset >= resp.Count {
			break
		}
	}

	return files, nil
}

func (c *DriveClientProxy) Remove(ctx context.Context, name string) error {
	logrus.Infof("drive client proxy remove, name: %v", name)
	c.limiter.Wait(ctx)
	fi, err := c.Stat(ctx, name)
	if err != nil {
		return err
	}
	fid := fi.(*WebdavFileInfo).fileID
	if fi.IsDir() {
		fid = fi.(*WebdavFileInfo).categoryID
	}

	pid := fi.(*WebdavFileInfo).parentID
	if err != nil {
		return err
	}

	resp, err := APIDeleteFile(c.httpClient, strconv.FormatInt(fid, 10), strconv.FormatInt(pid, 10))
	if err != nil {
		return err
	}

	if !resp.State {
		return fmt.Errorf("remove file fail, state is false")
	}

	name = slashClean(name)
	name = strings.TrimRight(name, "/")
	parentDir, _ := path.Split(name)
	c.flushDir(parentDir)

	return nil
}

func (c *DriveClientProxy) Mkdir(ctx context.Context, dir string) error {
	logrus.Infof("drive client proxy mkdir, dir: %v", dir)
	c.limiter.Wait(context.Background())
	getDirIDResp, err := APIGetDirID(c.httpClient, dir)
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
	getDirIDResp, err = APIGetDirID(c.httpClient, parentDir)
	if err != nil {
		return err
	}

	pid := getDirIDResp.CategoryID.String()
	if pid == "0" && parentDir != "/" {
		return nil
	}
	resp, err := APIAddDir(c.httpClient, pid, name)
	if err != nil {
		return err
	}
	if !resp.State {
		return fmt.Errorf("new dir fail, state is false")
	}

	c.flushDir(parentDir)
	return nil
}

func (c *DriveClientProxy) ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request, targetURL string) {
}

func (c *DriveClientProxy) ImportCredential(uid string, cid string, seid string) {
	cookies := map[string]string{
		"UID":  uid,
		"CID":  cid,
		"SEID": seid,
	}
	c.importCookies(CookieDomain115, "/", cookies)
}

func (c *DriveClientProxy) importCookies(domain string, path string, cookies map[string]string) {
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

func (c *DriveClientProxy) getURL(f os.FileInfo) (string, error) {
	pickCode := f.(*WebdavFileInfo).pickCode
	cacheKey := fmt.Sprintf("url:%s", pickCode)
	if value, err := c.cache.Get(cacheKey); err == nil {
		return value.(string), nil
	}

	c.limiter.Wait(context.Background())
	info, err := APIGetDownloadURL(c.httpClient, pickCode)
	if err != nil {
		return "", err
	}

	if err := c.cache.SetWithExpire(cacheKey, info.URL.URL, time.Minute*2); err != nil {
		logrus.WithError(err).Errorf("call c.cache.SetWithExpire fail, url: %s", info.URL.URL)
	}

	return info.URL.URL, nil
}

func (c *DriveClientProxy) flushDir(dir string) {
	dir = slashClean(dir)
	dir = strings.TrimRight(dir, "/")
	if len(dir) == 0 {
		dir = "/"
	}
	c.cache.Remove(fmt.Sprintf("files:%s", dir))
}

type WebdavFileInfo struct {
	os.FileInfo
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time

	fullPath   string
	categoryID int64
	fileID     int64
	parentID   int64
	sha1       string
	pickCode   string
}

func (f *WebdavFileInfo) Name() string       { return f.name }
func (f *WebdavFileInfo) Size() int64        { return f.size }
func (f *WebdavFileInfo) Mode() os.FileMode  { return f.mode }
func (f *WebdavFileInfo) ModTime() time.Time { return f.modTime }
func (f *WebdavFileInfo) IsDir() bool        { return f.mode.IsDir() }
func (f *WebdavFileInfo) Sys() interface{}   { return nil }
func (f *WebdavFileInfo) ContentType(ctx context.Context) (string, error) {
	if mimeType := mime.TypeByExtension(path.Ext(f.Name())); mimeType != "" {
		return mimeType, nil
	} else {
		return "application/octet-stream", nil
	}
}

func toWebdavFileInfo(info *FileInfo) *WebdavFileInfo {
	fm := os.FileMode(0)
	if info.IsDir() {
		fm = os.ModeDir
	}
	cid, _ := info.CategoryID.Int64()
	fid, _ := info.FileID.Int64()
	pid, _ := info.ParentID.Int64()
	return &WebdavFileInfo{
		name:       info.GetName(),
		size:       info.GetSize(),
		mode:       fm,
		modTime:    info.GetUpdateTime(),
		categoryID: cid,
		fileID:     fid,
		parentID:   pid,
		sha1:       info.Sha1,
		pickCode:   info.PickCode,
	}
}
