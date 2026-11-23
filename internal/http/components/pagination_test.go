package components

import (
	"reflect"
	"testing"
)

func TestPageWindow(t *testing.T) {
	got := PageWindow(1, 14)
	want := []string{"1", "2", "3", "4", "…", "10", "11", "12", "13", "14"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("page 1 = %v, want %v", got, want)
	}
	got = PageWindow(7, 14)
	want = []string{"1", "2", "3", "4", "…", "6", "7", "8", "…", "10", "11", "12", "13", "14"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("page 7 = %v, want %v", got, want)
	}
	got = PageWindow(2, 5)
	want = []string{"1", "2", "3", "4", "5"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("short = %v, want %v", got, want)
	}
}
