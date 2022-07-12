package drive

import (
	"net/http"
	"time"
)

type File interface {
	GetName() string
	GetURL() string
	GetSize() int64
	GetUpdateTime() time.Time
	GetCreateTime() time.Time
	IsDir() bool
}

type DriveClient interface {
	GetFiles(path string) ([]File, error)
	GetFile(path string) (File, error)
	GetFileURL(file File) (string, error)
	Proxy(w http.ResponseWriter, req *http.Request, fileURL string)
}
