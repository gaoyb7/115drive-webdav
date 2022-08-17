package drive

import (
	"net/http"
	"time"
)

type File interface {
	GetName() string
	GetSize() int64
	GetUpdateTime() time.Time
	GetCreateTime() time.Time
	IsDir() bool
}

type DriveClient interface {
	GetFiles(dir string) ([]File, error)
	GetFile(filePath string) (File, error)
	RemoveFile(filePath string) error
	MoveFile(srcPath string, dstPath string) error
	MakeDir(dir string) error
	ServeContent(w http.ResponseWriter, req *http.Request, fi File)
}
