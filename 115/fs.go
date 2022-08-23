package _115

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/webdav"
)

type WebdavFile struct {
	ctx      context.Context
	client   *DriveClientProxy
	info     os.FileInfo
	body     io.ReadCloser
	fullPath string
	offset   int64
}

func (f *WebdavFile) Close() error {
	if f.body != nil {
		return f.body.Close()
	}
	return nil
}

func (f *WebdavFile) Read(p []byte) (int, error) {
	// req = req.Clone(req.Context())
	targetURL, err := f.client.getURL(f.info)
	if err != nil {
		logrus.WithError(err).Errorf("call f.client.getURL fail")
		return 0, err
	}

	req := f.ctx.Value("req").(*http.Request)
	req = req.Clone(req.Context())
	req.RequestURI = ""
	u, _ := url.Parse(targetURL)
	req.URL = u
	req.Host = u.Host
	req.Header.Set("Referer", "https://115.com/")
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Host", req.Host)
	logrus.Infof("req: %+v", req)
	resp, err := f.client.httpClient.Do(req)
	if err != nil {
		logrus.WithError(err).Errorf("call f.client.httpClient.Do fail")
		return 0, err
	}
	logrus.Infof("resp: %+v", resp.StatusCode)
	f.body = resp.Body
	return f.body.Read(p)
}

func (f *WebdavFile) Write(p []byte) (int, error) {
	return 0, os.ErrPermission
}

func (f *WebdavFile) Readdir(count int) ([]os.FileInfo, error) {
	cid := f.info.(*WebdavFileInfo).categoryID
	logrus.Infof("webdav file read dir, cid: %v", cid)
	files, err := f.client.ReadDirByID(context.Background(), strconv.FormatInt(cid, 10))
	if err != nil {
		logrus.WithError(err).Errorf("call f.client.ReadDirByID fail")
		return nil, err
	}

	return files, nil
}

func (f *WebdavFile) Seek(offset int64, whence int) (int64, error) {
	size := f.info.Size()
	_ = http.ServeContent
	switch whence {
	case io.SeekStart:
		f.offset = 0
	case io.SeekEnd:
		f.offset = size
	}
	f.offset += offset
	return f.offset, nil
}

func (f *WebdavFile) Stat() (os.FileInfo, error) {
	return f.info, nil
}

type FileSystem struct {
	client *DriveClientProxy
}

func NewFileSystem(client *DriveClientProxy) *FileSystem {
	return &FileSystem{client}
}

func (f *FileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	logrus.Infof("file system mkdir, name: %v, perm: %v", name, perm)
	err := f.client.Mkdir(ctx, name)
	if err != nil {
		logrus.WithError(err).Errorf("cal f.client.Mkdir fail")
		return err
	}
	return nil
}

func (f *FileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	logrus.Infof("file system open file, name: %v, flag: %v, perm: %v", name, flag, perm)
	if flag&(os.O_SYNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_RDWR) != 0 {
		logrus.Errorf("flag not support")
		return nil, os.ErrInvalid
	}

	fi, err := f.client.Stat(ctx, name)
	if err != nil {
		logrus.WithError(err).Errorf("call f.client.Stat fail")
	}

	return &WebdavFile{
		ctx:      ctx,
		client:   f.client,
		info:     fi,
		fullPath: slashClean(name),
	}, nil
}

func (f *FileSystem) RemoveAll(ctx context.Context, name string) error {
	logrus.Infof("file system remove all, name: %v", name)
	err := f.client.Remove(ctx, name)
	if err != nil {
		logrus.WithError(err).Errorf("call f.client.Remove fail")
		return err
	}

	return nil
}

func (f *FileSystem) Rename(ctx context.Context, oldName, newName string) error {
	logrus.Infof("file system rename, old_name: %v, new_name: %v", oldName, newName)
	// TODO: impl
	return errors.New("not support")
}

func (f *FileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	logrus.Infof("file system stat, name: %v", name)
	fi, err := f.client.Stat(ctx, name)
	if err != nil {
		logrus.WithError(err).Errorf("call f.client.Stat fail")
		return nil, err
	}
	return fi, nil
}
