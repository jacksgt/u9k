package misc

import "testing"

func TestStringInSlice(t *testing.T) {
	var got, want bool

	got = StringInSlice("hello", []string{"hello", "world"})
	want = true
	if got != want {
		t.Errorf("StringInSlice incorrect, got: %v, want: %v.", got, want)
	}

	got = StringInSlice("hello", []string{})
	want = false
	if got != want {
		t.Errorf("StringInSlice incorrect, got: %v, want: %v.", got, want)
	}

	got = StringInSlice("hello", []string{"foo", "bar", "baz"})
	want = false
	if got != want {
		t.Errorf("StringInSlice incorrect, got: %v, want: %v.", got, want)
	}
}
