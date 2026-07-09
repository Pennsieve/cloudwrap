package config

import "testing"

func TestPath(t *testing.T) {
	c := New("staging", "auth-service")
	if got, want := c.Path(), "/staging/auth-service/"; got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}
}
