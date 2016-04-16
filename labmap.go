package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
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

type Reply struct {
	Status   string      `json:"status"`
	Error    string      `json:"error,omitempty"`
	Response interface{} `json:"response,omitempty"`
}

func (r *Reply) Reply(w http.ResponseWriter) {
	b, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func (r *Reply) Success(w http.ResponseWriter, resp interface{}) {
	r.Status = "Success"
	r.Response = resp
	r.Reply(w)
}

func (r *Reply) Failed(w http.ResponseWriter, errstr string) {
	r.Status = "Failed"
	r.Error = errstr
	r.Reply(w)
}

var (
	machines []string
	cabinet  map[string]*Cabinet
	lock     sync.Mutex
)

func serveCabinets(w http.ResponseWriter, r *http.Request) {
	var rpy Reply

	lock.Lock()
	defer lock.Unlock()

	var val interface{}

	machine := r.URL.Path

	if machine != "" {
		c, ok := cabinet[machine]
		if !ok {
			rpy.Failed(w, http.StatusText(http.StatusNotFound))
			return
		}
		val = c
	} else {
		val = cabinet
	}

	rpy.Success(w, val)
}

func serveMachines(w http.ResponseWriter, r *http.Request) {
	var rpy Reply

	lock.Lock()
	defer lock.Unlock()

	var m struct {
		Machines []string `json:"machines"`
	}

	m.Machines = machines

	rpy.Success(w, m)
}

func updateMap(words []string) {
	machine := words[0]
	cab := words[1]
	pos := strings.TrimPrefix(words[2], "pos")
	outlet := strings.TrimPrefix(words[5], "pdu")
	com1 := strings.TrimPrefix(words[3], "com1-")
	com2 := strings.TrimPrefix(words[4], "com2-")
	kvm := ""
	if len(words) == 7 {
		kvm = strings.TrimPrefix(words[6], "kvm")
	}

	c := &Cabinet{
		VTM0:     machine + "-vtm0",
		VTM1:     machine + "-vtm1",
		Cabinet:  strings.TrimPrefix(cab, "lnx"),
		Position: pos,
		Outlet:   outlet,
		PDU0:     cab + "-pdu0",
		PDU1:     cab + "-pdu1",
	}

	if com1 == "yes" {
		c.COM1 = "telnet " + cab + "-debug 100" + pos + "1"
	}

	if com2 == "yes" {
		c.COM2 = "telnet " + cab + "-debug 100" + pos + "2"
	}

	if kvm != "" {
		c.KVM = "lnx" + kvm + "-kvm"
	}

	cabinet[machine] = c
	machines = append(machines, machine)
}

func readmap(mapfile string) error {
	file, err := os.Open(mapfile)
	if err != nil {
		return err
	}
	defer file.Close()

	machines = make([]string, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		updateMap(strings.Fields(scanner.Text()))
	}

	return nil
}

func scan(mapfile string, refresh int) {
	for {
		lock.Lock()

		if err := readmap(mapfile); err != nil {
			log.Println("readmap", err)
		}

		log.Println("scan complete")

		lock.Unlock()

		if refresh == 0 {
			return
		}

		time.Sleep(time.Minute * time.Duration(refresh))
	}
}

const (
	CabinetBase = "/v1/cabinet/"
	MachineBase = "/v1/machines/"
)

func main() {
	var (
		port    = flag.String("port", "8889", "http port number")
		labmap  = flag.String("map", "lab.map", "lab configuration map")
		refresh = flag.Int("refresh", 60, "Time between map refresh scans")
	)

	flag.Parse()

	machines = make([]string, 0)
	cabinet = make(map[string]*Cabinet)

	go scan(*labmap, *refresh)

	log.Println("listening on port", *port)

	mux := http.NewServeMux()
	mux.Handle(MachineBase, http.StripPrefix(MachineBase, http.HandlerFunc(serveMachines)))
	mux.Handle(CabinetBase, http.StripPrefix(CabinetBase, http.HandlerFunc(serveCabinets)))

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
