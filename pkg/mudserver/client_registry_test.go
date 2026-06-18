package mudserver

import (
	"strconv"
	"sync"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
)

func TestClientRegistryConcurrentAccess(t *testing.T) {
	registry := newClientRegistry()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := "user-" + strconv.Itoa(i%10)
			registry.Set(id, &Connection{User: &entities.User{Entity: &entities.Entity{ID: id}}})
			_, _ = registry.Get(id)
			registry.ForEach(func(_ string, _ *Connection) {})
			registry.Delete(id)
		}()
	}

	wg.Wait()
}

func TestClientRegistryDeleteIfDoesNotRemoveReplacement(t *testing.T) {
	registry := newClientRegistry()
	oldConnection := &Connection{User: &entities.User{Entity: &entities.Entity{ID: "user-1"}}}
	newConnection := &Connection{User: &entities.User{Entity: &entities.Entity{ID: "user-1"}}}

	registry.Set("user-1", oldConnection)
	registry.Set("user-1", newConnection)

	if registry.DeleteIf("user-1", oldConnection) {
		t.Fatal("expected stale connection delete to be ignored")
	}
	got, ok := registry.Get("user-1")
	if !ok {
		t.Fatal("expected replacement connection to remain registered")
	}
	if got != newConnection {
		t.Fatal("expected current connection to be the replacement")
	}

	if !registry.DeleteIf("user-1", newConnection) {
		t.Fatal("expected current connection delete to succeed")
	}
	if _, ok := registry.Get("user-1"); ok {
		t.Fatal("expected current connection to be deleted")
	}
}
