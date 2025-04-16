package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

type config struct {
	Uid      string `json:"uid"`
	Cid      string `json:"cid"`
	Seid     string `json:"seid"`
	Kid     string `json:"kid"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"pwd"`
}

var (
	Config config
)

var (
	cliConfig   = flag.String("config", "", "config file")
	cliUid      = flag.String("uid", "", "115 cookie uid")
	cliCid      = flag.String("cid", "", "115 cookie cid")
	cliSeid     = flag.String("seid", "", "115 cookie seid")
	cliKid     = flag.String("kid", "", "115 cookie kid")
	cliHost     = flag.String("host", "0.0.0.0", "webdav server host")
	cliPort     = flag.Int("port", 8080, "webdav server port")
	cliUser     = flag.String("user", "user", "webdav auth username")
	cliPassword = flag.String("pwd", "123456", "webdav auth password")
)

func init() {
	flag.Parse()
	if len(*cliConfig) > 0 {
		load(*cliConfig)
		return
	}

	Config.Uid = *cliUid
	Config.Cid = *cliCid
	Config.Seid = *cliSeid
	Config.Kid = *cliKid
	Config.Host = *cliHost
	Config.Port = *cliPort
	Config.User = *cliUser
	Config.Password = *cliPassword
}

func load(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		logrus.WithError(err).Panicf("call ioutil.ReadFile fail, filename: %v", filename)
	}

	err = json.Unmarshal(data, &Config)
	if err != nil {
		logrus.WithError(err).Panicf("call json.Unmarshal fail, filename: %v", filename)
	}
}
