//go:build integration
// +build integration

package ovs

import (
	"testing"
	"time"

	plsv1 "github.com/Networks-it-uc3m/l2sm-switch/api/v1"
)

func TestIntegration_AddDeleteBridge(t *testing.T) {
	svc := NewOvsService()

	bridge := "int-test-br"
	err := svc.AddBridge(bridge)
	if err != nil {
		t.Fatalf("AddBridge failed: %v", err)
	}

	// Allow some time for OVS to register
	time.Sleep(1 * time.Second)

	err = svc.DeleteBridge(bridge)
	if err != nil {
		t.Fatalf("DeleteBridge failed: %v", err)
	}
}

func TestIntegration_SetProtocolAndController(t *testing.T) {
	svc := NewOvsService()

	bridge := "int-test-br2"
	_ = svc.AddBridge(bridge)
	defer svc.DeleteBridge(bridge)

	err := svc.SetProtocol(bridge, "OpenFlow13")
	if err != nil {
		t.Fatalf("SetProtocol failed: %v", err)
	}

	err = svc.SetController(bridge, "tcp:127.0.0.1:6633")
	if err != nil {
		t.Fatalf("SetController failed: %v", err)
	}

	controllers, err := svc.GetController(bridge)
	if err != nil {
		t.Fatalf("GetController failed: %v", err)
	}

	if len(controllers) == 0 {
		t.Error("Expected controller to be set, got none")
	}
}

func TestIntegration_CreateVxlan(t *testing.T) {
	svc := NewOvsService()

	bridge := "int-test-br3"
	_ = svc.AddBridge(bridge)
	defer svc.DeleteBridge(bridge)

	vx := plsv1.Vxlan{
		VxlanId:  "vx0",
		LocalIp:  "10.0.0.1",
		RemoteIp: "10.0.0.2",
		UdpPort:  "4789",
	}

	err := svc.CreateVxlan(bridge, vx)
	if err != nil {
		t.Fatalf("CreateVxlan failed: %v", err)
	}

	vxlans, err := svc.GetVxlans(bridge)
	if err != nil {
		t.Fatalf("GetVxlans failed: %v", err)
	}

	found := false
	for _, v := range vxlans {
		if v.VxlanId == "vx0" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected vxlan vx0 to be found")
	}
}
