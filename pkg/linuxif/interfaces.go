package linuxif

import (
	"net"
	"sort"
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
