package flag

import (
	"flag"
)

var (
	CliUid      string
	CliCid      string
	CliSeid     string
	CliHost     string
	CliPort     int
	CliUser     string
	CliPassword string
	ConfigFile  string
)

func init() {
	flag.StringVar(&ConfigFile, "config", "", "config file")
	flag.StringVar(&CliUid, "uid", "", "115 cookie uid")
	flag.StringVar(&CliCid, "cid", "", "115 cookie cid")
	flag.StringVar(&CliSeid, "seid", "", "115 cookie seid")
	flag.StringVar(&CliHost, "host", "0.0.0.0", "webdav server host")
	flag.StringVar(&CliUser, "user", "user", "webdav auth username")
	flag.StringVar(&CliPassword, "pwd", "123456", "webdav auth password")
	flag.IntVar(&CliPort, "port", 8080, "webdav server port")
}

func init() {
	flag.Parse()
}
