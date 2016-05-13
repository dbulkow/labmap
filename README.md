# Lab Map

Serves Linux lab layout data.  A lab map holds the cabinet/rack, location in cabinet COM port connections PDU and KVM details.

~~~~
lin301  lnx1 pos0 com1-no  com2-yes pdu5 kvm3
lin302  lnx2 pos0 com1-no  com2-yes pdu5 kvm3
lin303  lnx6 pos0 com1-no  com2-yes pdu6 kvm6
lin304  lnx3 pos0 com1-yes com2-no  pdu5 kvm3
lin305  lnx4 pos0 com1-yes com2-yes pdu4 kvm4
lin306  lnx7 pos0 com1-no  com2-yes pdu5 kvm6
~~~~

The goal of labmap is to provide the backing data needed for the VTM links web page.  lab.map should be deployed from the lab_config git repository so that changes in git are reflected in the web page.

# JSON requests and replies

Only GET methods are supported.

## /v1/machines/

Returns a list of machines from lab.map.  Used by labhtml to order the machine list on the VTM links paage.

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

## /v1/cabinet/

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

Names can be decoded using the macmap service.
