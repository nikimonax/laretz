package middleware

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

func OriginCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		u, err := url.Parse(origin)
		if err != nil || u.Host == "" || !hostsMatch(u.Host, r.Host) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func hostsMatch(originHost, requestHost string) bool {
	if strings.EqualFold(originHost, requestHost) {
		return true
	}
	return isLoopbackHost(hostWithoutPort(originHost)) && isLoopbackHost(hostWithoutPort(requestHost))
}

func hostWithoutPort(hostport string) string {
	if host, _, err := net.SplitHostPort(hostport); err == nil {
		return host
	}
	return hostport
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
