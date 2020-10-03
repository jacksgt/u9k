package models

import (
	"testing"
	"time"

	"u9k/types"
)

func TestPrettyExpiresIn(t *testing.T) {
	var got, want string
	f := File{}
	f.CreateTimestamp = time.Now()

	f.Expire = types.Duration(time.Hour*24*7 + time.Hour)
	got = f.PrettyExpiresIn()
	want = "1 week"
	if got != want {
		t.Errorf("PrettyExpiresIn incorrect, got: %s, want: %s.", got, want)
	}

	f.Expire = types.Duration(0)
	got = f.PrettyExpiresIn()
	want = "Never"
	if got != want {
		t.Errorf("PrettyExpiresIn incorrect, got: %s, want: %s.", got, want)
	}
}

func TestPrettyFileSize(t *testing.T) {
	var got, want string
	f := File{}

	f.Size = 1024 * 1024
	got = f.PrettyFileSize()
	want = "1.0 MiB"
	if got != want {
		t.Errorf("PrettyFileSize incorrect, got: %s, want: %s.", got, want)
	}

	f.Size = 1024 * 1024 * 3947
	got = f.PrettyFileSize()
	want = "3.9 GiB"
	if got != want {
		t.Errorf("PrettyFileSize incorrect, got: %s, want: %s.", got, want)
	}

	f.Size = 1777
	got = f.PrettyFileSize()
	want = "1.7 KiB"
	if got != want {
		t.Errorf("PrettyFileSize incorrect, got: %s, want: %s.", got, want)
	}
}
