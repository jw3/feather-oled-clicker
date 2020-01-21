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

			binary, binEx := exec.LookPath(cfg.Command)
			if binEx != nil {
				log.Printf("command not found: %v", binEx)
				break
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
		}
	}
}
