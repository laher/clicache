package main

import (
	"testing"
	"time"
)

func TestHash(t *testing.T) {

	h1 := hash([]string{"a", "b"})
	h2 := hash([]string{"a", "bc"})
	h3 := hash([]string{"a", "b"})
	if h1 == h2 {
		t.Error("Hashes should differ")
	}
	if h1 != h3 {
		t.Error("Hashes should be equal")
	}
}

func TestFile(t *testing.T) {
	hash := "123"
	ti := time.Time{}

	f1, _ := file(hash, ti, 5*time.Minute)
	f2, _ := file("124", ti, 5*time.Minute)
	if f1 == f2 {
		t.Error("Filenames should differ", f1, f2)
	}
	f3, _ := file(hash, ti.Add(10*time.Minute), 5*time.Minute)
	if f1 == f3 {
		t.Error("Filenames should differ", f1, f3)
	}
	f4, _ := file(hash, ti.Add(5*time.Second), 5*time.Minute)
	if f1 != f4 {
		t.Error("Filenames should not differ", f1, f4)
	}
}
