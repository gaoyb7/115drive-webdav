package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gaoyb7/115drive-webdav/common/flag"
	"github.com/sirupsen/logrus"
)

type config struct {
	Uid      string `json:"uid"`
	Cid      string `json:"cid"`
	Seid     string `json:"seid"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"pwd"`
}

var Config config

func init() {
	if flag.ConfigFile != "" {
		load(flag.ConfigFile)
	} else {
		Config.Uid = flag.CliUid
		Config.Cid = flag.CliCid
		Config.Seid = flag.CliSeid
		Config.Host = flag.CliHost
		Config.Port = flag.CliPort
		Config.User = flag.CliUser
		Config.Password = flag.CliPassword
	}
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
