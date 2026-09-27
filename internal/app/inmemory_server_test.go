package app_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"
)

// TestInMemoryServerServesHealth uses Go 1.27's httptest.NewTestServer (fake network) so the
// example handler can be exercised through a real http.Client without binding a TCP port.
func TestInMemoryServerServesHealth(t *testing.T) {
	c := newClient(t)
	server := httptest.NewTestServer(t, c.handler)

	resp, err := server.Client().Get(server.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK || string(body) != "ok" {
		t.Fatalf("health = %d %q", resp.StatusCode, body)
	}
}

// TestClientTimeoutUsesSynctestClock proves a client timeout fires under synctest's fake clock
// when the handler sleeps longer than the deadline — without waiting in real time.
func TestClientTimeoutUsesSynctestClock(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			synctest.Sleep(2 * time.Second)
			_, _ = io.WriteString(w, "late")
		})
		server := httptest.NewTestServer(t, handler)
		client := server.Client()
		client.Timeout = time.Second

		_, err := client.Get(server.URL)
		if err == nil {
			t.Fatal("expected the client to time out")
		}
	})
}
