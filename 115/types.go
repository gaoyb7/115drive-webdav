package _115

import "encoding/json"

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

	// IsStar json.Number `json:"m"`
	// Labels []*LabelInfo `json:"fl"`

	// MediaDuration describes duration in seconds for audio / video.
	// MediaDuration float64 `json:"play_long"`

	// Special fields for video
	// IsVideo         int `json:"iv"`
	// VideoDefinition int `json:"vdi"`
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
	// O          string      `json:"o"`
	Offset   int64      `json:"offset"`
	Order    string     `json:"order"`
	PageSize int64      `json:"page_size"`
	Path     []FileInfo `json:"path"`
	// O          string      `json:"o"`
	// RAll     int64      `json:"r_all"`
	// Star     int64  `json:"star"`
	State bool `json:"state"`
	// Stdir    int64  `json:"stdir"`
	Suffix string `json:"suffix"`
	// SysCount int64  `json:"sys_count"`
	// Type     int64  `json:"type"`
}

func (f *FileInfo) IsDir() bool {
	fid, _ := f.FileID.Int64()
	return fid == 0
}

func (f *FileInfo) GetName() string {
	return f.Name
}
