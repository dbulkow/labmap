package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	CabinetBase = "/v1/cabinet/"
	MachineBase = "/v1/machines/"
)

type Cabinet struct {
	VTM0     string `json:"vtm0"`
	VTM1     string `json:"vtm1"`
	Cabinet  string `json:"cabinet"`
	Position string `json:"position"`
	COM1     string `json:"com1"`
	Serial1  string `json:"serial1"`
	Params1  string `json:"params1"`
	COM2     string `json:"com2"`
	Serial2  string `json:"serial2"`
	Params2  string `json:"params2"`
	Outlet   string `json:"outlet"`
	KVM      string `json:"kvm"`
	PDU0     string `json:"pdu0"`
	PDU1     string `json:"pdu1"`
}

type Reply struct {
	Status   string              `json:"status"`
	Error    string              `json:"error,omitempty"`
	Cabinet  *Cabinet            `json:"cabinet,omitempty"`
	Cabinets map[string]*Cabinet `json:"cabinets,omitempty"`
	Machines []string            `json:"machines,omitempty"`
}

func (r *Reply) Reply(w http.ResponseWriter) {
	b, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.Header().Set("Cache-control", "no-cache")
	w.Write(b)
}

func (r *Reply) Success(w http.ResponseWriter) {
	r.Status = "Success"
	r.Reply(w)
}

func (r *Reply) Failed(w http.ResponseWriter, errstr string) {
	r.Status = "Failed"
	r.Error = errstr
	r.Reply(w)
}

func getData(url, uri string) (*Reply, error) {
	client := &http.Client{Timeout: time.Second * 20}

	resp, err := client.Get(url + uri)
	if err != nil {
		return nil, fmt.Errorf("Get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %s", http.StatusText(resp.StatusCode))
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("readall: %v", err)
	}

	rpy := &Reply{}

	if err := json.Unmarshal(b, rpy); err != nil {
		return nil, fmt.Errorf("unmarshal: %v", err)
	}

	if rpy.Status != "Success" {
		return nil, fmt.Errorf("status: %s", rpy.Status)
	}

	return rpy, nil
}

func GetCabinet(url, machine string) (*Cabinet, error) {
	rpy, err := getData(url, CabinetBase)
	if err != nil {
		return nil, fmt.Errorf("get cabinet (%s%s): %v", url, CabinetBase, err)
	}

	return rpy.Cabinet, nil
}

func Cabinets(url string) (map[string]*Cabinet, error) {
	rpy, err := getData(url, CabinetBase)
	if err != nil {
		return nil, fmt.Errorf("get cabinets (%s%s): %v", url, CabinetBase, err)
	}

	return rpy.Cabinets, nil
}

func Machines(url string) ([]string, error) {
	rpy, err := getData(url, MachineBase)
	if err != nil {
		return nil, fmt.Errorf("get machines (%s%s): %v", url, MachineBase, err)
	}

	return rpy.Machines, nil
}
