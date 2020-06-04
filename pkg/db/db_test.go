package db

import "testing"

func TestNew(t *testing.T) {
	db := New("test")

	if db == nil {
		t.Error("db object is nil")
	}

	if db.Connected {
		t.Error("db should not be connected")
	}

}
