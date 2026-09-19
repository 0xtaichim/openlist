package remotepath

import "testing"

func TestClean(t *testing.T) {
	cases := map[string]string{
		"":          "/",
		".":         "/",
		"/":         "/",
		"/Movies/":  "/Movies",
		"Movies":    "/Movies",
		"/./a/../b": "/b",
	}
	for in, want := range cases {
		if got := Clean(in); got != want {
			t.Errorf("Clean(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRel(t *testing.T) {
	got, err := Rel("/Movies", "/Movies/A/B.mkv")
	if err != nil {
		t.Fatal(err)
	}
	if got != "A/B.mkv" {
		t.Fatalf("got %q", got)
	}

	got, err = Rel("/", "/A.mkv")
	if err != nil || got != "A.mkv" {
		t.Fatalf("root rel: %q %v", got, err)
	}

	if _, err := Rel("/Movies", "/Other/a.mkv"); err == nil {
		t.Fatalf("expected error for path outside root")
	}
}
