// Package httpx holds the transport concerns that sit either side of a page: reading form
// posts, reading cookies, serving static assets, and compressing responses.
package httpx

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

// Upload is a file part from a multipart body.
//
// Only the field name and the filename are kept: this example never stores the bytes, and
// holding an unvalidated upload in memory would be a liability rather than a feature.
type Upload struct {
	FieldName string
	Filename  string
}

// Body is a parsed form submission.
type Body struct {
	// Fields holds every value posted for each name, so repeated checkboxes are not collapsed.
	Fields map[string][]string
	// Upload is the file part, when the request had one.
	Upload *Upload
}

// BodyError is a request body that could not be read, with the status to reply with.
type BodyError struct {
	Status  int
	Message string
}

// Error implements error.
func (e *BodyError) Error() string { return e.Message }

// Field returns the first value posted for name, or an empty string.
func (b *Body) Field(name string) string {
	if values := b.Fields[name]; len(values) > 0 {
		return values[0]
	}
	return ""
}

// Values returns every value posted for name.
func (b *Body) Values(name string) []string {
	return b.Fields[name]
}

// ReadBody reads and parses a request body, refusing anything larger than maxBytes.
//
// The limit is applied while reading rather than after, so an oversized post is rejected without
// ever being held in memory.
func ReadBody(request *http.Request, maxBytes int64) (*Body, error) {
	if request.Body == nil {
		return &Body{Fields: map[string][]string{}}, nil
	}
	raw, err := io.ReadAll(io.LimitReader(request.Body, maxBytes+1))
	if err != nil {
		return nil, &BodyError{Status: http.StatusBadRequest, Message: "Could not read the request body"}
	}
	if int64(len(raw)) > maxBytes {
		return nil, &BodyError{Status: http.StatusRequestEntityTooLarge, Message: "Payload too large"}
	}
	return ParseBody(request.Header.Get("Content-Type"), raw, maxBytes)
}

// ParseBody parses a urlencoded or multipart form body.
//
// An empty body is not an error: a form can legitimately post nothing, and the missing CSRF
// token is what rejects it later.
func ParseBody(contentType string, raw []byte, maxBytes int64) (*Body, error) {
	if int64(len(raw)) > maxBytes {
		return nil, &BodyError{Status: http.StatusRequestEntityTooLarge, Message: "Payload too large"}
	}
	if len(raw) == 0 {
		return &Body{Fields: map[string][]string{}}, nil
	}
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, &BodyError{Status: http.StatusUnsupportedMediaType, Message: "Unsupported media type"}
	}
	switch mediaType {
	case "application/x-www-form-urlencoded":
		return parseURLEncoded(string(raw))
	case "multipart/form-data":
		return parseMultipart(params["boundary"], raw, maxBytes)
	default:
		return nil, &BodyError{Status: http.StatusUnsupportedMediaType, Message: "Unsupported media type"}
	}
}

func parseURLEncoded(raw string) (*Body, error) {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return nil, &BodyError{Status: http.StatusBadRequest, Message: "Malformed form body"}
	}
	return &Body{Fields: values}, nil
}

func parseMultipart(boundary string, raw []byte, maxBytes int64) (*Body, error) {
	if boundary == "" {
		return nil, &BodyError{Status: http.StatusBadRequest, Message: "Missing multipart boundary"}
	}
	reader := multipart.NewReader(strings.NewReader(string(raw)), boundary)
	body := &Body{Fields: map[string][]string{}}
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			return body, nil
		}
		if err != nil {
			return nil, &BodyError{Status: http.StatusBadRequest, Message: "Malformed multipart body"}
		}
		name := part.FormName()
		filename := part.FileName()
		if filename != "" {
			body.Upload = &Upload{FieldName: name, Filename: filename}
			if err := drain(part, maxBytes); err != nil {
				return nil, err
			}
			continue
		}
		if name == "" {
			if err := drain(part, maxBytes); err != nil {
				return nil, err
			}
			continue
		}
		value, err := io.ReadAll(io.LimitReader(part, maxBytes))
		if err != nil {
			return nil, &BodyError{Status: http.StatusBadRequest, Message: "Malformed multipart body"}
		}
		body.Fields[name] = append(body.Fields[name], string(value))
	}
}

// drain discards a part's content. The bytes of an upload are never kept, but the part still has
// to be read so the reader can reach the next boundary.
func drain(part io.Reader, maxBytes int64) error {
	if _, err := io.Copy(io.Discard, io.LimitReader(part, maxBytes)); err != nil {
		return &BodyError{Status: http.StatusBadRequest, Message: fmt.Sprintf("Malformed multipart body: %v", err)}
	}
	return nil
}
