package common

type Cfg struct {
	Command     string
	Items       [] Item
	Concurrency int
}

type Item struct {
	Title   string
	Modules [] struct {
		Id    string
		Model string
	}
}

const (
	ClickerdConf = "/usr/local/etc/clickerd.conf"
)
