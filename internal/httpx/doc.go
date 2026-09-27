// Package httpx holds transport concerns that sit either side of a page.
//
// That includes bounded body reads, cookie helpers, the on-disk / in-memory asset
// map for Frontend static files and the compiled stylesheet, and Brotli-first
// response compression with Gzip as the fallback when the client does not
// advertise br.
package httpx
