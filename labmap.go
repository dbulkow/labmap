package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	kv "yin.mno.stratus.com/gogs/dbulkow/kv"
	"yin.mno.stratus.com/gogs/dbulkow/labmap/api"
)

var (
	machines []string
	cabinet  map[string]*api.Cabinet
	keystore kv.KV
	lock     sync.Mutex
)

func serveCabinets(w http.ResponseWriter, r *http.Request) {
	var rpy api.Reply

	lock.Lock()
	defer lock.Unlock()

	if err := readmap(); err != nil {
		log.Printf("serve machines readmap: %v", err)
	}

	machine := r.URL.Path

	if machine != "" {
		c, ok := cabinet[machine]
		if !ok {
			rpy.Failed(w, http.StatusText(http.StatusNotFound))
			return
		}
		rpy.Cabinet = c
	} else {
		rpy.Cabinets = cabinet
	}

	rpy.Success(w)
}

func serveMachines(w http.ResponseWriter, r *http.Request) {
	var rpy api.Reply

	lock.Lock()
	defer lock.Unlock()

	if err := readmap(); err != nil {
		log.Printf("serve machines readmap: %v", err)
	}

	rpy.Machines = machines
	rpy.Success(w)
}

type ComPort struct {
	Enabled  bool   `json:"enabled"`
	Speed    int    `json:"speed,omitempty"`
	Bits     int    `json:"bits,omitempty"`
	StopBits int    `json:"stopbits,omitempty"`
	Parity   string `json:"parity,omitempty"`
	Device   string `json:"device,omitempty"`
}

func (c *ComPort) String() string {
	if !c.Enabled {
		return "no"
	}

	return fmt.Sprintf("%d,%d,%d,%s:%s", c.Speed, c.Bits, c.StopBits, c.Parity, c.Device)
}

type Config struct {
	Name     string  `json:"name"`
	Cabinet  int     `json:"cabinet"`
	Position int     `json:"position"`
	COM1     ComPort `json:"com1"`
	COM2     ComPort `json:"com2"`
	PDU      int     `json:"pdu"`
	KVM      int     `json:"kvm"`
}

type byMachine []string

func (b byMachine) Len() int      { return len(b) }
func (b byMachine) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byMachine) Less(i, j int) bool {
	if strings.HasPrefix(b[i], "lin") && !strings.HasPrefix(b[j], "lin") {
		return true
	}
	if !strings.HasPrefix(b[i], "lin") && strings.HasPrefix(b[j], "lin") {
		return false
	}
	if strings.HasPrefix(b[i], "lin") && strings.HasPrefix(b[j], "lin") {
		if b[i][3] > b[j][3] {
			return true
		}
		if b[i][3] < b[j][3] {
			return false
		}
		return strings.Compare(b[i], b[j]) < 0
	}
	return strings.Compare(b[i], b[j]) < 0
}

func updateMap(val string) {
	cfg := &Config{}

	if err := json.Unmarshal([]byte(val), cfg); err != nil {
		return
	}

	c := &api.Cabinet{
		VTM0:     fmt.Sprintf("%s-vtm0", cfg.Name),
		VTM1:     fmt.Sprintf("%s-vtm1", cfg.Name),
		Cabinet:  fmt.Sprintf("%d", cfg.Cabinet),
		Position: fmt.Sprintf("%d", cfg.Position),
		Outlet:   fmt.Sprintf("%d", cfg.PDU),
		PDU0:     fmt.Sprintf("lnx%d-pdu0", cfg.Cabinet),
		PDU1:     fmt.Sprintf("lnx%d-pdu1", cfg.Cabinet),
		KVM:      fmt.Sprintf("lnx%d-kvm", cfg.KVM),
	}

	if cfg.COM1.Enabled {
		c.COM1 = fmt.Sprintf("telnet lnx%d-debug 100%d1", cfg.Cabinet, cfg.Position)
		c.Params1 = cfg.COM1.String()
		c.Serial1 = cfg.COM1.Device
	}

	if cfg.COM2.Enabled {
		c.COM2 = fmt.Sprintf("telnet lnx%d-debug 100%d2", cfg.Cabinet, cfg.Position)
		c.Params2 = cfg.COM2.String()
		c.Serial2 = cfg.COM2.Device
	}

	cabinet[cfg.Name] = c
	machines = append(machines, cfg.Name)

	sort.Sort(byMachine(machines))
}

func readmap() error {
	pairs, err := keystore.List("labconfig")
	if err != nil {
		return err
	}

	machines = make([]string, 0)

	for _, p := range pairs {
		updateMap(p.Val)
	}

	return nil
}

func main() {
	var port = flag.String("port", "8080", "http port number")

	flag.Parse()

	machines = make([]string, 0)
	cabinet = make(map[string]*api.Cabinet)
	keystore = &kv.Consul{}

	mux := http.NewServeMux()
	mux.Handle(api.MachineBase, http.StripPrefix(api.MachineBase, http.HandlerFunc(serveMachines)))
	mux.Handle(api.CabinetBase, http.StripPrefix(api.CabinetBase, http.HandlerFunc(serveCabinets)))

	srv := &http.Server{
		Addr:           ":" + *port,
		Handler:        mux,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSNextProto:   nil,
	}

	log.Println("listening on port", *port)

	log.Fatal(srv.ListenAndServe())
}
