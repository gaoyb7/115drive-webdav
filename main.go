package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"

	_115 "github.com/gaoyb7/115drive-webdav/115"
	"github.com/gaoyb7/115drive-webdav/webdav"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	webdavHandler webdav.Handler

	cliUid      = flag.String("uid", "", "115 cookie uid")
	cliCid      = flag.String("cid", "", "115 cookie cid")
	cliSeid     = flag.String("seid", "", "115 cookie seid")
	cliHost     = flag.String("host", "0.0.0.0", "webdav server host")
	cliPort     = flag.Int("port", 8080, "webdav server port")
	cliUser     = flag.String("user", "user", "webdav auth username")
	cliPassword = flag.String("pwd", "123456", "webdav auth password")
)

func main() {
	flag.Parse()
	logrus.SetReportCaller(true)
	_115.MustInit115DriveClient(*cliUid, *cliCid, *cliSeid)
	startWebdavServer(*cliHost, *cliPort)
}

func startWebdavServer(host string, port int) {
	prefix := "/dav"
	webdavHandler = webdav.Handler{
		ServerHost:  host,
		ServerPort:  port,
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
		defer func() {
			if err := recover(); err != nil {
				if realErr, ok := err.(error); ok {
					if errors.Is(realErr, http.ErrAbortHandler) {
						return
					}
				}
				logrus.Errorf("panic: %v", err)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		_115.Get115DriveClient().Proxy(c.Writer, c.Request)
	})

	r.GET("/", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	if err := r.Run(fmt.Sprintf("%s:%d", host, port)); err != nil {
		logrus.Panic(err)
	}
}

func webdavHandle(c *gin.Context) {
	webdavHandler.ServeHTTP(c.Writer, c.Request)
}
