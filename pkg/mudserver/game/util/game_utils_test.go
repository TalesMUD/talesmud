package util

import (
	"reflect"
	"testing"
)

func TestRemoveStringFromSliceRemovesLastElement(t *testing.T) {
	got := RemoveStringFromSlice([]string{"north", "south", "east"}, "east")
	want := []string{"north", "south"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}

func TestRemoveStringFromSliceRemovesOnlyMatch(t *testing.T) {
	got := RemoveStringFromSlice([]string{"north", "south", "east"}, "south")
	want := []string{"north", "east"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}
