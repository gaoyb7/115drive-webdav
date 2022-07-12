// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package webdav provides a WebDAV server implementation.
package webdav // import "golang.org/x/net/webdav"

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	_115 "github.com/gaoyb7/115drive-webdav/115"
	"github.com/gaoyb7/115drive-webdav/common"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	// ServerHost
	ServerHost string
	// ServerPort
	ServerPort int
	// Prefix is the URL path prefix to strip from WebDAV resource paths.
	Prefix string
	// DriveClient is 115 drive client.
	DriveClient *_115.DriveClient
	// Logger is an optional error logger. If non-nil, it will be called
	// for all HTTP requests.
	Logger func(*http.Request, error)
}

func (h *Handler) stripPrefix(p string) (string, int, error) {
	if h.Prefix == "" {
		return p, http.StatusOK, nil
	}
	if r := strings.TrimPrefix(p, h.Prefix); len(r) < len(p) {
		return r, http.StatusOK, nil
	}
	return p, http.StatusNotFound, errPrefixMismatch
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status, err := http.StatusBadRequest, errUnsupportedMethod
	switch r.Method {
	case "OPTIONS":
		status, err = h.handleOptions(w, r)
	case "GET", "HEAD", "POST":
		status, err = h.handleGetHeadPost(w, r)
	case "DELETE":
		status, err = http.StatusMethodNotAllowed, errUnsupportedMethod
	case "PUT":
		status, err = http.StatusMethodNotAllowed, errUnsupportedMethod
	case "MKCOL":
		status, err = http.StatusMethodNotAllowed, errUnsupportedMethod
	case "COPY", "MOVE":
		status, err = http.StatusMethodNotAllowed, errUnsupportedMethod
	case "LOCK":
		status, err = http.StatusMethodNotAllowed, errUnsupportedMethod
	case "UNLOCK":
		status, err = http.StatusMethodNotAllowed, errUnsupportedMethod
	case "PROPFIND":
		status, err = h.handlePropfind(w, r)
	case "PROPPATCH":
		status, err = http.StatusMethodNotAllowed, errUnsupportedMethod
	}

	if status != 0 {
		w.WriteHeader(status)
		if status != http.StatusNoContent {
			w.Write([]byte(StatusText(status)))
		}
	}
	if h.Logger != nil && err != nil {
		h.Logger(r, err)
	}
}

func (h *Handler) handleOptions(w http.ResponseWriter, r *http.Request) (status int, err error) {
	reqPath, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		return status, err
	}
	// allow := "OPTIONS, LOCK, PUT, MKCOL"
	allow := "OPTIONS"
	if fi, err := h.DriveClient.GetFile(reqPath); err == nil {
		if fi.IsDir() {
			// allow = "OPTIONS, LOCK, DELETE, PROPPATCH, COPY, MOVE, UNLOCK, PROPFIND"
			allow = "OPTIONS, PROPFIND"
		} else {
			// allow = "OPTIONS, LOCK, GET, HEAD, POST, DELETE, PROPPATCH, COPY, MOVE, UNLOCK, PROPFIND, PUT"
			allow = "OPTIONS, GET, HEAD, POST, PROPFIND"
		}
	} else {
		if !errors.Is(err, common.ErrNotFound) {
			logrus.WithError(err).Errorf("handleOptions, call h.DriveClient.GetFile fail, req_path: %s", reqPath)
		}
	}
	w.Header().Set("Allow", allow)
	// http://www.webdav.org/specs/rfc4918.html#dav.compliance.classes
	w.Header().Set("DAV", "1, 2")
	// http://msdn.microsoft.com/en-au/library/cc250217.aspx
	w.Header().Set("MS-Author-Via", "DAV")
	return 0, nil
}

func (h *Handler) handleGetHeadPost(w http.ResponseWriter, r *http.Request) (status int, err error) {
	reqPath, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		return status, err
	}

	fi, err := h.DriveClient.GetFile(reqPath)
	if err != nil {
		logrus.WithError(err).Errorf("handleGetHeadPost, call h.DriveClient.GetFile fail, req_path: %v", reqPath)
		return http.StatusNotFound, err
	}
	if fi.IsDir() {
		return http.StatusMethodNotAllowed, nil
	}

	etag, err := findETag(r.Context(), reqPath, fi)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	fileURL, err := h.DriveClient.GetURL(fi.PickCode)
	if err != nil {
		logrus.WithError(err).Errorf("handleGetHeadPost, call h.DriveClient.GetURL fail, pick_code: %s", fi.PickCode)
		return http.StatusInternalServerError, err
	}
	w.Header().Set("ETag", etag)
	h.DriveClient.Proxy(w, r, fileURL)
	return 0, nil
}

func (h *Handler) handlePropfind(w http.ResponseWriter, r *http.Request) (status int, err error) {
	reqPath, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		return status, err
	}

	ctx := r.Context()
	fi, err := h.DriveClient.GetFile(reqPath)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return http.StatusNotFound, err
		}
		logrus.WithError(err).Errorf("handlePropfind, call h.DriveClient.GetFile fail, req_path: %s", reqPath)
		return http.StatusMethodNotAllowed, err
	}
	depth := infiniteDepth
	if hdr := r.Header.Get("Depth"); hdr != "" {
		depth = parseDepth(hdr)
		if depth == invalidDepth {
			return http.StatusBadRequest, errInvalidDepth
		}
	}
	pf, status, err := readPropfind(r.Body)
	if err != nil {
		return status, err
	}

	mw := multistatusWriter{w: w}

	walkFn := func(reqPath string, info *_115.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var pstats []Propstat
		if pf.Propname != nil {
			pnames, err := propnames(ctx, info)
			if err != nil {
				return err
			}
			pstat := Propstat{Status: http.StatusOK}
			for _, xmlname := range pnames {
				pstat.Props = append(pstat.Props, Property{XMLName: xmlname})
			}
			pstats = append(pstats, pstat)
		} else if pf.Allprop != nil {
			pstats, err = allprop(ctx, info, pf.Prop)
		} else {
			pstats, err = props(ctx, info, pf.Prop)
		}
		if err != nil {
			return err
		}
		href := path.Join(h.Prefix, reqPath)
		if href != "/" && info.IsDir() {
			href += "/"
		}
		return mw.write(makePropstatResponse(href, pstats))
	}

	walkErr := walkFS(ctx, depth, reqPath, fi, walkFn)
	closeErr := mw.close()
	if walkErr != nil {
		return http.StatusInternalServerError, walkErr
	}
	if closeErr != nil {
		return http.StatusInternalServerError, closeErr
	}
	return 0, nil
}

func makePropstatResponse(href string, pstats []Propstat) *response {
	resp := response{
		Href:     []string{(&url.URL{Path: href}).EscapedPath()},
		Propstat: make([]propstat, 0, len(pstats)),
	}
	for _, p := range pstats {
		var xmlErr *xmlError
		if p.XMLError != "" {
			xmlErr = &xmlError{InnerXML: []byte(p.XMLError)}
		}
		resp.Propstat = append(resp.Propstat, propstat{
			Status:              fmt.Sprintf("HTTP/1.1 %d %s", p.Status, StatusText(p.Status)),
			Prop:                p.Props,
			ResponseDescription: p.ResponseDescription,
			Error:               xmlErr,
		})
	}
	return &resp
}

const (
	infiniteDepth = -1
	invalidDepth  = -2
)

// parseDepth maps the strings "0", "1" and "infinity" to 0, 1 and
// infiniteDepth. Parsing any other string returns invalidDepth.
//
// Different WebDAV methods have further constraints on valid depths:
//   - PROPFIND has no further restrictions, as per section 9.1.
//   - COPY accepts only "0" or "infinity", as per section 9.8.3.
//   - MOVE accepts only "infinity", as per section 9.9.2.
//   - LOCK accepts only "0" or "infinity", as per section 9.10.3.
//
// These constraints are enforced by the handleXxx methods.
func parseDepth(s string) int {
	switch s {
	case "0":
		return 0
	case "1":
		return 1
	case "infinity":
		return infiniteDepth
	}
	return invalidDepth
}

// http://www.webdav.org/specs/rfc4918.html#status.code.extensions.to.http11
const (
	StatusMulti               = 207
	StatusUnprocessableEntity = 422
	StatusLocked              = 423
	StatusFailedDependency    = 424
	StatusInsufficientStorage = 507
)

func StatusText(code int) string {
	switch code {
	case StatusMulti:
		return "Multi-Status"
	case StatusUnprocessableEntity:
		return "Unprocessable Entity"
	case StatusLocked:
		return "Locked"
	case StatusFailedDependency:
		return "Failed Dependency"
	case StatusInsufficientStorage:
		return "Insufficient Storage"
	}
	return http.StatusText(code)
}

var (
	errDestinationEqualsSource = errors.New("webdav: destination equals source")
	errDirectoryNotEmpty       = errors.New("webdav: directory not empty")
	errInvalidDepth            = errors.New("webdav: invalid depth")
	errInvalidDestination      = errors.New("webdav: invalid destination")
	errInvalidIfHeader         = errors.New("webdav: invalid If header")
	errInvalidLockInfo         = errors.New("webdav: invalid lock info")
	errInvalidLockToken        = errors.New("webdav: invalid lock token")
	errInvalidPropfind         = errors.New("webdav: invalid propfind")
	errInvalidProppatch        = errors.New("webdav: invalid proppatch")
	errInvalidResponse         = errors.New("webdav: invalid response")
	errInvalidTimeout          = errors.New("webdav: invalid timeout")
	errNoFileSystem            = errors.New("webdav: no file system")
	errNoLockSystem            = errors.New("webdav: no lock system")
	errNotADirectory           = errors.New("webdav: not a directory")
	errPrefixMismatch          = errors.New("webdav: prefix mismatch")
	errRecursionTooDeep        = errors.New("webdav: recursion too deep")
	errUnsupportedLockInfo     = errors.New("webdav: unsupported lock info")
	errUnsupportedMethod       = errors.New("webdav: unsupported method")
)
