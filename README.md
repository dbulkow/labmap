# Lab Map

Serves Linux lab layout data.  A lab map holds the cabinet/rack, location in cabinet COM port connections PDU and KVM details.  The goal of labmap is to provide the backing data needed for services that want to know cabinet-level details for ftServer machines, such as the VTM links web page.

Configuration data is loaded from [lab.map](http://yin.mno.stratus.com/gogs/dbulkow/lab_config/raw/master/config/lab.map), which contains data like the following:

~~~~
lin301   lnx1 pos0 com1-no                        com2-57600,8,1,N:/dev/ttyUSB0  pdu5 kvm3
lin304   lnx3 pos0 com1-57600,8,1,N:/dev/ttyUSB0  com2-no                        pdu5 kvm3
lin01    lnx1 pos1 com1-no                        com2-57600,8,1,N:/dev/ttyUSB1  pdu6 kvm3
lin02    lnx2 pos1 com1-no                        com2-57600,8,1,N:/dev/ttyUSB1  pdu6 kvm3
bahamut  lnx4 pos2 com1-57600,8,1,N:/dev/ttyUSB3  com2-no                        pdu6 kvm4
elliot   lnx7 pos3 com1-no                        com2-57600,8,1,N:/dev/ttyUSB3  pdu8 kvm6
~~~~

Configuration is stored in [Consul](http://consul.io) key/value storage, distributed across many of the lab infrastructure for redundancy and access.

To deploy the labmap service listening on a non-default port number, specify a port number on the command line.

~~~~
Usage of ./labmap:
  -port string
    	http port number (default "8080")
~~~~

# Labmap API - golang

~~~~
func GetCabinet(url, machine string) (*Cabinet, error)
func Cabinets(url string) (map[string]*Cabinet, error)
func Machines(url string) ([]string, error)
~~~~

# Key/Value Storage

Lab configuration data is stored under the _directory_ `labconfig`.  The hostname is used as the key and the value is JSON encoded data using the following golang structures:

~~~~
type ComPort struct {
	Enabled  bool   `json:"enabled"`
	Speed    int    `json:"speed,omitempty"`
	Bits     int    `json:"bits,omitempty"`
	StopBits int    `json:"stopbits,omitempty"`
	Parity   string `json:"parity,omitempty"`
	Device   string `json:"device,omitempty"`
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
~~~~

# Web API

## List machines

Returns a list of machines from lab.map.  Used by labhtml to order the machine list on the VTM links paage.

~~~~
GET /v1/machines/
~~~~

#### Response

~~~~
Status: 200 OK
Content-Type: application/json
~~~~
~~~~
{
    "status": "Success",
    "machines": [
        "lin01",
        "lin02",
        "lin03",
    ]
}
~~~~

#### Golang

~~~~
type Reply struct {
	Status   string   `json:"status"`
	Error    string   `json:"error,omitempty"`
	Machines []string `json:"machines"`
}
~~~~

| Field    | Format   | Description |
| -------- | -------- | ----------- |
| status   | string   | "Success" or "Failed" |
| error    | string   | When status is "Failed" the error field will be error test from the server |
| machines | array of string | Machine names |

## List of Cabinets

Without a machine name in the URL the reply will be a map of Cabinets.

~~~~
GET /v1/cabinet/
~~~~

#### Response

~~~~
Status: 200 OK
Content-Type: application/json
~~~~
~~~~
{
    "status": "Success",
    "cabinets": {
        "lin01": {
            "vtm0": "lin01-vtm0",
	    "vtm1": "lin01-vtm1",
	    "cabinet": "1",
	    "position": "2",
	    "com1": "",
	    "com2": "telnet lnx1-debug 12345",
	    "outlet": "15",
	    "kvm": "lnx1-kvm",
	    "pdu0": "lnx1-pdu0",
	    "pdu1": "lnx1-pdu1"
	},
        "lin02": {
            "vtm0": "lin02-vtm0",
	    "vtm1": "lin02-vtm1",
	    "cabinet": "1",
	    "position": "3",
	    "com1": "",
	    "com2": "telnet lnx1-debug 12346",
	    "outlet": "18",
	    "kvm": "lnx1-kvm",
	    "pdu0": "lnx1-pdu0",
	    "pdu1": "lnx1-pdu1"
	}
    }
}
~~~~

With a machine name, the reply will contain a single Cabinet.

~~~~
GET /v1/cabinet/lin01
~~~~

#### Response

~~~~
Status: 200 OK
Content-Type: application/json
~~~~
~~~~
{
    "status": "Success",
    "cabinet": {
        "vtm0": "lin01-vtm0",
        "vtm1": "lin01-vtm1",
        "cabinet": "1",
        "position": "2",
        "com1": "",
        "com2": "telnet lnx1-debug 12345",
        "outlet": "15",
        "kvm": "lnx1-kvm",
        "pdu0": "lnx1-pdu0",
        "pdu1": "lnx1-pdu1"
    }
}
~~~~

#### Golang

~~~~
type Reply struct {
	Status   string              `json:"status"`
	Error    string              `json:"error,omitempty"`
	Cabinet  *Cabinet            `json:"cabinet,omitempty"`
	Cabinets map[string]*Cabinet `json:"cabinets,omitempty"`
}

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
~~~~

### Reply

| Field | Format | Description |
| ----- | ------ | ----------- |
| status | string | "Success" or "Failed" |
| error  | string | When status is "Failed" the error field will be error test from the server |
| cabinet | Cabinet | If requesting data for a single machine, the Cabinet structure will be occupied |
| cabinets | []Cabinet | If not restricting to one machine, an array of Cabinet structures |

### Cabinet

| Field | Format | Description |
| ----- | ------ | ----------- |
| vtm0     | string | Name of VTM0 |
| vtm1     | string | Name of VTM1 |
| cabinet  | string | Cabinet/Rack number |
| position | string | Position in cabinet.  The top machine is 0. |
| com1     | string | Telnet command line for COM1 |
| com2     | string | Telnet command line for COM2 |
| outlet   | string | Outlet number on PDU |
| kvm      | string | Name of KVM |
| pdu0     | string | Name of PDU0 |
| pdu1     | string | Name of PDU1 |

Hostname addresses can be determined using the macmap service.
