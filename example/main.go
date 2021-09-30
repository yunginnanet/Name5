package main

import (
	"fmt"
	"os"

	"github.com/yunginnanet/Name5"
)

func main() {
	dnsmap := Name5.NewDNSMap(os.Args[1])
	fmt.Println(dnsmap.IPToPTR, "\n\n")
	fmt.Println(dnsmap.NameToIP)
}
