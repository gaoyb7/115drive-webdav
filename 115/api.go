package _115

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gaoyb7/115drive-webdav/common"
)

const (
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36 115Browser/7.1.0"

	APIURLGetFiles       = "https://webapi.115.com/files"
	APIURLGetFileInfo    = "https://webapi.115.com/files/get_info"
	APIURLGetDownloadURL = "https://proapi.115.com/app/chrome/downurl"

	CookieDomain115   = ".115.com"
	CookieDomainAnxia = ".anxia.com"
)

func APIGetFiles(client *http.Client, cid string, pageSize int64, offset int64) (*APIGetFilesResp, error) {
	req, err := http.NewRequest(http.MethodGet, APIURLGetFiles, nil)
	if err != nil {
		return nil, fmt.Errorf("api get files, call http.NewRequest fail, err: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
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

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api get files, call client.Do fail, err: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("api get files, call ioutil.ReadAll fail, err: %w", err)
	}

	respData := APIGetFilesResp{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return nil, fmt.Errorf("api get files, call json.Unmarshal fail, err: %w", err)
	}
	return &respData, nil
}

func APIGetFileInfo(client *http.Client, fid string) (*APIGetFileInfoResp, error) {
	req, err := http.NewRequest(http.MethodGet, APIURLGetFileInfo, nil)
	if err != nil {
		return nil, fmt.Errorf("api get file info, call http.NewRequest fail, err: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	q := req.URL.Query()
	q.Add("file_id", fid)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api get file info, call client.Do fail, err: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("api get file info, call ioutil.ReadAll fail, err: %w", err)
	}

	respData := APIGetFileInfoResp{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return nil, fmt.Errorf("api get file info, call json.Unmarshal fail, err: %w", err)
	}
	return &respData, err
}

func APIGetDownloadURL(client *http.Client, pickCode string) (*DownloadInfo, error) {
	key := GenerateKey()
	params, _ := json.Marshal(map[string]string{"pickcode": pickCode})
	form := url.Values{}
	form.Set("data", Encode(params, key))
	data := strings.NewReader(form.Encode())
	req, err := http.NewRequest(http.MethodPost, APIURLGetDownloadURL, data)
	if err != nil {
		return nil, fmt.Errorf("api get download url, call http.NewRequest fail, err: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ContentLength = data.Size()
	q := req.URL.Query()
	q.Add("t", strconv.FormatInt(time.Now().Unix(), 10))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api get download url, call client.Do fail, err: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("api get download url, call ioutil.ReadAll fail, err: %w", err)
	}

	respData := APIBaseResp{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return nil, fmt.Errorf("api get download url, call json.Unmarshal fail, err: %w", err)
	}

	var resultData string
	if err = json.Unmarshal(respData.Data, &resultData); err != nil {
		return nil, fmt.Errorf("api get download url, call json.Unmarshal fail, err: %w", err)
	}

	data2, err := Decode(resultData, key)
	if err != nil {
		return nil, fmt.Errorf("api get download url, call Decode fail, err: %w", err)
	}
	result := DownloadData{}
	if err := json.Unmarshal(data2, &result); err != nil {
		return nil, fmt.Errorf("api get download url, call json.Unmarshal fail, err: %w", err)
	}

	for _, info := range result {
		fileSize, _ := info.FileSize.Int64()
		if fileSize == 0 {
			return nil, common.ErrNotFound
		}
		return info, nil
	}

	return nil, nil
}
