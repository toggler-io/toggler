package httputils

import (
	"net"
	"net/http"
	"regexp"
)

var isEndingToPortNumber = regexp.MustCompile(`:\d+$`)
var byComma = regexp.MustCompile(` *, *`)

// GetClientIP expect that if a proxy or loadbalancer is being used,
// the unit is correctly configured to pass forward the client's ip address by one of the two convention.
//
// X-Real-Ip - fetches first true IP (if the requests sits behind multiple NAT sources/load balancer)
// X-Forwarded-For - if for some reason X-Real-Ip is blank and does not return response, get from X-Forwarded-For
// Remote Address - last resort (usually won't be reliable as this might be the last ip or if it is a naked http request to server ie no load balancer)
func GetClientIP(r *http.Request) string {
	if v := r.Header.Get("X-Real-Ip"); v != `` {
		return v
	}

	if rawIPs := r.Header.Get("X-Forwarded-For"); rawIPs != "" {
		ips := byComma.Split(rawIPs, 2)
		return ips[0]
	}

	if isEndingToPortNumber.MatchString(r.RemoteAddr) {
		if h, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			return h
		}
	}

	return r.RemoteAddr
}
