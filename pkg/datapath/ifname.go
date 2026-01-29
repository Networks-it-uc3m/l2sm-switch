package datapath

import (
	"fmt"
	"strconv"
	"strings"
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
	TypeUnknown
)

var iftypes = map[IfType]string{
	TypePort:    "",
	TypeProbe:   "probe",
	TypeUnknown: "UNKNOWN",
}

func New(switchName string) Ifid {
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

// Parse interface by given name. Will return the id, the type of interface, the switch if. If the name is not matched to a datapath interface,
// an error is returned, alongside empty fields.
func Parse(ifname string) (int, IfType, Ifid, error) {

	ifid := Ifid{}

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
