package datapath

import (
	"fmt"
	"strconv"
	"strings"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
)

const (
	PREFIX    = "ls"
	TOKEN_LEN = 5
)

type Ifid struct {
	token string
}
type IfType int

const (
	TypePort IfType = iota
	TypeProbe
	TypePeer
	TypeUnknown
)

var iftypes = map[IfType]string{
	TypePort:    "",
	TypeProbe:   "probe",
	TypePeer:    "peer",
	TypeUnknown: "UNKNOWN",
}

func NewIfId(switchName string) Ifid {
	id := GenerateID(switchName)
	if len(id) < TOKEN_LEN {
		id = id + strings.Repeat("0", (TOKEN_LEN-len(id)))
	}
	return Ifid{token: id[:TOKEN_LEN]}
}
func (ifid *Ifid) Port(id int) string {
	return fmt.Sprintf("%s%s%d", PREFIX, ifid.token, id)
}

func (ifid *Ifid) Probe(id int) string {
	return fmt.Sprintf("%s%s%s%d", PREFIX, ifid.token, iftypes[TypeProbe], id)
}

// GeneratePeerName returns the Linux peer interface name that corresponds to a
// datapath port. Preferred format is "l2peer<port-id>".
func GeneratePeerName(port plsv1.Port) string {
	if port.Id != nil {
		return fmt.Sprintf("%s%s%d", PREFIX, iftypes[TypePeer], *port.Id)
	}

	id, typ, _, err := Parse(port.Name)
	if err == nil && typ == TypePort {
		return fmt.Sprintf("%s%s%d", PREFIX, iftypes[TypePeer], id)
	}

	return fmt.Sprintf("%s%s%s", PREFIX, iftypes[TypePeer], port.Name)
}

// Parse interface by given name. Will return the id, the type of interface, the switch if. If the name is not matched to a datapath interface,
// an error is returned, alongside empty fields.
func Parse(ifname string) (int, IfType, Ifid, error) {

	ifid := Ifid{}

	peerPrefix := fmt.Sprintf("%s%s", PREFIX, iftypes[TypePeer])
	if after, ok := strings.CutPrefix(ifname, peerPrefix); ok {
		idStr := after
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return 0, TypeUnknown, ifid, fmt.Errorf("could not find valid id")
		}
		return id, TypePeer, ifid, nil
	}

	// if it is not managed by switch if, or the length is not appropiate (prefix length + token length + id length), this interface is not ours.
	if !IsManaged(ifname) || len(ifname) < (len(PREFIX)+TOKEN_LEN+1) {
		return 0, TypeUnknown, ifid, fmt.Errorf("interface is not managed by talpa")
	}

	// token is what comes between the prefix "ls" and the token_len position.
	ifid.token = ifname[len(PREFIX):(TOKEN_LEN + len(PREFIX))]
	//id is the rest of the ifname, unless it has the prefix probe, in which case it is  what comes after it
	idStr, isProbe := strings.CutPrefix(ifname[(TOKEN_LEN+len(PREFIX)):], iftypes[TypeProbe])

	// parse the id. if it is not an int, the naming is probably wrongly generated, so we cant identify it correctly
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, TypeUnknown, ifid, fmt.Errorf("could not find valid id")
	}

	// if it is of type probe, b = true
	if isProbe {
		return id, TypeProbe, ifid, nil
	}

	return id, TypePort, ifid, nil
}
func (ifid *Ifid) IsManaged(name string) bool {
	return strings.HasPrefix(name, fmt.Sprintf("%s%s", PREFIX, ifid.token))
}
func IsManaged(name string) bool {
	return strings.HasPrefix(name, fmt.Sprintf("%s", PREFIX))
}
