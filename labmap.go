package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Cabinet struct {
	VTM0     string `json:"vtm0"`
	VTM1     string `json:"vtm1"`
	Cabinet  string `json:"cabinet"`
	Position string `json:"position"`
	COM1     string `json:"com1"`
	COM2     string `json:"com2"`
	Outlet   string `json:"outlet"`
	KVM      string `json:"kvm"`
	PDU0     string `json:"pdu0"`
	PDU1     string `json:"pdu1"`
}

var (
	machines []string
	cabinet  map[string]*Cabinet
)

func serveCabinets(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(cabinet)
	if err != nil {
		// XXX internal error
	}

	fmt.Fprintf(w, string(b))
}

func serveMachines(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(machines)
	if err != nil {
		// XXX internal error
	}

	fmt.Fprintf(w, string(b))
}

func readmap(mapfile string) error {
	file, err := os.Open(mapfile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())

		machine := words[0]
		cab := words[1]
		pos := strings.TrimPrefix(words[2], "pos")
		outlet := strings.TrimPrefix(words[5], "pdu")

		c := &Cabinet{
			VTM0:     machine + "-vtm0",
			VTM1:     machine + "-vtm1",
			Cabinet:  strings.TrimPrefix(cab, "lnx"),
			Position: pos,
			Outlet:   outlet,
			PDU0:     cab + "-pdu0",
			PDU1:     cab + "-pdu0",
		}

		if strings.TrimPrefix(words[3], "com1-") == "yes" {
			c.COM1 = "telnet " + cab + "-debug 100" + pos + "1"
		}
		if strings.TrimPrefix(words[3], "com2-") == "yes" {
			c.COM2 = "telnet " + cab + "-debug 100" + pos + "2"
		}

		if len(words) == 7 {
			c.KVM = "lnx" + strings.TrimPrefix(words[6], "kvm") + "-kvm"
		}

		cabinet[machine] = c
		machines = append(machines, machine)
	}

	return nil
}

func main() {
	port := flag.String("port", "8889", "http port number")
	labmap := flag.String("map", "lab.map", "lab configuration map")

	flag.Parse()

	machines = make([]string, 0)
	cabinet = make(map[string]*Cabinet)

	if err := readmap(*labmap); err != nil {
		log.Fatal("readmap", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/machines/", serveMachines)
	mux.HandleFunc("/v1/cabinet/", serveCabinets)

	srv := &http.Server{
		Addr:           ":" + *port,
		Handler:        mux,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSNextProto:   nil,
	}

	log.Fatal(srv.ListenAndServe())
}
