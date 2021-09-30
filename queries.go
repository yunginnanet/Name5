package Name5

import (
	"time"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
	"github.com/yunginnanet/Rate5"
)

// TODO: benchmarks
// TODO: better ratelimiting logic
// global dummy type for rater
type resolveSlower struct {
	key string
}

func (rs *resolveSlower) UniqueKey() string {
	return rs.key
}

var dnsRater *rate5.Limiter

/*
SmuggleError crafts a custom DNS answer to include our error message.
This allows the function to only have one return, thus simplifying our pipeline. (maybe)
*/
func SmuggleError(err error) *dns.Msg {
	dnserr := new(dns.Msg)
	dnserr.Answer = make([]dns.RR, 1)
	dnserr.Answer[0] = new(dns.NULL)
	dnserr.Answer[0].(*dns.NULL).Data = err.Error()
	return dnserr
}

func query(domain string, qtype uint16) *dns.Msg {
	if dnsRater.Check(&resolveSlower{key: "yeet"}) {
		time.Sleep(10 * time.Millisecond)
	}
	domain = dns.Fqdn(domain)
	m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.RecursionDesired = true
	m1.Question = make([]dns.Question, 1)
	m1.Question[0] = dns.Question{Name: domain, Qtype: qtype, Qclass: dns.ClassINET}
	lookup := &dns.Client{
		Timeout:        time.Duration(6) * time.Second,
		SingleInflight: true,
	}
	in, _, err := lookup.Exchange(m1, "127.0.0.1:53")
	if err != nil {
		log.Error().Err(err).Msg("query")
		resp := SmuggleError(err)
		resp.Question = append(resp.Question, m1.Question[0])
		return resp
	}
	in.Answer = dns.Dedup(in.Answer, nil)
	in.Ns = dns.Dedup(in.Ns, nil)
	in.Extra = dns.Dedup(in.Extra, nil)
	return in
}

// Query4 requests an A record answer
func Query4(domain string) *dns.Msg {
	return query(domain, 1)
}

// Query6 requests an AAAA record answer
func Query6(domain string) *dns.Msg {
	return query(domain, 28)
}

// QueryPTR retrievs reverse DNS records
func QueryPTR(ip string) *dns.Msg {
	return query(ip, 12)
}
