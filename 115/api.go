package _115

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gaoyb7/115drive-webdav/common"
	"github.com/go-resty/resty/v2"
)

const (
	UserAgent = "Mozilla/5.0 115Browser/23.9.3.2"

	APIURLGetFiles       = "https://webapi.115.com/files"
	APIURLGetDownloadURL = "https://proapi.115.com/app/chrome/downurl"
	APIURLGetDirID       = "https://webapi.115.com/files/getid"
	APIURLDeleteFile     = "https://webapi.115.com/rb/delete"
	APIURLAddDir         = "https://webapi.115.com/files/add"
	APIURLMoveFile       = "https://webapi.115.com/files/move"
	APIURLRenameFile     = "https://webapi.115.com/files/batch_rename"
	APIURLLoginCheck     = "https://passportapi.115.com/app/1.0/web/1.0/check/sso"
)

func APIGetFiles(client *resty.Client, cid string, pageSize int64, offset int64) (*APIGetFilesResp, error) {
	result := APIGetFilesResp{}
	_, err := client.R().
		SetQueryParams(map[string]string{
			"aid":              "1",
			"cid":              cid,
			"o":                "user_ptime",
			"asc":              "0",
			"offset":           strconv.FormatInt(offset, 10),
			"show_dir":         "1",
			"limit":            strconv.FormatInt(pageSize, 10),
			"snap":             "0",
			"record_open_time": "1",
			"format":           "json",
			"fc_mix":           "0",
		}).
		SetResult(&result).
		ForceContentType("application/json").
		Get(APIURLGetFiles)
	if err != nil {
		return nil, fmt.Errorf("api get files fail, err: %v", err)
	}

	return &result, nil
}

func APIGetDownloadURL(client *resty.Client, pickCode string) (*DownloadInfo, error) {
	key := GenerateKey()
	params, _ := json.Marshal(map[string]string{"pickcode": pickCode})

	result := APIBaseResp{}
	_, err := client.R().
		SetQueryParam("t", strconv.FormatInt(time.Now().Unix(), 10)).
		SetFormData(map[string]string{
			"data": string(Encode(params, key)),
		}).
		SetResult(&result).
		ForceContentType("application/json").
		Post(APIURLGetDownloadURL)
	if err != nil {
		return nil, fmt.Errorf("api get download url fail, err: %v", err)
	}

	var encodedData string
	if err = json.Unmarshal(result.Data, &encodedData); err != nil {
		return nil, fmt.Errorf("api get download url, call json.Unmarshal fail, body: %s", string(result.Data))
	}
	decodedData, err := Decode(encodedData, key)
	if err != nil {
		return nil, fmt.Errorf("api get download url, call Decode fail, err: %w", err)
	}

	resp := DownloadData{}
	if err := json.Unmarshal(decodedData, &resp); err != nil {
		return nil, fmt.Errorf("api get download url, call json.Unmarshal fail, body: %s", string(decodedData))
	}

	for _, info := range resp {
		fileSize, _ := info.FileSize.Int64()
		if fileSize == 0 {
			return nil, common.ErrNotFound
		}
		return info, nil
	}

	return nil, nil
}

func APIGetDirID(client *resty.Client, dir string) (*APIGetDirIDResp, error) {
	if strings.HasPrefix(dir, "/") {
		dir = dir[1:]
	}

	result := APIGetDirIDResp{}
	_, err := client.R().
		SetQueryParam("path", dir).
		SetResult(&result).
		ForceContentType("application/json").
		Get(APIURLGetDirID)
	if err != nil {
		return nil, fmt.Errorf("api get dir id fail, err: %v", err)
	}

	return &result, nil
}

func APIDeleteFile(client *resty.Client, fid string, pid string) (*APIDeleteFileResp, error) {
	result := APIDeleteFileResp{}
	_, err := client.R().
		SetFormData(map[string]string{
			"fid[0]":      fid,
			"pid":         pid,
			"ignore_warn": "1",
		}).
		SetResult(&result).
		ForceContentType("application/json").
		Post(APIURLDeleteFile)
	if err != nil {
		return nil, fmt.Errorf("api delete file fail, err: %v", err)
	}

	return &result, nil
}

func APIAddDir(client *resty.Client, pid string, cname string) (*APIAddDirResp, error) {
	result := APIAddDirResp{}
	_, err := client.R().
		SetFormData(map[string]string{
			"pid":   pid,
			"cname": cname,
		}).
		SetResult(&result).
		ForceContentType("application/json").
		Post(APIURLAddDir)
	if err != nil {
		return nil, fmt.Errorf("api add dir fail, err: %v", err)
	}

	return &result, nil
}

func APIMoveFile(client *resty.Client, fid string, pid string) (*APIMoveFileResp, error) {
	result := APIMoveFileResp{}
	_, err := client.R().
		SetFormData(map[string]string{
			"fid[0]": fid,
			"pid":    pid,
		}).
		SetResult(&result).
		ForceContentType("application/json").
		Post(APIURLMoveFile)
	if err != nil {
		return nil, fmt.Errorf("api move file fail, err: %v", err)
	}

	return &result, nil
}

func APIRenameFile(client *resty.Client, fid string, name string) (*APIRenameFileResp, error) {
	result := APIRenameFileResp{}
	_, err := client.R().
		SetFormData(map[string]string{
			fmt.Sprintf("files_new_name[%s]", fid): name,
		}).
		SetResult(&result).
		ForceContentType("application/json").
		Post(APIURLRenameFile)
	if err != nil {
		return nil, fmt.Errorf("api rename file fail, err: %v", err)
	}

	return &result, nil

}

func APILoginCheck(client *resty.Client) (int64, error) {
	result := APILoginCheckResp{}
	_, err := client.R().
		SetResult(&result).
		ForceContentType("application/json").
		Get(APIURLLoginCheck)
	if err != nil {
		return 0, fmt.Errorf("api login check fail, err: %v", err)
	}

	userID, _ := result.Data.UserID.Int64()
	return userID, nil
}
