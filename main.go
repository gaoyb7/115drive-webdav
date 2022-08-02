package main

import (
	"fmt"
	"net/http"

	_115 "github.com/gaoyb7/115drive-webdav/115"
	"github.com/gaoyb7/115drive-webdav/webdav"

	"github.com/gaoyb7/115drive-webdav/common/config"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	cfg = config.Config
)

func main() {
	logrus.SetReportCaller(true)
	_115.MustInit115DriveClient(cfg.Uid, cfg.Cid, cfg.Seid)
	startWebdavServer(cfg.Host, cfg.Port)
}

func startWebdavServer(host string, port int) {
	webdavHandler := webdav.Handler{
		DriveClient: _115.Get115DriveClient(),
		LockSystem:  webdav.NewMemLS(),
		Logger: func(req *http.Request, err error) {
			if err != nil {
				logrus.WithField("method", req.Method).WithField("path", req.URL.Path).Errorf("err: %v", err)
			}
		},
	}
	webdavHandleFunc := func(c *gin.Context) {
		webdavHandler.ServeHTTP(c.Writer, c.Request)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	dav := r.Group("", gin.BasicAuth(gin.Accounts{
		cfg.User: cfg.Password,
	}))
	dav.Any("/*path", webdavHandleFunc)
	dav.Handle("PROPFIND", "/*path", webdavHandleFunc)
	dav.Handle("MKCOL", "/*path", webdavHandleFunc)
	dav.Handle("LOCK", "/*path", webdavHandleFunc)
	dav.Handle("UNLOCK", "/*path", webdavHandleFunc)
	dav.Handle("PROPPATCH", "/*path", webdavHandleFunc)
	dav.Handle("COPY", "/*path", webdavHandleFunc)
	dav.Handle("MOVE", "/*path", webdavHandleFunc)

	if err := r.Run(fmt.Sprintf("%s:%d", host, port)); err != nil {
		logrus.Panic(err)
	}
}
