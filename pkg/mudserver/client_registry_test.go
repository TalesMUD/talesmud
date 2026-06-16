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
