package main

import (
	. "../common"
	"fmt"
	"github.com/go-yaml/yaml"
	ppc "github.com/jw3/ppc/cli"
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

	cfg := ppc.NewConfiguration()

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

	println("ready! http://0.0.0.0:9001")

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


func call(item *Item, cfg *ppc.Config) error {
	for _, m := range item.Modules {
		movementCommand := "move" // todo;; externalize
		endpoint := fmt.Sprintf("http://%s/devices/%s/%s", cfg.ApiUri, m.Id, movementCommand)

		v := url.Values{}
		v.Set("args", m.Model)
		if _, e := http.PostForm(endpoint, v); e != nil {
			log.Printf("move failed for %v", m.Id)
			return e
		}
	}
	return nil
}
