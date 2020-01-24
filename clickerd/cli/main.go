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

func main() {
	app := cli.NewApp()

	app.Commands = []*cli.Command{
		{
			Name:        "list",
			Usage:       "List items",
			UsageText:   "clicker list [options]",
			Description: "List all available items.",
			ArgsUsage:   "[options]",
			Action:      list,
		},
		{
			Name:        "show",
			Usage:       "Show Item",
			UsageText:   "clicker show [options]",
			Description: "Show model of item.",
			ArgsUsage:   "[options]",
			Action:      show,
		},
		{
			Name:        "click",
			Usage:       "Click an item",
			UsageText:   "clicker click [options] <item-id>",
			Description: "Click an item.",
			ArgsUsage:   "[options] <item-id>",
			Action:      click,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func parseClickerConf() (Cfg, error) {
	cfg := Cfg{}
	json, e := ioutil.ReadFile("./clickerd.conf")
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
		fmt.Printf("\t%v. %v\n", idx + 1, item.Title)
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
	fmt.Printf("\t%v. %v\n", id + 1, item.Title)
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
		uri := fmt.Sprintf("http://%s/devices/%s/", "localhost:9000/v1", m.Id)

		if _, e := http.PostForm(uri + "cancel", url.Values{}); e != nil {
			log.Printf( "cancel failed for %v", m.Id)
			return e
		}

		v := url.Values{}
		v.Set("args", m.Model)
		if _, e := http.PostForm(uri + "addNodes", v); e != nil {
			log.Printf( "cancel failed for %v", m.Id)
			return e
		}

		if _, e := http.PostForm(uri + "align", url.Values{}); e != nil {
			log.Printf( "cancel failed for %v", m.Id)
			return e
		}
	}
	return nil
}
