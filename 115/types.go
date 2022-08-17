package _115

import (
	"encoding/json"
	"time"
)

type DownloadURL struct {
	URL    string      `json:"url"`
	Client json.Number `json:"client"`
	Desc   string      `json:"desc"`
	OssID  string      `json:"oss_id"`
}

type DownloadInfo struct {
	FileName string      `json:"file_name"`
	FileSize json.Number `json:"file_size"`
	PickCode string      `json:"pick_code"`
	URL      DownloadURL `json:"url"`
}

type DownloadData map[string]*DownloadInfo

type APIBaseResp struct {
	State bool            `json:"state"`
	Msg   string          `json:"msg"`
	Errno json.Number     `json:"errno"`
	Data  json.RawMessage `json:"data,omitempty"`
}

type FileInfo struct {
	AreaID     json.Number `json:"aid"`
	CategoryID json.Number `json:"cid"`
	FileID     json.Number `json:"fid"`
	ParentID   json.Number `json:"pid"`

	Name     string      `json:"n"`
	Type     string      `json:"ico"`
	Size     json.Number `json:"s"`
	Sha1     string      `json:"sha"`
	PickCode string      `json:"pc"`

	CreateTime json.Number `json:"tp"`
	UpdateTime json.Number `json:"te"`
}

type APIGetFileInfoResp struct {
	State   bool        `json:"state"`
	Code    json.Number `json:"code"`
	Message string      `json:"message"`
	Data    []FileInfo  `json:"data"`
}

type APIGetFilesResp struct {
	AreaID     string      `json:"aid"`
	CategoryID json.Number `json:"cid"`
	Count      int64       `json:"count"`
	Cur        int64       `json:"cur"`
	Data       []FileInfo  `json:"data"`
	DataSource string      `json:"data_source"`
	ErrNo      int64       `json:"errNo"`
	Error      string      `json:"error"`
	Limit      int64       `json:"limit"`
	MaxSize    int64       `json:"max_size"`
	MinSize    int64       `json:"min_size"`
	Offset     int64       `json:"offset"`
	Order      string      `json:"order"`
	PageSize   int64       `json:"page_size"`
	Path       []FileInfo  `json:"path"`
	State      bool        `json:"state"`
	Suffix     string      `json:"suffix"`
}

type APIGetDirIDResp struct {
	ErrNo      json.Number `json:"errno"`
	Error      string      `json:"error"`
	CategoryID json.Number `json:"id"`
	IsPrivate  json.Number `json:"is_private"`
	State      bool        `json:"state"`
}

type APIDeleteFileResp struct {
	// ErrNo json.Number `json:"errno"`
	Error string `json:"error"`
	State bool   `json:"state"`
}

type APIAddDirResp struct {
	// ErrNo json.Number `json:"errno"`
	Error string `json:"error"`
	State bool   `json:"state"`
}

type APICopyFileResp struct {
	// ErrNo json.Number `json:"errno"`
	Error string `json:"error"`
	State bool   `json:"state"`
}

type APIMoveFileResp struct {
	// ErrNo json.Number `json:"errno"`
	Error string `json:"error"`
	State bool   `json:"state"`
}

type APIRenameFileResp struct {
	// ErrNo json.Number `json:"errno"`
	Error string `json:"error"`
	State bool   `json:"state"`
}

type APILoginCheckResp struct {
	ErrNo json.Number `json:"errno"`
	Error string      `json:"error"`
	Data  struct {
		Expire json.Number `json:"expire"`
		UserID json.Number `json:"user_id"`
		Link   string      `json:"link"`
	} `json:"data"`
}

func (f *FileInfo) GetName() string {
	return f.Name
}

func (f *FileInfo) GetSize() int64 {
	size, _ := f.Size.Int64()
	return size
}

func (f *FileInfo) GetUpdateTime() time.Time {
	updateTime, _ := f.UpdateTime.Int64()
	return time.Unix(updateTime, 0).UTC()
}

func (f *FileInfo) GetCreateTime() time.Time {
	updateTime, _ := f.UpdateTime.Int64()
	return time.Unix(updateTime, 0).UTC()
}

func (f *FileInfo) IsDir() bool {
	fid, _ := f.FileID.Int64()
	return fid == 0
}
