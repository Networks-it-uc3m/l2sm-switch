package linuxif

import (
	"fmt"
	"net"
	"sort"

	"github.com/vishvananda/netlink"
)

// ListNames returns all interface names currently present in the network namespace.
func ListNames() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(interfaces))
	for _, ifc := range interfaces {
		names = append(names, ifc.Name)
	}
	sort.Strings(names)
	return names, nil
}

func Exists(name string) bool {
	if name == "" {
		return false
	}
	_, err := net.InterfaceByName(name)
	if err == nil {
		return true
	}
	// Common "not found" cases are *net.OpError / syscall.ENODEV on Unix,
	// but the exact error varies by OS. If you only care about existence,
	// treat any error as "doesn't exist" unless you want to distinguish perms/etc.
	return false
}

func AddInterfaceToLinuxBridge(interfaceName, switchName string) error {
	l, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("could not find link by name. erro: %v", err)
	}
	master, err := netlink.LinkByName(switchName)
	err = netlink.LinkSetMaster(l, master)
	if err != nil {
		return fmt.Errorf("command error: %v", err)
	}
	return nil

}

func AddVethPair(vethName, peerName string) error {
	v := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: vethName,
		},
		PeerName: peerName,
	}

	if err := netlink.LinkAdd(v); err != nil {
		return fmt.Errorf("add veth %s<->%s: %v", vethName, peerName, err)
	}

	hostL, err := netlink.LinkByName(vethName)
	if err != nil {
		return fmt.Errorf("get link %s: %v", vethName, err)
	}

	peerL, err := netlink.LinkByName(peerName)
	if err != nil {
		return fmt.Errorf("get link %s: %v", peerName, err)
	}

	if err := netlink.LinkSetUp(hostL); err != nil {
		return fmt.Errorf("set up %s: %v", vethName, err)
	}

	if err := netlink.LinkSetUp(peerL); err != nil {
		return fmt.Errorf("set up %s: %v", peerName, err)
	}

	return nil
}
