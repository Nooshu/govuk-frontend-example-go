package httpx

import (
	"errors"
	"io"
	"testing"
)

type failCloser struct {
	writeErr error
	closeErr error
}

func (f failCloser) Write([]byte) (int, error) {
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	return 1, nil
}

func (f failCloser) Close() error { return f.closeErr }

func TestCompressFallsBackWhenEncodingFails(t *testing.T) {
	body := []byte("hello govuk hello govuk hello govuk")
	previousBrotli := newBrotliWriter
	previousGzip := newGzipWriter
	t.Cleanup(func() {
		newBrotliWriter = previousBrotli
		newGzipWriter = previousGzip
	})

	newBrotliWriter = func(io.Writer) io.WriteCloser {
		return failCloser{writeErr: errors.New("brotli write")}
	}
	compressed, coding := Compress(body, "br", "text/plain")
	if coding != "" || string(compressed) != string(body) {
		t.Fatalf("brotli write failure = %q %q", coding, compressed)
	}

	newBrotliWriter = func(io.Writer) io.WriteCloser {
		return failCloser{closeErr: errors.New("brotli close")}
	}
	compressed, coding = Compress(body, "br", "text/plain")
	if coding != "" || string(compressed) != string(body) {
		t.Fatalf("brotli close failure = %q %q", coding, compressed)
	}

	newGzipWriter = func(io.Writer) io.WriteCloser {
		return failCloser{writeErr: errors.New("gzip write")}
	}
	compressed, coding = Compress(body, "gzip", "text/html")
	if coding != "" || string(compressed) != string(body) {
		t.Fatalf("gzip write failure = %q %q", coding, compressed)
	}

	newGzipWriter = func(io.Writer) io.WriteCloser {
		return failCloser{closeErr: errors.New("gzip close")}
	}
	compressed, coding = Compress(body, "gzip", "text/html")
	if coding != "" || string(compressed) != string(body) {
		t.Fatalf("gzip close failure = %q %q", coding, compressed)
	}
}
