package baseline

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// CookieOptions overrides the cookie defaults in policy.json for one Set-Cookie header.
type CookieOptions struct {
	// SameSite is "Lax", "Strict", or "None". Empty uses the policy default.
	SameSite string
	// Secure and HTTPOnly are pointers so "not set" can be told apart from "set to false".
	Secure   *bool
	HTTPOnly *bool
	// Path defaults to the policy path when empty.
	Path string
	// MaxAge is the cookie lifetime in seconds. Nil omits Max-Age, making it a session cookie.
	MaxAge *int
	// HostPrefix asserts that the caller meant to use a __Host- cookie.
	HostPrefix bool
}

var (
	cookieName  = regexp.MustCompile("^[!#$%&'*+\\-.^_`|~0-9A-Za-z]+$")
	cookieValue = regexp.MustCompile(`^[\x21\x23-\x2B\x2D-\x3A\x3C-\x5B\x5D-\x7E]*$`)
)

// SetCookie builds a Set-Cookie header value from the policy defaults.
//
// It refuses combinations browsers silently reject — SameSite=None without Secure, or a __Host-
// cookie that is not Secure, not Path=/, or carries a Domain — so a cookie that would be dropped
// is a start-up error rather than a session that mysteriously never persists.
func (p *Policy) SetCookie(name, value string, options CookieOptions) (string, error) {
	if !cookieName.MatchString(name) {
		return "", fmt.Errorf("baseline: invalid cookie name: %q", name)
	}
	if !cookieValue.MatchString(value) {
		return "", fmt.Errorf("baseline: invalid cookie value for %q", name)
	}

	sameSite := options.SameSite
	if sameSite == "" {
		sameSite = p.Cookie.SameSite
	}
	if sameSite != "Lax" && sameSite != "Strict" && sameSite != "None" {
		return "", fmt.Errorf("baseline: invalid SameSite: %q", sameSite)
	}

	secure := p.Cookie.Secure
	if options.Secure != nil {
		secure = *options.Secure
	}
	httpOnly := p.Cookie.HTTPOnly
	if options.HTTPOnly != nil {
		httpOnly = *options.HTTPOnly
	}
	if sameSite == "None" && !secure {
		return "", fmt.Errorf("baseline: SameSite=None requires Secure")
	}

	path := options.Path
	if path == "" {
		path = p.Cookie.Path
	}
	if !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("baseline: cookie Path must start with /, got %q", path)
	}

	hostPrefixed := strings.HasPrefix(name, "__Host-")
	if options.HostPrefix && !hostPrefixed {
		return "", fmt.Errorf("baseline: HostPrefix requires a __Host- cookie name")
	}
	if hostPrefixed && !secure {
		return "", fmt.Errorf("baseline: __Host- cookies require Secure")
	}
	if hostPrefixed && path != "/" {
		return "", fmt.Errorf("baseline: __Host- cookies require Path=/")
	}
	if options.MaxAge != nil && *options.MaxAge < 0 {
		return "", fmt.Errorf("baseline: Max-Age must not be negative")
	}

	parts := []string{name + "=" + value}
	if options.MaxAge != nil {
		parts = append(parts, "Max-Age="+strconv.Itoa(*options.MaxAge))
	}
	parts = append(parts, "Path="+path)
	if secure {
		parts = append(parts, "Secure")
	}
	if httpOnly {
		parts = append(parts, "HttpOnly")
	}
	parts = append(parts, "SameSite="+sameSite)
	return strings.Join(parts, "; "), nil
}
