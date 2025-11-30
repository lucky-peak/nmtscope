package config

var CONFIG = Config{
	Jcmd:      "jcmd",
	Pid:       -1,
	ReportDir: "/tmp/nmt",
	Port:      8088,
	Interval:  10,
	Retention: 60,
}

type Config struct {
	Jcmd      string `json:"jcmd"`
	Pid       int    `json:"pid"`
	ReportDir string `json:"report_dir"`
	Port      int    `json:"port"`
	Interval  int    `json:"interval"`
	Retention int    `json:"retention"`
}
