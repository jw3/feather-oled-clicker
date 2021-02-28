package main

import (
	. "../common"
	"encoding/json"
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/tarm/serial"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {
	ppcUri, ok := os.LookupEnv(EnvVarPpcUri)
	if !ok {
		ppcUri = DefaultPpcUri
	}

	log.Println("starting clickerd to ppc @ ", ppcUri)

	// connect to serial; should read from arg
	c := &serial.Config{Name: "/dev/ttyACM0", Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	toSerialCh := make(chan string)
	fromSerialCh := make(chan string)

	go func() {
		var buf string
		for {
			r := make([]byte, 128)
			n, err := s.Read(r)
			if err != nil {
				log.Fatal(err)
			}
			buf += string(r[:n])

			if strings.Contains(buf, "\r\n") {
				splits := strings.Split(buf, "\r\n")
				if !strings.HasSuffix(buf, "\r\n") {
					buf = splits[len(splits)-1 ]
					splits = splits[:len(splits)-1]
				} else {
					// clear the buffer, it was fully consumed in splits
					buf = ""
				}

				for _, s := range splits {
					if len(s) > 0 {
						fromSerialCh <- s
					}
				}
			}
		}
	}()

	go func() {
		for {
			buf := <-toSerialCh
			_, err = s.Write([]byte(fmt.Sprintln(buf)))
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	// read def file
	// todo;; externalize this in env or arg
	cfg := Cfg{}
	cfgf, e := ioutil.ReadFile("/usr/local/etc/clickerd.conf")
	if e != nil {
		log.Fatalf("failed to read configuration: %v", e)
	}

	if e := yaml.UnmarshalStrict([]byte(cfgf), &cfg); e != nil {
		log.Fatalf("failed to unmarshal configuration: %v", e)
	}

	// initialize serial comm
	toSerialCh <- "hello"
	log.Print("waiting on ack: ")
	ackHello := <-fromSerialCh
	if ackHello != "HELLO" {
		log.Fatalf("ack failed: %v", ackHello)
	}
	log.Println("done!")

	// send list of items
	for _, it := range cfg.Items {
		log.Printf("sending model: %v", it.Title)
		toSerialCh <- fmt.Sprintf("%v", it.Title)
	}
	toSerialCh <- "READY"

	// listen for selections
	for {
		e := <-fromSerialCh
		switch {

		// X=id
		case strings.HasPrefix(e, "X="):
			id, _ := strconv.Atoi(strings.TrimPrefix(e, "X="))
			if !(id < len(cfg.Items)) {
				log.Printf("selected index out of bounds: %v", id)
				break
			}

			item := cfg.Items[id]
			log.Printf("selected model: %v", item.Title)

			_ = call(&item, ppcUri)
		}
	}
}

func call(item *Item, ppcUri string) error {
	for _, m := range item.Modules {
		movementCommand := "move" // todo;; externalize
		endpoint := fmt.Sprintf("http://%s/v1/devices/%s/%s", ppcUri, m.Id, movementCommand)

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