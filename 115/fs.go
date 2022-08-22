package _115

import (
	"context"
	"errors"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/webdav"
)

type WebdavFile struct {
	client   *DriveClientProxy
	info     os.FileInfo
	readPos  int64
	writePos int64
}

func (f *WebdavFile) Close() error {
	f.readPos = 0
	f.writePos = 0
	return nil
}

func (f *WebdavFile) Read(p []byte) (int, error) {
	// TODO: impl
	return 0, nil
}

func (f *WebdavFile) Write(p []byte) (int, error) {
	return 0, os.ErrPermission
}

func (f *WebdavFile) Readdir(count int) ([]os.FileInfo, error) {
	cid := f.info.(*WebdavFileInfo).categoryID
	files, err := f.client.ReadDirByID(context.Background(), strconv.FormatInt(cid, 10))
	if err != nil {
		logrus.WithError(err).Errorf("call f.client.ReadDirByID fail")
		return nil, err
	}

	return files, nil
}

func (f *WebdavFile) Seek(offset int64, whence int) (int64, error) {
	// TODO: impl
	return 0, nil
}

func (f *WebdavFile) Stat() (os.FileInfo, error) {
	return f.info, nil
}

type FileSystem struct {
	client *DriveClientProxy
}

func (f *FileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	err := f.client.Mkdir(ctx, name)
	if err != nil {
		logrus.WithError(err).Errorf("cal f.client.Mkdir fail")
		return err
	}
	return nil
}

func (f *FileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	if flag&(os.O_SYNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_RDWR) != 0 {
		logrus.Errorf("flag not support")
		return nil, os.ErrInvalid
	}

	fi, err := f.client.Stat(ctx, name)
	if err != nil {
		logrus.WithError(err).Errorf("call f.client.Stat fail")
	}

	return &WebdavFile{
		client: f.client,
		info:   fi,
	}, nil
}

func (f *FileSystem) RemoveAll(ctx context.Context, name string) error {
	err := f.client.Remove(ctx, name)
	if err != nil {
		logrus.WithError(err).Errorf("call f.client.Remove fail")
		return err
	}

	return nil
}

func (f *FileSystem) Rename(ctx context.Context, oldName, newName string) error {
	// TODO: impl
	return errors.New("not support")
}

func (f *FileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	fi, err := f.client.Stat(ctx, name)
	if err != nil {
		logrus.WithError(err).Errorf("call f.client.Stat fail")
		return nil, err
	}
	return fi, nil
}
