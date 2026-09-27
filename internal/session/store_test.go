package session_test

import (
	"sync"
	"testing"
	"testing/synctest"

	"github.com/Nooshu/govuk-frontend-example-go/internal/session"
)

func TestMemoryStoreCreateGetSave(t *testing.T) {
	store := session.NewMemoryStore()
	created := store.Create()
	if created.ID == "" || created.CSRF == "" {
		t.Fatalf("expected id and csrf, got %#v", created)
	}
	if created.ID == created.CSRF {
		t.Fatal("id and csrf should differ")
	}

	got, ok := store.Get(created.ID)
	if !ok || got != created {
		t.Fatalf("Get after Create: ok=%v got=%p want=%p", ok, got, created)
	}

	if _, ok := store.Get("missing"); ok {
		t.Fatal("expected missing id to be absent")
	}

	created.CookieChoice = session.ChoiceAccept
	store.Save(created)
	got, ok = store.Get(created.ID)
	if !ok || got.CookieChoice != session.ChoiceAccept {
		t.Fatalf("Save did not persist choice: %#v", got)
	}
}

func TestNewIsNotStored(t *testing.T) {
	store := session.NewMemoryStore()
	fresh := session.New()
	if _, ok := store.Get(fresh.ID); ok {
		t.Fatal("New must not insert into a store")
	}
}

func TestReferenceFor(t *testing.T) {
	if got := session.ReferenceFor("abcdef123456"); got != "RLABCDEF" {
		t.Fatalf("got %q", got)
	}
	if got := session.ReferenceFor("ab"); got != "RLAB" {
		t.Fatalf("short id: got %q", got)
	}
}

// TestMemoryStoreConcurrentUsesSynctest exercises Create/Get/Save from many goroutines inside a
// synctest bubble so the race detector and scheduler run without wall-clock waits.
func TestMemoryStoreConcurrentUsesSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		store := session.NewMemoryStore()
		const workers = 32
		var wg sync.WaitGroup
		wg.Add(workers)
		for range workers {
			go func() {
				defer wg.Done()
				created := store.Create()
				created.CookieChoice = session.ChoiceAccept
				store.Save(created)
				got, ok := store.Get(created.ID)
				if !ok || got.CookieChoice != session.ChoiceAccept {
					t.Errorf("session %q missing after save", created.ID)
				}
			}()
		}
		wg.Wait()
	})
}
