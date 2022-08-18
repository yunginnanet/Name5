package main

import (
	"os"

	"git.tcp.direct/kayos/name5"
)

func printPTR(dnm *name5.DNSMap) {
	println("\n----------ptr-")
	defer println("-------------\n")
	var count = 0
	for kv := range dnm.IPToPTR.IterBuffered() {
		println("ip:", kv.Key, "\nptr:", kv.Val.(string))
		if count == dnm.IPToPTR.Count()-1 {
			break
		}
		println("+ - + - +")
		count++
	}
}

func printNameToIP(dnm *name5.DNSMap) (nores []string) {
	println("\n---------name-")
	defer println("-------------\n")
	var count = 0
	for kv := range dnm.NameToIPs.IterBuffered() {
		vals := kv.Val.(*name5.IPName)
		if len(vals.IPs) == 0 {
			nores = append(nores, kv.Key)
			continue
		}
		prefix := ".,.,.\n"
		if count == 0 {
			prefix = ""
		}
		println(prefix+"name:", kv.Key)
		for _, addr := range vals.IPs {
			println(addr)
		}
		if count == dnm.NameToIPs.Count()-1 {
			break
		}
		count++
	}
	return
}

func main() {
	dnm := name5.NewDNSMap(os.Args[1])
	nores := printNameToIP(dnm)
	printPTR(dnm)
	if len(nores) < 1 {
		return
	}
	println("\nobjects without unique (or any) IPs:\n")
	for _, name := range nores {
		println(name)
	}
}
