package main

import (
	. "../common"
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var cloudHost string
var cloudPort int
var configFile string

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

	if !(id < len(cfg.Items)) {
		log.Printf("selected item id is out of bounds: %v", id)
		return e
	}

	item := cfg.Items[id]
	log.Printf("selected model: %v", item.Title)

	return call(&item)
}

func call(item *Item) error {
	for _, m := range item.Modules {
		uri := fmt.Sprintf("http://%s:%v/v1/devices/%s/", cloudHost, cloudPort, m.Id)

		if _, e := http.PostForm(uri+"cancel", url.Values{}); e != nil {
			log.Printf("cancel failed for %v", m.Id)
			return e
		}

		v := url.Values{}
		v.Set("args", m.Model)
		if _, e := http.PostForm(uri+"addNodes", v); e != nil {
			log.Printf("cancel failed for %v", m.Id)
			return e
		}

		if _, e := http.PostForm(uri+"align", url.Values{}); e != nil {
			log.Printf("cancel failed for %v", m.Id)
			return e
		}
	}
	return nil
}
