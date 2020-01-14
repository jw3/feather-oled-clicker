package main

import (
	"github.com/tarm/serial"
	"log"
)

type Cfg struct {
	Items [] struct {
		Id    string
		Title string
		Uri   string
		Body  string
	}
}

func main() {
	log.Println("starting clickerd")

	c := &serial.Config{Name: "/dev/ttyACM0", Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	toSerialCh := make(chan string)
	fromSerialCh := make(chan string)

	go func() {
		buf := make([]byte, 1024)
		for {
			_, err := s.Read(buf)
			if err != nil {
				log.Fatal(err)
			}

			fromSerialCh <- string(buf)
		}
	}()

	go func() {
		for {
			buf := <-toSerialCh
			_, err = s.Write([]byte(buf))
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	for {
		e := <-fromSerialCh
		print(string(e))
	}
}
