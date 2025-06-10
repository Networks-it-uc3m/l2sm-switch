package ovs

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
)

type MockClient struct {
	Commands map[string][]byte        // Key: args string, Value: mocked output
	Errors   map[string]error         // Key: args string, Value: mocked error
	Called   []string                 // Tracks all calls
	Buffers  map[string]*bytes.Buffer // For OutputToBuffer simulation
}

func (m *MockClient) CombinedOutput(args ...string) ([]byte, error) {
	key := strings.Join(args, " ")
	m.Called = append(m.Called, key)
	return m.Commands[key], m.Errors[key]
}

func (m *MockClient) Run(args ...string) error {
	key := strings.Join(args, " ")
	m.Called = append(m.Called, key)
	return m.Errors[key]
}

func (m *MockClient) Output(args ...string) ([]byte, error) {
	key := strings.Join(args, " ")
	m.Called = append(m.Called, key)
	return m.Commands[key], m.Errors[key]
}

func (m *MockClient) OutputToBuffer(stdout *bytes.Buffer, args ...string) error {
	key := strings.Join(args, " ")
	m.Called = append(m.Called, key)

	if err := m.Errors[key]; err != nil {
		return err
	}

	if buf, ok := m.Buffers[key]; ok {
		stdout.Write(buf.Bytes())
		return nil
	}

	if out, ok := m.Commands[key]; ok {
		stdout.Write(out)
		return nil
	}

	return errors.New("no output defined for command")
}

func TestAddBridge(t *testing.T) {
	mock := &MockClient{
		Commands: map[string][]byte{"add-br br0": []byte("")},
		Errors:   map[string]error{},
	}
	svc := OvsService{exec: mock}
	if err := svc.AddBridge("br0"); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestDeleteBridge(t *testing.T) {
	mock := &MockClient{
		Commands: map[string][]byte{"del-br br0": []byte("")},
		Errors:   map[string]error{},
	}
	svc := OvsService{exec: mock}
	if err := svc.DeleteBridge("br0"); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestSetDatapathID(t *testing.T) {
	mock := &MockClient{
		Commands: map[string][]byte{"set bridge br0 other-config:datapath-id=1234": []byte("")},
		Errors:   map[string]error{},
	}
	svc := OvsService{exec: mock}
	if err := svc.SetDatapathID("br0", "1234"); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestSetProtocol(t *testing.T) {
	mock := &MockClient{
		Commands: map[string][]byte{"set bridge br0 protocols=OpenFlow13": []byte("")},
		Errors:   map[string]error{},
	}
	svc := OvsService{exec: mock}
	if err := svc.SetProtocol("br0", "OpenFlow13"); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestSetController(t *testing.T) {
	mock := &MockClient{
		Commands: map[string][]byte{"set-controller br0 tcp:127.0.0.1:6633": []byte("")},
		Errors:   map[string]error{},
	}
	svc := OvsService{exec: mock}
	if err := svc.SetController("br0", "tcp:127.0.0.1:6633"); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestCreateVxlan(t *testing.T) {
	vx := plsv1.Vxlan{
		VxlanId: "vx0", LocalIp: "10.0.0.1", RemoteIp: "10.0.0.2", UdpPort: "4789",
	}
	key := "add-port br0 vx0 -- set interface vx0 type=vxlan options:key=flow options:remote_ip=10.0.0.2 options:local_ip=10.0.0.1 options:dst_port=4789"
	mock := &MockClient{
		Commands: map[string][]byte{key: []byte("")},
		Errors:   map[string]error{},
	}
	svc := OvsService{exec: mock}
	if err := svc.CreateVxlan("br0", vx); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestAddPort(t *testing.T) {
	mock := &MockClient{
		Commands: map[string][]byte{"add-port br0 eth1": []byte("")},
		Errors:   map[string]error{},
	}
	svc := OvsService{exec: mock}
	if err := svc.AddPort("br0", "eth1"); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestGetPorts(t *testing.T) {
	mock := &MockClient{
		Commands: map[string][]byte{"list-ports br0": []byte("eth1\neth2")},
		Errors:   map[string]error{},
	}
	svc := OvsService{exec: mock}
	ports, err := svc.GetPorts("br0")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(ports) != 2 {
		t.Fatalf("expected 2 ports, got: %d", len(ports))
	}
}

func TestGetController(t *testing.T) {
	mock := &MockClient{
		Commands: map[string][]byte{"get-controller br0": []byte("tcp:127.0.0.1:6633")},
		Errors:   map[string]error{},
	}
	svc := OvsService{exec: mock}
	ctrl, err := svc.GetController("br0")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(ctrl) != 1 || ctrl[0] != "tcp:127.0.0.1:6633" {
		t.Errorf("unexpected controller output: %v", ctrl)
	}
}

func TestGetVxlans(t *testing.T) {
	raw := `{
		"data": [
			["vx0", ["map", [["remote_ip", "10.0.0.2"], ["local_ip", "10.0.0.1"], ["dst_port", "4789"]]]]
		],
		"headings": ["name", "options"]
	}`

	key := "--column=name,options --format=json --data=json find Interface type=vxlan"

	mock := &MockClient{
		Commands: map[string][]byte{key: []byte(raw)},
		Errors:   map[string]error{},
	}

	svc := OvsService{exec: mock}
	vxlans, err := svc.GetVxlans("br0")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(vxlans) != 1 {
		t.Fatalf("expected 1 vxlan, got: %d", len(vxlans))
	}

	vx, ok := vxlans["vx0"]
	if !ok {
		t.Fatalf("expected vxlan 'vx0' to be present")
	}

	if vx.RemoteIp != "10.0.0.2" || vx.LocalIp != "10.0.0.1" || vx.UdpPort != "4789" {
		t.Errorf("unexpected vxlan fields: %+v", vx)
	}
}
