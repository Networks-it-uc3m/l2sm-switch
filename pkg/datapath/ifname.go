package datapath

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	PREFIX     = "l2"
	PREFIX_LEN = 2
	TOKEN_LEN  = 5
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

var types = map[IfType]string{
	TypePort:    "",
	TypeProbe:   "probe",
	TypeUnknown: "UNKNOWN",
}

func New(switchName string) Ifid {
	return Ifid{token: GenerateID(switchName)[:6]}
}
func (ifid *Ifid) Port(id int) string {
	return fmt.Sprintf("%s%s-%d", PREFIX, ifid.token, id)
}

func (ifid *Ifid) Probe(id int) string {
	return fmt.Sprintf("%s%sprobe-%d", PREFIX, ifid.token, id)
}

func Parse(ifname string) (int, IfType, Ifid, error) {

	ifid := Ifid{}
	if !IsManaged(ifname) {
		return 0, TypeUnknown, ifid, fmt.Errorf("interface is not managed by talpa")
	}

	// todo: check length
	ifid.token = ifname[PREFIX_LEN:(TOKEN_LEN - PREFIX_LEN)]
	id, b := strings.CutPrefix(ifname[(TOKEN_LEN-PREFIX_LEN):], types[TypeProbe])

	parsedId, err := strconv.Atoi(id)
	if err != nil {
		return 0, TypeUnknown, ifid, fmt.Errorf("could not find valid id")
	}

	if b {
		return parsedId, TypeProbe, ifid, nil
	}

	return parsedId, TypePort, ifid, nil
}
func (ifid *Ifid) IsManaged(name string) bool {
	return strings.HasPrefix(name, fmt.Sprintf("%s%s", PREFIX, ifid.token))
}
func IsManaged(name string) bool {
	return strings.HasPrefix(name, fmt.Sprintf("%s", PREFIX))
}
