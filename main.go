package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"

	_115 "github.com/gaoyb7/115drive-webdav/115"
	"github.com/gaoyb7/115drive-webdav/webdav"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	webdavHandler webdav.Handler

	cliPort     = flag.Int("port", 8080, "server port")
	cliUid      = flag.String("uid", "", "115 cookie uid")
	cliCid      = flag.String("cid", "", "115 cookie cid")
	cliSeid     = flag.String("seid", "", "115 cookie seid")
	cliUser     = flag.String("user", "user", "webdav auth username")
	cliPassword = flag.String("pwd", "123456", "webdav auth password")
)

func main() {
	flag.Parse()
	logrus.SetReportCaller(true)
	_115.MustInit115DriveClient(*cliUid, *cliCid, *cliSeid)
	startWebdavServer(*cliPort)
}

func startWebdavServer(port int) {
	prefix := "/dav"
	webdavHandler = webdav.Handler{
		Prefix:      prefix,
		LockSystem:  webdav.NewMemLS(),
		DriveClient: _115.Get115DriveClient(),
		Logger: func(req *http.Request, err error) {
			if err != nil {
				logrus.WithField("method", req.Method).WithField("path", req.URL.Path).Errorf("err: %v", err)
			}
		},
	}

	r := gin.Default()
	dav := r.Group(prefix, gin.BasicAuth(gin.Accounts{
		*cliUser: *cliPassword,
	}))
	dav.Any("", webdavHandle)
	dav.Any("/*path", webdavHandle)
	dav.Handle("PROPFIND", "/*path", webdavHandle)
	dav.Handle("MKCOL", "/*path", webdavHandle)
	dav.Handle("LOCK", "/*path", webdavHandle)
	dav.Handle("UNLOCK", "/*path", webdavHandle)
	dav.Handle("PROPPATCH", "/*path", webdavHandle)
	dav.Handle("COPY", "/*path", webdavHandle)
	dav.Handle("MOVE", "/*path", webdavHandle)

	r.Any("/proxy", func(c *gin.Context) {
		// get target url
		targetURL := c.Query("target")
		u, err := url.Parse(targetURL)
		if err != nil {
			logrus.WithError(err).Errorf("call url.Parse fail")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// proxy request
		proxyReq, _ := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
		for h, v := range c.Request.Header {
			proxyReq.Header[h] = v
		}
		proxyReq.Header.Set("Referer", "https://115.com/")
		proxyReq.Header.Set("User-Agent", _115.UserAgent)
		proxyReq.Header.Set("Host", u.Host)
		proxyResp, err := _115.Get115DriveClient().GetWebClient().Do(proxyReq)
		if err != nil {
			logrus.WithError(err).Errorf("proxy request fail")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		defer proxyResp.Body.Close()

		// proxy response
		contentType := proxyResp.Header.Get("Content-Type")
		extraHeaders := make(map[string]string)
		for h, v := range proxyResp.Header {
			extraHeaders[h] = v[0]
		}
		c.DataFromReader(proxyResp.StatusCode, proxyResp.ContentLength, contentType, proxyResp.Body, extraHeaders)
	})

	r.GET("/", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	if err := r.Run(fmt.Sprintf("0.0.0.0:%d", port)); err != nil {
		logrus.Panic(err)
	}
}

func webdavHandle(c *gin.Context) {
	webdavHandler.ServeHTTP(c.Writer, c.Request)
}
