package main

import (
	. "../common"
	"encoding/json"
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var cloudHost = "localhost"
var cloudPort = 9000
var configFile string
var cycleRepeat bool
var cycleLength int64

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		&cli.PathFlag{
			Name:        "config",
			Value:       "clickerd.conf",
			Usage:       "Clickerd config file",
			Aliases:     []string{"c"},
			Destination: &configFile,
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:        "list",
			Usage:       "List items",
			UsageText:   "clicker list",
			Description: "List all available items.",
			Aliases:     []string{"ls"},
			Action:      list,
		},
		{
			Name:        "show",
			Usage:       "Show Item",
			UsageText:   "clicker show <item-id>",
			Description: "Show model of item.",
			Action:      show,
		},
		{
			Name:        "click",
			Usage:       "Click an item",
			UsageText:   "clicker click <item-id>",
			Description: "Click an item.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "cloud-host",
					Value:       "192.168.2.1",
					Usage:       "Cloud API Server Host",
					Aliases:     []string{"H"},
					Destination: &cloudHost,
				},
				&cli.IntFlag{
					Name:        "cloud-port",
					Value:       9000,
					Usage:       "Cloud API Server Port",
					Aliases:     []string{"P"},
					Destination: &cloudPort,
				}},
			Action: click,
		},
		{
			Name:        "cycle",
			Usage:       "Click each item in the list",
			UsageText:   "clicker cycle",
			Description: "Show model of item.",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "repeat",
					Value:       false,
					Usage:       "Repeat the list",
					Aliases:     []string{"r"},
					Destination: &cycleRepeat,
				},
				&cli.Int64Flag{
					Name:        "length",
					Value:       30,
					Usage:       "Length in seconds",
					Aliases:     []string{"l"},
					Destination: &cycleLength,
				}},
			Action: cycle,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func parseClickerConf() (Cfg, error) {
	cfg := Cfg{}
	json, e := ioutil.ReadFile(configFile)
	if e == nil {
		e = yaml.UnmarshalStrict(json, &cfg)
	}
	return cfg, e
}

func list(c *cli.Context) error {
	cfg, e := parseClickerConf()
	if e != nil {
		log.Printf("failed to read clicker configuration")
		return e
	}

	for idx, item := range cfg.Items {
		fmt.Printf("\t%v. %v\n", idx+1, item.Title)
	}

	return nil
}

func show(c *cli.Context) error {
	id, e := strconv.Atoi(c.Args().Get(0))
	if e != nil {
		log.Fatalf("Invalid numeric item id: %v", e)
		return e
	}

	cfg, e := parseClickerConf()
	if e != nil {
		log.Fatalf("failed to read clicker configuration: %v", e)
		return e
	}

	if !(id < len(cfg.Items)) {
		log.Fatalf("selected item id is out of bounds: %v", id)
		return e
	}

	item := cfg.Items[id]
	fmt.Printf("\t%v. %v\n", id+1, item.Title)
	for _, m := range item.Modules {
		fmt.Printf("\tid: %v\n", m.Id)
		fmt.Printf("\t    %v\n", m.Model)
	}

	return nil
}

func click(c *cli.Context) error {
	id, e := strconv.Atoi(c.Args().Get(0))
	if e != nil {
		log.Printf("Invalid numeric item id: %v", e)
	}

	cfg, e := parseClickerConf()
	if e != nil {
		log.Printf("failed to read clicker configuration: %v", e)
		return e
	}

	iid := id - 1
	if !(iid < len(cfg.Items)) {
		log.Printf("selected item id is out of bounds: %v", iid)
		return e
	}

	item := cfg.Items[iid]
	log.Printf("selected model: %v", item.Title)

	return call(&item)
}

func call(item *Item) error {
	for _, m := range item.Modules {
		movementCommand := "move" // todo;; externalize
		endpoint := fmt.Sprintf("http://%s:%v/v1/devices/%s/%s", cloudHost, cloudPort, m.Id, movementCommand)

		println(endpoint)
		cells := make([]CellZ, 1)
		e := json.Unmarshal([]byte(m.Model), &cells)
		if e != nil {
			log.Printf("failed to unmarshal model %v", m.Id)
			return e
		}

		for _, c := range cells {
			s, e := json.Marshal(c)
			if e != nil {
				log.Printf("failed to marshal model %v", m.Id)
				return e
			}

			v := url.Values{}
			v.Set("args", string(s))
			if _, e := http.PostForm(endpoint, v); e != nil {
				log.Printf("move failed for %v", m.Id)
				return e
			}
		}
	}
	return nil
}

func cycle(c *cli.Context) error {
	cfg, e := parseClickerConf()
	if e != nil {
		log.Printf("failed to read clicker configuration")
		return e
	}

	for {
		for idx, item := range cfg.Items {
			fmt.Printf("\t%v. %v\n", idx+1, item.Title)
			e = call(&item)
			time.Sleep(time.Duration(cycleLength) * time.Second)
		}
		if !cycleRepeat {
			break
		}
	}

	return nil
}