package _115

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	UserAgent = "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36 115Browser/7.1.0"

	APIURLGetFiles       = "https://webapi.115.com/files"
	APIURLGetDownloadURL = "https://proapi.115.com/app/chrome/downurl"
	APIURLGetFileInfo    = "https://webapi.115.com/files/get_info"
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

func slashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}

func splitPath(name string) []string {
	name = strings.Trim(name, "/")
	if len(name) == 0 {
		return nil
	}

	return strings.Split(name, "/")
}

func (c *DriveClient) GetFiles(path string) ([]FileInfo, error) {
	path = slashClean(path)
	cacheKey := fmt.Sprintf("files:%s", path)
	if value, err := c.cache.Get(cacheKey); err == nil {
		return value.([]FileInfo), nil
	}

	cid := "0"
	// TODO: get all pages
	resp, err := c.APIGetFiles(cid, 1000, 0)
	if err != nil {
		return nil, err
	}
	paths := splitPath(path)
	for idx := 0; idx < len(paths); idx++ {
		found := false
		for _, fileInfo := range resp.Data {
			if fileInfo.Name == paths[idx] {
				if !fileInfo.IsDir() {
					logrus.Errorf("not dir")
					return nil, ErrNotFound
				}
				found = true
				cid = fileInfo.CategoryID.String()
				break
			}
		}
		if !found {
			return nil, ErrNotFound
		}
		resp, err = c.APIGetFiles(cid, 1000, 0)
		if err != nil {
			return nil, err
		}
	}

	if err := c.cache.SetWithExpire(cacheKey, resp.Data, time.Minute*2); err != nil {
		logrus.WithError(err).Errorf("call c.cache.SetWithExpire fail, path: %s", path)
	}

	return resp.Data, nil
}

func (c *DriveClient) GetFile(filePath string) (*FileInfo, error) {
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
	for _, fileInfo := range files {
		if fileInfo.Name == fileName {
			return &fileInfo, nil
		}
	}

	return nil, ErrNotFound
}

func (c *DriveClient) GetURL(pickCode string) (string, error) {
	cacheKey := fmt.Sprintf("url:%s", pickCode)
	if value, err := c.cache.Get(cacheKey); err == nil {
		return value.(string), nil
	}

	info, err := c.APIGetDownloadURL(pickCode)
	if err != nil {
		return "", err
	}

	if err := c.cache.SetWithExpire(cacheKey, info.URL.URL, time.Minute*2); err != nil {
		logrus.WithError(err).Errorf("call c.cache.SetWithExpire fail, url: %s", info.URL.URL)
	}

	return info.URL.URL, nil
}

func (c *DriveClient) APIGetDownloadURL(pickCode string) (*DownloadInfo, error) {
	key := GenerateKey()
	params, _ := json.Marshal(map[string]string{"pickcode": pickCode})
	form := url.Values{}
	form.Set("data", Encode(params, key))
	data := strings.NewReader(form.Encode())
	req, err := http.NewRequest(http.MethodPost, APIURLGetDownloadURL, data)
	if err != nil {
		return nil, fmt.Errorf("call http.NewRequest fail, err: %w", err)
	}

	req.Header.Add("User-Agent", UserAgent)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.ContentLength = data.Size()
	q := req.URL.Query()
	q.Add("t", strconv.FormatInt(time.Now().Unix(), 10))
	req.URL.RawQuery = q.Encode()

	resp, err := c.webClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call client.Do fail, err: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("call ioutil.ReadAll fail, err: %w", err)
	}

	respData := APIBaseResp{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return nil, fmt.Errorf("call json.Unmarshal fail, err: %w", err)
	}

	var resultData string
	if err = json.Unmarshal(respData.Data, &resultData); err != nil {
		return nil, fmt.Errorf("call json.Unmarshal fail, err: %w", err)
	}

	data2, err := Decode(resultData, key)
	if err != nil {
		return nil, fmt.Errorf("call Decode fail, err: %w", err)
	}
	result := DownloadData{}
	if err := json.Unmarshal(data2, &result); err != nil {
		return nil, fmt.Errorf("call json.Unmarshal fail, err: %w", err)
	}

	for _, info := range result {
		fileSize, _ := info.FileSize.Int64()
		if fileSize == 0 {
			return nil, ErrNotFound
		}
		return info, nil
	}

	return nil, nil
}

func (c *DriveClient) APIGetFileInfo(fid string) (*APIGetFileInfoResp, error) {
	req, err := http.NewRequest(http.MethodGet, APIURLGetFileInfo, nil)
	if err != nil {
		return nil, fmt.Errorf("call http.NewRequest fail, err: %w", err)
	}

	req.Header.Add("User-Agent", UserAgent)
	q := req.URL.Query()
	q.Add("file_id", fid)
	req.URL.RawQuery = q.Encode()

	resp, err := c.webClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call c.webClient.Do fail, err: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("call ioutil.ReadAll fail, err: %w", err)
	}

	respData := APIGetFileInfoResp{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return nil, fmt.Errorf("call json.Unmarshal fail, err: %w", err)
	}
	return &respData, err
}

func (c *DriveClient) APIGetFiles(cid string, pageSize int64, offset int64) (*APIGetFilesResp, error) {
	req, err := http.NewRequest(http.MethodGet, APIURLGetFiles, nil)
	if err != nil {
		return nil, fmt.Errorf("call http.NewRequest fail, err: %w", err)
	}

	req.Header.Add("User-Agent", UserAgent)
	q := req.URL.Query()
	q.Add("aid", "1")
	q.Add("cid", cid)
	q.Add("o", "file_name")
	q.Add("asc", "0")
	q.Add("offset", strconv.FormatInt(offset, 10))
	q.Add("show_dir", "1")
	q.Add("limit", strconv.FormatInt(pageSize, 10))
	q.Add("snap", "0")
	q.Add("record_open_time", "1")
	q.Add("format", "json")
	q.Add("fc_mix", "0")
	req.URL.RawQuery = q.Encode()

	resp, err := c.webClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call c.webClient.Do fail, err: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("call ioutil.ReadAll fail, err: %w", err)
	}

	respData := APIGetFilesResp{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return nil, fmt.Errorf("call json.Unmarshal fail, err: %w", err)
	}
	return &respData, nil
}
