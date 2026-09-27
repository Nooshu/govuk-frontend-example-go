package baseline

import (
	"fmt"
	"net/http"
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

// SetCookie builds a Set-Cookie header value from the policy defaults.
//
// Validation (SameSite=None requires Secure, __Host- rules, Path shape) stays here so a cookie
// browsers would silently drop is an error. Serialization uses [http.Cookie.String] from the
// standard library.
func (p *Policy) SetCookie(name, value string, options CookieOptions) (string, error) {
	sameSite := options.SameSite
	if sameSite == "" {
		sameSite = p.Cookie.SameSite
	}
	mode, err := sameSiteMode(sameSite)
	if err != nil {
		return "", err
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

	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: mode,
	}
	if options.MaxAge != nil {
		cookie.MaxAge = *options.MaxAge
	}
	if err := cookie.Valid(); err != nil {
		return "", fmt.Errorf("baseline: invalid cookie: %w", err)
	}
	return cookie.String(), nil
}

func sameSiteMode(value string) (http.SameSite, error) {
	switch value {
	case "Lax":
		return http.SameSiteLaxMode, nil
	case "Strict":
		return http.SameSiteStrictMode, nil
	case "None":
		return http.SameSiteNoneMode, nil
	default:
		return 0, fmt.Errorf("baseline: invalid SameSite: %q", value)
	}
}
