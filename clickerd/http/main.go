package main

import (
	. "../common"
	"fmt"
	"github.com/go-yaml/yaml"
	mycli "github.com/jw3/ppc/cli"
	"github.com/xujiajun/gorouter"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)


func main() {
	clickerConf, e := parseClickerConf()
	if e != nil {
		log.Fatalf("failed to read clicker configuration: %v", e)
	}

	cfg := mycli.NewConfiguration()

	items := make(chan Item)
	mux := gorouter.New()
	http.HandleFunc("/health", handleHealth)
	mux.POST("/click/:model", func(w http.ResponseWriter, r *http.Request) {
		sid := gorouter.GetParam(r, "model")
		id, e := strconv.Atoi(sid)
		if e != nil {
			log.Printf("Invalid numeric item id: %v", sid)
		}

		iid := id - 1
		if !(iid < len(clickerConf.Items)) {
			log.Printf("selected item id is out of bounds: %v", iid)
		}

		item := clickerConf.Items[iid]
		items <- item
	})


	go func() {
		for {
			item := <-items
			log.Printf("selected model: %v", item.Title)
			call(&item, cfg)
		}
	}()

	http.Handle("/", mux)
	log.Fatal(http.ListenAndServe(":9001", nil))
}


func handleHealth(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
}


func parseClickerConf() (Cfg, error) {
	cfg := Cfg{}
	json, e := ioutil.ReadFile("/usr/local/etc/clickerd.conf")
	if e == nil {
		e = yaml.UnmarshalStrict(json, &cfg)
	}
	return cfg, e
}


func call(item *Item, cfg *mycli.Config) error {
	for _, m := range item.Modules {
		uri := fmt.Sprintf("http://%s/devices/%s/", cfg.ApiUri, m.Id)

		if _, e := http.PostForm(uri+"cancel", url.Values{}); e != nil {
			log.Printf("cancel failed for %v", m.Id)
			return e
		}

		v := url.Values{}
		v.Set("args", m.Model)
		if _, e := http.PostForm(uri+"addNodes", v); e != nil {
			log.Printf("addNodes failed for %v", m.Id)
			return e
		}

		if _, e := http.PostForm(uri+"align", url.Values{}); e != nil {
			log.Printf("align failed for %v", m.Id)
			return e
		}
	}
	return nil
}