package Name5

import (
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type ipname struct {
	Name string
	IPs  []string
}

// DNSMap is used for resolving and keeping track of DNS query results and their relationships
type DNSMap struct {
	IPToPTR      map[string]string
	NameToIP     map[string]*ipname
	Waitlist4    map[string]int
	Waitlist6    map[string]int
	CNAMETargets map[string]int

	born time.Time
	mu   *sync.RWMutex
}

// NewDNSMap creates a new DNSMap type
func NewDNSMap(name string) *DNSMap {
	name = dns.Fqdn(name)
	dnsmap := &DNSMap{ //   ipaddr[domain][ipaddr]boolean
		IPToPTR:      make(map[string]string),
		NameToIP:     make(map[string]*ipname),
		Waitlist4:    make(map[string]int),
		Waitlist6:    make(map[string]int),
		CNAMETargets: make(map[string]int),

		born: time.Now(),
		mu:   &sync.RWMutex{},
	}

	dnsmap.addToWaitlist(dns.Fqdn(name))
	go dnsmap.process(Query4(name))
	go dnsmap.process(Query6(name))
	dnsmap.waitUntilDone()
	return dnsmap
}

func (d *DNSMap) waitUntilDone() {
	for {
		d.mu.RLock()
		if time.Since(d.born) > 5*time.Minute {
			d.mu.RUnlock()
			log.Error().Msg("DNSMap timed out after 5 minutes")
			return
		}
		if len(d.Waitlist4) == 0 && len(d.Waitlist6) == 0 {
			d.mu.RUnlock()
			return
		}
		d.mu.RUnlock()
		time.Sleep(20 * time.Millisecond)
	}
}

func (d *DNSMap) addToWaitlist(name string) {
	name = dns.Fqdn(name)
	d.mu.Lock()
	if _, ok := d.Waitlist4[name]; !ok {
		d.Waitlist4[name] = 1
	}
	if _, ok := d.Waitlist6[name]; !ok {
		d.Waitlist6[name] = 1
	}
	d.mu.Unlock()
}

func (d *DNSMap) isInWaitlist(name string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	name = dns.Fqdn(name)
	_, ok4 := d.Waitlist4[name]
	_, ok6 := d.Waitlist6[name]
	if !ok4 && !ok6 {
		return false
	}
	return true
}

func (d *DNSMap) isPTRResolved(object string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if _, ok := d.IPToPTR[object]; ok {
		return true
	}
	return false
}

func (d *DNSMap) process(in *dns.Msg) {
	var (
		addrs    []string
		question string
		ptr      = ""
		cname    = ""
		qtype    uint16
	)
	question = in.Question[0].Name
	qtype = in.Question[0].Qtype

	for _, resource := range in.Answer {
		switch record := resource.(type) {
		case *dns.NULL:
			log.Error().Caller().Msg(record.Data)
			continue
		case *dns.A:
			addrs = append(addrs, record.A.String())
			ptrReq, _ := dns.ReverseAddr(record.A.String())
			d.addToWaitlist(ptrReq)
			d.process(QueryPTR(ptrReq))
		case *dns.AAAA:
			addrs = append(addrs, record.AAAA.String())
			ptrReq, _ := dns.ReverseAddr(record.AAAA.String())
			d.addToWaitlist(ptrReq)
			d.process(QueryPTR(ptrReq))
		case *dns.PTR:
			ptr = record.Ptr
			d.mu.RLock()
			if _, ok := d.NameToIP[ptr]; !ok {
				d.mu.RUnlock()
				if !d.isInWaitlist(ptr) {
					d.addToWaitlist(ptr)
					go d.process(Query4(ptr))
					go d.process(Query6(ptr))
				}
			} else {
				d.mu.RUnlock()
			}
		case *dns.CNAME:
			cname = record.Target
			d.mu.Lock()
			d.CNAMETargets[cname] = 1
			d.mu.Unlock()
			d.mu.RLock()
			if _, ok := d.NameToIP[cname]; !ok {
				d.mu.RUnlock()
				if !d.isInWaitlist(cname) {
					d.addToWaitlist(cname)
					go d.process(Query4(cname))
					go d.process(Query6(cname))
				}
			} else {
				d.mu.RUnlock()
			}
		default:
			log.Warn().Caller().Interface("resource", record).Msg("unhandled record")
		}
	}

	if cname == "" && ptr == "" && len(addrs) == 0 {
		d.removeWait(question, qtype)
	}

	if ptr != "" {
		d.mu.Lock()
		if _, ok := d.IPToPTR[question]; !ok {
			d.IPToPTR[question] = ptr
		}
		d.mu.Unlock()
	}

	if len(addrs) > 0 {
		d.mu.Lock()
		var resolved *ipname
		var ok bool
		if resolved, ok = d.NameToIP[question]; ok {
			resolved.IPs = append(resolved.IPs, addrs...)
		} else {
			d.NameToIP[question] = &ipname{Name: question, IPs: addrs}
		}
		d.mu.Unlock()
	}

	if ptr != "" {
		d.mu.Lock()
		d.IPToPTR[question] = ptr
		d.mu.Unlock()
	}

	d.removeWait(question, qtype)
}

func (d *DNSMap) removeWait(question string, qtype uint16) {
	if !d.isInWaitlist(question) {
		return
	}
	d.mu.Lock()
	switch qtype {
	case dns.TypeA:
		delete(d.Waitlist4, question)
	case dns.TypeAAAA:
		delete(d.Waitlist6, question)
	case dns.TypeCNAME:
	case dns.TypePTR:
		delete(d.Waitlist4, question)
		delete(d.Waitlist6, question)
	default:
		log.Warn().Caller().Str("question", question).
			Uint16("qtype", qtype).
			Msg("unhandled blank response")
	}
	d.mu.Unlock()
	return
}
