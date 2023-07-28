package config

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/screego/server/config/ipdns"
)

func parseIPProvider(ips []string, config string) (ipdns.Provider, []FutureLog) {
	if len(ips) == 0 {
		panic("must have at least one ip")
	}

	first := ips[0]
	if strings.HasPrefix(first, "dns:") {
		if len(ips) > 1 {
			return nil, []FutureLog{futureFatal(fmt.Sprintf("invalid %s: when dns server is specified, only one value is allowed", config))}
		}

		return parseDNS(strings.TrimPrefix(first, "dns:")), nil
	}

	return parseStatic(ips, config)
}

func parseStatic(ips []string, config string) (*ipdns.Static, []FutureLog) {
	var static ipdns.Static

	firstV4, errs := applyIPTo(config, ips[0], &static)
	if errs != nil {
		return nil, errs
	}

	if len(ips) == 1 {
		return &static, nil
	}

	secondV4, errs := applyIPTo(config, ips[1], &static)
	if errs != nil {
		return nil, errs
	}

	if firstV4 == secondV4 {
		return nil, []FutureLog{futureFatal(fmt.Sprintf("invalid %s: the ips must be of different type ipv4/ipv6", config))}
	}

	if len(ips) > 2 {
		return nil, []FutureLog{futureFatal(fmt.Sprintf("invalid %s: too many ips supplied", config))}
	}

	return &static, nil
}

func applyIPTo(config, ip string, static *ipdns.Static) (bool, []FutureLog) {
	parsed := net.ParseIP(ip)
	if parsed == nil || ip == "0.0.0.0" {
		return false, []FutureLog{futureFatal(fmt.Sprintf("invalid %s: %s", config, ip))}
	}

	v4 := parsed.To4() != nil
	if v4 {
		static.V4 = parsed
	} else {
		static.V6 = parsed
	}
	return v4, nil
}

func parseDNS(dnsString string) *ipdns.DNS {
	var dns ipdns.DNS

	parts := strings.SplitN(dnsString, "@", 2)

	dns.Domain = parts[0]
	dns.DNS = "system"
	if len(parts) == 2 {
		dns.DNS = parts[1]
		dns.Resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 10 * time.Second}
				return d.DialContext(ctx, network, parts[1])
			},
		}
	}

	return &dns
}
