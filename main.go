package main

import (
	"flag"
	"fmt"
	"net/http"

	_115 "github.com/gaoyb7/115drive-webdav/115"
	"github.com/gaoyb7/115drive-webdav/webdav"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
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
		*cliUser: *cliPassword,
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
