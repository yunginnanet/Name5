package name5

import (
	"context"
	"net"
	"sync/atomic"
	"time"

	"github.com/miekg/dns"
	"github.com/orcaman/concurrent-map"
	"github.com/rs/zerolog/log"
)

type ipname struct {
	Name string
	IPs  []string
}

// DNSMap is used for resolving and keeping track of DNS query results and their relationships
type DNSMap struct {
	IPToPTR      cmap.ConcurrentMap
	NameToIPs    cmap.ConcurrentMap // map[string]*ipname
	Working      *int64
	CNAMETargets *int64

	born *time.Time
	ctx  context.Context
}

func newPtr(i int64) *int64 {
	return &i
}

func (dnm *DNSMap) initCounters() {
	dnm.Working = newPtr(0)
	dnm.CNAMETargets = newPtr(0)
}

func (dnm *DNSMap) initMaps() {
	dnm.IPToPTR = cmap.New()
	dnm.NameToIPs = cmap.New() // make(map[string]*ipname)
}

// NewDNSMap creates a new DNSMap type
func NewDNSMap(name string) *DNSMap {
	tn := time.Now()
	name = dns.Fqdn(name)
	dnsmap := &DNSMap{ //   ipaddr[domain][ipaddr]boolean
		born: &tn,
	}

	dnsmap.initMaps()
	dnsmap.initCounters()

	ipa := net.ParseIP(name)
	if _, ok := dns.IsDomainName(name); !ok && ipa != nil {
		var err error
		name, err = dns.ReverseAddr(name)
		if err != nil {
			return &DNSMap{}
		}
	}

	atomic.AddInt64(dnsmap.Working, 2)
	go dnsmap.process(Query4(name))
	go dnsmap.process(Query6(name))
	dnsmap.waitUntilDone()
	return dnsmap
}

func (dnm *DNSMap) waitUntilDone() {
	time.Sleep(1 * time.Second)
	for {
		select {
		case <-dnm.ctx.Done():
			return
		default:
			time.Sleep(5 * time.Second)
			if atomic.LoadInt64(dnm.Working) == 0 &&
				atomic.LoadInt64(dnm.CNAMETargets) == 0 {
				return
			}
		}
	}
}

func (dnm *DNSMap) isPTRResolved(object string) bool {
	return dnm.IPToPTR.Has(object)
}

func (dnm *DNSMap) chipOff(delay int) {
	go func() {
		time.Sleep(time.Duration(delay) * time.Millisecond)
		atomic.AddInt64(dnm.Working, -1)
	}()
}

func (dnm *DNSMap) process(in *dns.Msg) {
	question := in.Question[0].Name
	var addrs []string
	for _, resource := range in.Answer {

		log.Trace().Int64("working", atomic.LoadInt64(dnm.Working)).
			Interface("resource", resource).Msg("working...")

		switch record := resource.(type) {
		case *dns.NULL:
			log.Error().Caller().Msg(record.Data)
			dnm.chipOff(500)
			continue
		case *dns.A:
			a := record.A.String()
			addrs = append(addrs, a)
			atomic.AddInt64(dnm.Working, 1)
			go dnm.process(QueryPTR(a))
			dnm.chipOff(100)
		case *dns.AAAA:
			aaaa := record.AAAA.String()
			addrs = append(addrs, aaaa)
			atomic.AddInt64(dnm.Working, 1)
			go dnm.process(QueryPTR(aaaa))
			dnm.chipOff(100)
		case *dns.PTR:
			ptr := record.Ptr
			ogIP, ok := arpaToIP.Get(in.Question[0].Name)
			if !ok {
				log.Warn().Caller().Interface("in", in).
					Msg("no IP for rev ARPA name from question...")
			}
			dnm.IPToPTR.Set(ogIP.(string), ptr)
			if !dnm.NameToIPs.Has(ptr) {
				atomic.AddInt64(dnm.Working, 2)
				go dnm.process(Query4(ptr))
				go dnm.process(Query6(ptr))
			}
			dnm.chipOff(100)
		case *dns.CNAME:
			cname := record.Target
			if !dnm.NameToIPs.Has(cname) {
				atomic.AddInt64(dnm.Working, 2)
				go dnm.process(Query4(cname))
				go dnm.process(Query6(cname))
			}
			dnm.chipOff(100)
		default:
			log.Warn().Caller().Interface("resource", record).Msg("unhandled record")
		}
	}
	if len(addrs) < 0 {
		return
	}
	current, ok := dnm.NameToIPs.Get(question)
	if !ok {
		dnm.NameToIPs.Set(question, &ipname{Name: question, IPs: addrs})
		return
	}
	resolved := current.(*ipname)
	resolved.IPs = append(resolved.IPs, addrs...)
	dnm.NameToIPs.Set(question, resolved)
}
