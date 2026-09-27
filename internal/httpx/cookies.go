package httpx

// Request cookie helpers built on net/http.

import (
	"net/http"
)

// CookieValue returns the value of the named cookie on the request, or "" when it is absent.
//
// It uses [http.Request.Cookie] from the standard library rather than parsing the Cookie header
// by hand. Session ids from this service are hex and do not need percent-decoding.
func CookieValue(request *http.Request, name string) string {
	cookie, err := request.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}
