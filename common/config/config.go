package config

import (
	"encoding/json"
	"os"

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
	Replace  string `json:"replace"`
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
		Config.Replace = flag.Replace
	}
}

func load(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		pwd, _ := os.Getwd()
		logrus.WithField("pwd", pwd).WithField("filename", filename).Errorf("err: %v", err)
	}
	err = json.Unmarshal(data, &Config)
	if err != nil {
		logrus.WithField("filename", filename).Errorf("err: %v", err)
	}
}
