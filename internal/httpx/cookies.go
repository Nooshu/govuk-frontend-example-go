package httpx

import (
	"net/url"
	"strings"
)

// ParseCookies reads a Cookie request header.
//
// A value that is not valid percent-encoding is kept as written rather than dropped, because a
// cookie set by something else on the same host should not break session lookup.
func ParseCookies(header string) map[string]string {
	cookies := map[string]string{}
	if header == "" {
		return cookies
	}
	for _, part := range strings.Split(header, ";") {
		separator := strings.Index(part, "=")
		if separator == -1 {
			continue
		}
		name := strings.TrimSpace(part[:separator])
		if name == "" {
			continue
		}
		value := strings.TrimSpace(part[separator+1:])
		if decoded, err := url.QueryUnescape(value); err == nil {
			value = decoded
		}
		cookies[name] = value
	}
	return cookies
}
