package main

import (
	. "../common"
	"bytes"
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
	json, e := ioutil.ReadFile("/home/wassj/dev/code/jw3/feather-oled-clicker/clickerd/.local/clickerd.conf")
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

	binary, binEx := exec.LookPath(cfg.Command)
	if binEx != nil {
		log.Printf("item command not found: %v", binEx)
		return e
	}

	for _, m := range item.Modules {
		args := []string{m.Id, m.Model}
		log.Printf("%v %v %v", cfg.Command, m.Id, m.Model)

		cmd := exec.Command(binary, args...)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		e := cmd.Run()
		if e != nil {
			log.Printf("failed to run command: %v", e)
			log.Println(out.String())
			log.Println(stderr.String())
		}
	}

	return nil
}
