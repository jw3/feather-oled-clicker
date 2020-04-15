package common

type Cfg struct {
	Command     string
	Items       []Item
	Concurrency int
}

type Item struct {
	Title   string
	Modules []struct {
		Id    string
		Model string
	}
}

type CellZ struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

const (
	ClickerdConf = "/usr/local/etc/clickerd.conf"
)
