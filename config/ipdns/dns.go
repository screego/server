package ipdns

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type DNS struct {
	sync.Mutex

	DNS      string
	Resolver *net.Resolver
	Domain   string

	refetch time.Time
	v4      net.IP
	v6      net.IP
	err     error
}

func (s *DNS) Get() (net.IP, net.IP, error) {
	s.Lock()
	defer s.Unlock()

	if s.refetch.Before(time.Now()) {
		oldV4, oldV6 := s.v4, s.v6
		s.v4, s.v6, s.err = s.lookup()
		if s.err == nil {
			if !oldV4.Equal(s.v4) || !oldV6.Equal(s.v6) {
				log.Info().Str("v4", s.v4.String()).
					Str("v6", s.v6.String()).
					Str("domain", s.Domain).
					Str("dns", s.DNS).
					Msg("DNS External IP")
			}
			s.refetch = time.Now().Add(time.Minute)
		} else {
			// don't spam the dns server
			s.refetch = time.Now().Add(time.Second)
			log.Err(s.err).Str("domain", s.Domain).Str("dns", s.DNS).Msg("DNS External IP")
		}
	}

	return s.v4, s.v6, s.err
}

func (s *DNS) lookup() (net.IP, net.IP, error) {
	ips, err := s.Resolver.LookupIP(context.Background(), "ip", s.Domain)
	if err != nil {
		if dns, ok := err.(*net.DNSError); ok && s.DNS != "system" {
			dns.Server = ""
		}
		return nil, nil, err
	}

	var v4, v6 net.IP
	for _, ip := range ips {
		isV6 := strings.Contains(ip.String(), ":")
		if isV6 && v6 == nil {
			v6 = ip
		} else if !isV6 && v4 == nil {
			v4 = ip
		}
	}

	if v4 == nil && v6 == nil {
		return nil, nil, errors.New("dns record doesn't have an A or AAAA record")
	}

	return v4, v6, nil
}
