package render

import (
	"net/http"
	"testing"
)

func TestAcceptsType(t *testing.T) {
	var want, got bool
	r := &http.Request{}
	r.Header = make(http.Header)

	r.Header.Set("Accept", "text/html")
	want = true
	got = acceptsType(r, "text/html")
	if got != want {
		t.Errorf("AcceptsType incorrect, got: %v, want: %v.", got, want)
	}

	r.Header.Set("Accept", "*/*")
	want = true
	got = acceptsType(r, "text/html")
	if got != want {
		t.Errorf("AcceptsType incorrect, got: %v, want: %v.", got, want)
	}

	// if the client does not set an explicit Accept header, only allow text/html
	r.Header.Set("Accept", "*/*")
	want = false
	got = acceptsType(r, "plain/text")
	if got != want {
		t.Errorf("AcceptsType incorrect, got: %v, want: %v.", got, want)
	}

	r.Header.Set("Accept", "image/*")
	want = false
	got = acceptsType(r, "plain/text")
	if got != want {
		t.Errorf("AcceptsType incorrect, got: %v, want: %v.", got, want)
	}

	r.Header.Set("Accept", "text/html, application/xhtml+xml, application/xml;q=0.9, */*;q=0.8")
	want = true
	got = acceptsType(r, "text/html")
	if got != want {
		t.Errorf("AcceptsType incorrect, got: %v, want: %v.", got, want)
	}

	// TODO: still need to implement wildcards
	// r.Header.Set("Accept", "image/*")
	// want = true
	// got = acceptsType(r, "image/webp")
	// if got != want {
	// 	t.Errorf("AcceptsType incorrect, got: %v, want: %v.", got, want)
	// }
}
