package main

import (
	"os"

	"git.tcp.direct/kayos/name5"
)

func main() {
	var nores []string
	dnm := name5.NewDNSMap(os.Args[1])
	println("\n----------ptr-\n")
	for kv := range dnm.IPToPTR.IterBuffered() {
		println("ip:", kv.Key, "ptr:", kv.Val.(string))
	}
	println("\n-------------\n")
	for kv := range dnm.NameToIPs.IterBuffered() {
		vals := kv.Val.(*name5.IPName)
		if len(vals.IPs) == 0 {
			nores = append(nores, kv.Key)
			continue
		}
		println("name:", kv.Key)
		if len(vals.IPs) == 1 {
			print(vals.IPs[0] + "\n")
			continue
		}
		for _, addr := range vals.IPs {
			println(addr)
		}
		println("- - - - -")
	}
	if len(nores) < 1 {
		return
	}
	println("\nobjects with no results:")
	for _, name := range nores {
		println(name)
	}
}
