package app

// Buffered write path: ETag, baseline headers, compression, cookies.

import (
	"net/http"
	"os"

	"github.com/Nooshu/govuk-frontend-example-go/internal/baseline"
	"github.com/Nooshu/govuk-frontend-example-go/internal/httpx"
	"github.com/Nooshu/govuk-frontend-example-go/internal/pages"
	"github.com/Nooshu/govuk-frontend-example-go/internal/session"
)

// writePage renders a view and sends it with the document baseline.
//
// A page that shows the applicant's answers is sent as a sensitive document, which is no-store:
// their name, address, and date of birth must not be left in a shared cache or in the back/
// forward cache of a shared computer. Only public pages get an ETag, because a validator is
// only useful for something that may be cached.
func (a *App) writePage(w http.ResponseWriter, r *http.Request, view pages.View, current *session.Session) {
	body, err := a.renderer.Render(view, current, a.currentPath(r))
	if err != nil {
		a.logger.Error("rendering failed", "template", view.Template, "error", err)
		if view.Template == "problem" {
			a.writeText(w, r, http.StatusInternalServerError, "Sorry, there is a problem with the service")
			return
		}
		a.writePage(w, r, problemView(http.StatusInternalServerError), current)
		return
	}
	kind := baseline.KindDocument
	if view.Personal {
		kind = baseline.KindSensitiveDocument
	}
	header := w.Header()
	if err := a.policy.ApplyHeaders(header, baseline.HeaderOptions{
		Kind:            kind,
		SecureTransport: requestIsSecure(r),
		SetsCookie:      true,
		Preload:         a.assets.Page().Preloads,
	}); err != nil {
		a.logger.Error("applying response headers failed", "error", err)
		a.writeText(w, r, http.StatusInternalServerError, "Sorry, there is a problem with the service")
		return
	}
	a.setSessionCookie(w, r, current)

	raw := []byte(body)
	if kind == baseline.KindDocument {
		etag := baseline.StrongETag(raw)
		header.Set("ETag", etag)
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	a.writeBody(w, r, view.Status, raw, header.Get("Content-Type"))
}

// writeRedirect sends a 303 See Other, the status that turns a completed POST into a GET so a
// refresh cannot resubmit the answer.
func (a *App) writeRedirect(w http.ResponseWriter, r *http.Request, location string, current *session.Session) {
	header := w.Header()
	if err := a.policy.ApplyHeaders(header, baseline.HeaderOptions{
		Kind:            baseline.KindDocument,
		SecureTransport: requestIsSecure(r),
		SetsCookie:      true,
	}); err != nil {
		a.logger.Error("applying response headers failed", "error", err)
	}
	a.setSessionCookie(w, r, current)
	header.Set("Location", location)
	w.WriteHeader(http.StatusSeeOther)
}

func (a *App) writeRaw(w http.ResponseWriter, r *http.Request, raw rawResponse) {
	header := w.Header()
	if err := a.policy.ApplyHeaders(header, baseline.HeaderOptions{
		Kind:            raw.kind,
		SecureTransport: requestIsSecure(r),
		ContentType:     raw.contentType,
	}); err != nil {
		a.logger.Error("applying response headers failed", "error", err)
		a.writeText(w, r, http.StatusInternalServerError, "Sorry, there is a problem with the service")
		return
	}
	a.writeBody(w, r, raw.status, raw.body, header.Get("Content-Type"))
}

func (a *App) writeText(w http.ResponseWriter, r *http.Request, status int, body string) {
	header := w.Header()
	if err := a.policy.ApplyHeaders(header, baseline.HeaderOptions{
		Kind:            baseline.KindStaticAsset,
		SecureTransport: requestIsSecure(r),
		ContentType:     "text/plain; charset=utf-8",
	}); err != nil {
		a.logger.Error("applying response headers failed", "error", err)
	}
	a.writeBody(w, r, status, []byte(body), header.Get("Content-Type"))
}

func (a *App) writeAsset(w http.ResponseWriter, r *http.Request, asset httpx.Asset) {
	body := asset.Body
	if body == nil {
		read, err := os.ReadFile(asset.FilePath)
		if err != nil {
			a.writeText(w, r, http.StatusNotFound, "Not found")
			return
		}
		body = read
	}
	header := w.Header()
	if err := a.policy.ApplyHeaders(header, baseline.HeaderOptions{
		Kind:            asset.Kind,
		SecureTransport: requestIsSecure(r),
		ContentType:     asset.ContentType,
	}); err != nil {
		a.logger.Error("applying response headers failed", "error", err)
		a.writeText(w, r, http.StatusInternalServerError, "Sorry, there is a problem with the service")
		return
	}
	a.writeBody(w, r, http.StatusOK, body, header.Get("Content-Type"))
}

// writeBody compresses the body when it is worth it and the client accepts it.
//
// Brotli is the standard from baseline/policy.json; Gzip is used only when the client does not
// advertise br. Vary: Accept-Encoding is already set by the baseline, so a cache never serves an
// encoding the next client cannot read.
func (a *App) writeBody(w http.ResponseWriter, r *http.Request, status int, body []byte, contentType string) {
	encoded, encoding := httpx.Compress(body, r.Header.Get("Accept-Encoding"), contentType)
	if encoding != "" {
		w.Header().Set("Content-Encoding", encoding)
	}
	w.WriteHeader(status)
	if r.Method == http.MethodHead {
		return
	}
	if _, err := w.Write(encoded); err != nil {
		a.logger.Debug("writing the response body failed", "error", err)
	}
}

func (a *App) setSessionCookie(w http.ResponseWriter, r *http.Request, current *session.Session) {
	cookie, err := a.sessionSetCookie(r, current)
	if err != nil {
		a.logger.Error("building the session cookie failed", "error", err)
		return
	}
	w.Header().Add("Set-Cookie", cookie)
}
