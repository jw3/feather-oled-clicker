package main

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/tarm/serial"
	"io/ioutil"
	"log"
	"strings"
)

type Cfg struct {
	Uri   string
	Items [] struct {
		Id    string
		Title string
		Uri   string
		Body  string
	}
}

func main() {
	log.Println("starting clickerd")

	// connect to serial
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

	toSerialCh <- "hello"
	log.Print("waiting on ack: ")
	ackHello := <-fromSerialCh
	if ackHello != "HELLO" {
		log.Fatalf("ack failed: %v", ackHello)
	}
	log.Println("done!")

	// read def file
	cfg := Cfg{}
	cfgf, e := ioutil.ReadFile(".local/clickerd.conf")
	if e != nil {
		log.Fatalf("failed to read configuration: %v", e)
	}

	if e := yaml.Unmarshal([]byte(cfgf), &cfg); e != nil {
		log.Fatalf("failed to unmarshal configuration: %v", e)
	}

	// send list of items
	for i, it := range cfg.Items {
		if len(it.Uri) == 0 {
			it.Uri = cfg.Uri
		}
		println(i, it.Title, it.Uri, it.Body)
		toSerialCh <- fmt.Sprintf("%v. %v", i, it.Title)
	}

	// listen for selections
	for {
		e := <-fromSerialCh
		print(e)
	}
}
