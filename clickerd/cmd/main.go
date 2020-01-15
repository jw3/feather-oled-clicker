package main

import (
	"bytes"
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/tarm/serial"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type Cfg struct {
	Command     string
	Items       [] Item
	Concurrency int
}

type Item struct {
	Id      string
	Title   string
	Modules [] struct {
		Id    string
		Model string
	}
}

func main() {
	log.Println("starting clickerd")

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
				}

				for _, s := range splits {
					fromSerialCh <- s
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
	cfg := Cfg{}
	cfgf, e := ioutil.ReadFile(".local/clickerd.conf")
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

	var last *Item = new(Item)

	// listen for selections
	for {
		e := <-fromSerialCh
		switch {

		// X=id
		case strings.HasPrefix(e, "X="):
			id, _ := strconv.ParseInt(strings.TrimPrefix(e, "X="), 10, 8)

			log.Printf("selected index %v", id)
			item := cfg.Items[id]
			binary, _ := exec.LookPath(cfg.Command)

			for _, m := range item.Modules {
				args := []string{m.Id, m.Model}
				for i := range last.Modules {
					// look for the previous model and send it over
					if last.Modules[i].Id == m.Id {
						args = append(args, last.Modules[i].Model)
						break
					}
				}

				print(cfg.Command)
				print(" ")
				for i := range args {
					print(args[i])
					print(" ")
				}
				println()

				cmd := exec.Command(binary, args...)
				var out bytes.Buffer
				var stderr bytes.Buffer
				cmd.Stdout = &out
				cmd.Stderr = &stderr

				e := cmd.Run()
				if e != nil {
					log.Printf("failed to run command: %v", e)
				}
			}

			last = &item
		}
	}
}
