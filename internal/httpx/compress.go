package httpx

// Brotli-first response compression with Gzip fallback.

import (
	"bytes"
	"compress/gzip"
	"io"
	"regexp"
	"strings"

	"github.com/andybalholm/brotli"
)

// newBrotliWriter and newGzipWriter build the encoders. Tests replace them with writers that
// fail, which is how the uncompressed fallback is reached: bytes.Buffer itself never returns
// a write error.
var (
	newBrotliWriter = func(w io.Writer) io.WriteCloser { return brotli.NewWriter(w) }
	newGzipWriter   = func(w io.Writer) io.WriteCloser { return gzip.NewWriter(w) }
)

var compressible = regexp.MustCompile(`^(text/|application/(javascript|json)|image/svg\+xml)`)

// Compressible reports whether a response body is worth compressing.
//
// Images, fonts, and other already-compressed formats are left alone: re-compressing them costs
// CPU and usually makes them larger.
func Compressible(contentType string) bool {
	return compressible.MatchString(contentType)
}

// Compress encodes a response body for the client.
//
// Brotli is the standard from baseline/policy.json and is used whenever the client advertises
// br. Gzip is only the fallback for clients that do not. The returned coding is empty when the
// body was left uncompressed, in which case no Content-Encoding should be sent.
func Compress(body []byte, acceptEncoding, contentType string) ([]byte, string) {
	if len(body) == 0 || !Compressible(contentType) {
		return body, ""
	}
	accepted := encodings(acceptEncoding)
	if accepted["br"] {
		compressed, err := compressWith(body, newBrotliWriter)
		if err != nil {
			return body, ""
		}
		return compressed, "br"
	}
	if accepted["gzip"] {
		compressed, err := compressWith(body, newGzipWriter)
		if err != nil {
			return body, ""
		}
		return compressed, "gzip"
	}
	return body, ""
}

func compressWith(body []byte, newWriter func(io.Writer) io.WriteCloser) ([]byte, error) {
	var out bytes.Buffer
	writer := newWriter(&out)
	if _, err := writer.Write(body); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func encodings(header string) map[string]bool {
	found := map[string]bool{}
	for part := range strings.SplitSeq(header, ",") {
		token := strings.TrimSpace(strings.ToLower(part))
		if index := strings.Index(token, ";"); index != -1 {
			token = strings.TrimSpace(token[:index])
		}
		if token != "" {
			found[token] = true
		}
	}
	return found
}
