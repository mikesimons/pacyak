package main

import (
	"fmt"
	"net"
	"sort"
)

func makeInterfaceMap() map[string]string {
	interfaceTable := make(map[string]string)

	interfaces, _ := net.Interfaces()

	for _, intf := range interfaces {
		key := fmt.Sprintf("%d-%s", intf.Flags, intf.HardwareAddr.String())
		addrs, _ := intf.Addrs()
		for _, addr := range addrs {
			key = fmt.Sprintf("%s-%s", key, addr.String())
		}
		interfaceTable[intf.Name] = key
	}

	return interfaceTable
}

func interfaceMapKeys(ifs map[string]string) []string {
	var ret []string
	for k, _ := range ifs {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret
}

func interfaceListChanged(listA []string, listB []string) bool {
	if len(listA) != len(listB) {
		return true
	}

	equal := true
	for i, v := range listA {
		if listB[i] != v {
			equal = false
			break
		}
	}
	return !equal
}
