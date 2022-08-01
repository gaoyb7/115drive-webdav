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
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36 115Browser/23.9.3"

	APIURLGetFiles       = "https://webapi.115.com/files"
	APIURLGetFileInfo    = "https://webapi.115.com/files/get_info"
	APIURLGetDownloadURL = "https://proapi.115.com/app/chrome/downurl"
	APIURLGetDirID       = "https://webapi.115.com/files/getid"
	APIURLDeleteFile     = "https://webapi.115.com/rb/delete"
	APIURLLoginCheck     = "https://passportapi.115.com/app/1.0/web/1.0/check/sso"

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
	q.Add("o", "user_ptime")
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
		return nil, fmt.Errorf("api get files, call json.Unmarshal fail, body: %s", string(body))
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
		return nil, fmt.Errorf("api get file info, call json.Unmarshal fail, body: %s", string(body))
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
		return nil, fmt.Errorf("api get download url, call json.Unmarshal fail, body: %s", string(body))
	}

	var resultData string
	if err = json.Unmarshal(respData.Data, &resultData); err != nil {
		return nil, fmt.Errorf("api get download url, call json.Unmarshal fail, body: %s", string(respData.Data))
	}

	data2, err := Decode(resultData, key)
	if err != nil {
		return nil, fmt.Errorf("api get download url, call Decode fail, err: %w", err)
	}
	result := DownloadData{}
	if err := json.Unmarshal(data2, &result); err != nil {
		return nil, fmt.Errorf("api get download url, call json.Unmarshal fail, body: %s", string(data2))
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

func APIGetDirID(client *http.Client, dir string) (*APIGetDirIDResp, error) {
	if strings.HasPrefix(dir, "/") {
		dir = dir[1:]
	}

	req, err := http.NewRequest(http.MethodGet, APIURLGetDirID, nil)
	if err != nil {
		return nil, fmt.Errorf("api get dir id, call http.NewRequest fail, err: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	q := req.URL.Query()
	q.Add("path", dir)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api get dir id, call client.Do fail, err: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("api get dir id, call ioutil.ReadAll fail, err: %w", err)
	}

	respData := APIGetDirIDResp{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return nil, fmt.Errorf("api get dir id, call json.Unmarshal fail, body: %s", string(body))
	}
	return &respData, nil
}

func APIDeleteFile(client *http.Client, fid string, pid string) (*APIDeleteFileResp, error) {
	form := url.Values{}
	form.Set("fid[0]", fid)
	form.Set("pid", pid)
	form.Set("ignore_warn", "1")
	data := strings.NewReader(form.Encode())
	req, err := http.NewRequest(http.MethodPost, APIURLDeleteFile, data)
	if err != nil {
		return nil, fmt.Errorf("api delete file, call http.NewRequest fail, err: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ContentLength = data.Size()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api delete file, call client.Do fail, err: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("api delete file, call ioutil.ReadAll fail, err: %w", err)
	}

	respData := APIDeleteFileResp{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return nil, fmt.Errorf("api delete file, call json.Unmarshal fail, body: %s", string(body))
	}

	return &respData, nil
}

func APILoginCheck(client *http.Client) (int64, error) {
	req, err := http.NewRequest(http.MethodGet, APIURLLoginCheck, nil)
	if err != nil {
		return 0, fmt.Errorf("api login check, call http.NewRequest fail, err: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("api login check, call client.Do fail, err: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("api login check, call ioutil.ReadAll fail, err: %w", err)
	}

	respData := APILoginCheckResp{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return 0, fmt.Errorf("api login check, call json.Unmarshal fail, body: %s", string(body))
	}

	userID, _ := respData.Data.UserID.Int64()
	return userID, nil
}
