package client_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"openlist/client"
)

func TestIsNotFound(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{fmt.Errorf("nope"), false},
		{&client.HTTPError{StatusCode: http.StatusNotFound, Body: "missing"}, true},
		{&client.HTTPError{StatusCode: http.StatusBadRequest, Body: "bad"}, false},
		{&client.APIError{Code: 404, Message: "x"}, true},
		{&client.APIError{Code: 500, Message: "object not found"}, true},
		{&client.APIError{Code: 500, Message: "boom"}, false},
		{fmt.Errorf("wrap: %w", &client.APIError{Code: 500, Message: "does not exist"}), true},
	}
	for _, tc := range cases {
		if got := client.IsNotFound(tc.err); got != tc.want {
			t.Errorf("IsNotFound(%v) = %v, want %v", tc.err, got, tc.want)
		}
	}
	if !errors.As(&client.HTTPError{StatusCode: 400}, new(*client.HTTPError)) {
		t.Fatalf("HTTPError should support errors.As")
	}
}
