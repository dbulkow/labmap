package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMachine(t *testing.T) {
	machines = []string{
		"onemachine",
		"twomachine",
		"threemachine",
		"four",
	}

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()

	serveMachines(w, req)

	golden, err := ioutil.ReadFile("testdata/machines.golden")
	if err != nil {
		t.Fatal("read machine.golden:", err)
	}

	if bytes.Compare(golden, w.Body.Bytes()) != 0 {
		t.Error("response != golden")
	}
}

func TestCabinets(t *testing.T) {
	cabinet = make(map[string]*Cabinet)

	if err := readmap("testdata/cabinet.map"); err != nil {
		t.Fatal("readmap error:", err)
	}

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()

	serveCabinets(w, req)

	golden, err := ioutil.ReadFile("testdata/cabinets.golden")
	if err != nil {
		t.Fatal("read cabinets.golden:", err)
	}

	if bytes.Compare(golden, w.Body.Bytes()) != 0 {
		t.Error("response != golden")
	}
}

func TestCabinet(t *testing.T) {
	cabinet = make(map[string]*Cabinet)

	if err := readmap("testdata/cabinet.map"); err != nil {
		t.Fatal("readmap error:", err)
	}

	req, _ := http.NewRequest("GET", "lin302", nil)
	w := httptest.NewRecorder()

	serveCabinets(w, req)

	golden, err := ioutil.ReadFile("testdata/cabinet.golden")
	if err != nil {
		t.Fatal("read cabinets.golden:", err)
	}

	if bytes.Compare(golden, w.Body.Bytes()) != 0 {
		t.Error("response != golden")
	}
}

func TestCabinetNotFound(t *testing.T) {
	cabinet = make(map[string]*Cabinet)

	if err := readmap("testdata/cabinet.map"); err != nil {
		t.Fatal("readmap error:", err)
	}

	req, _ := http.NewRequest("GET", "linXXX", nil)
	w := httptest.NewRecorder()

	serveCabinets(w, req)

	var resp struct {
		Status string `json:"status"`
		Error  string `json:"error"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal("unmarshal response:", err)
	}

	if resp.Status != "Failed" {
		t.Fatalf("expected 'Failed', got '%s'\n", resp.Status)
	}

	if resp.Error != "Not Found" {
		t.Fatalf("expected 'Not Found', got '%s'\n", resp.Error)
	}
}

func TestReadmapNoFile(t *testing.T) {
	cabinet = make(map[string]*Cabinet)

	if err := readmap("testdata/no_file"); err == nil {
		t.Fatal("expected readmap error, got none")
	}
}
