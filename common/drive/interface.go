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
	GetFileURL(file File) (string, error)
	RemoveFile(filePath string) error
	NewDir(filepath string) error
	MoveFile(srcPath string, dstPath string) error
	Proxy(w http.ResponseWriter, req *http.Request, fileURL string)
}
