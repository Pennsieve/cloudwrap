package params

import (
	"reflect"
	"testing"
)

func TestEnvKey(t *testing.T) {
	cases := map[string]string{
		"/staging/svc/one-key":       "ONE_KEY",
		"/staging/svc/two":           "TWO",
		"/dev/auth-service/db-host":  "DB_HOST",
		"/dev/auth-service/AlreadyU": "ALREADYU",
		"no-slashes-key":             "NO_SLASHES_KEY",
		"/trailing/":                 "",
	}
	for name, want := range cases {
		if got := EnvKey(name); got != want {
			t.Errorf("EnvKey(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestShortName(t *testing.T) {
	if got := ShortName("/staging/svc/one-key"); got != "one-key" {
		t.Errorf("ShortName = %q, want %q", got, "one-key")
	}
	if got := ShortName("bare"); got != "bare" {
		t.Errorf("ShortName = %q, want %q", got, "bare")
	}
}

func TestMergeLeftWins(t *testing.T) {
	// service A (left) and service B (right) both define "shared". A wins.
	perService := [][]Parameter{
		{
			{Name: "/env/a/one-key", Value: "a1"},
			{Name: "/env/a/shared", Value: "from-a"},
		},
		{
			{Name: "/env/b/shared", Value: "from-b"},
			{Name: "/env/b/two", Value: "b2"},
		},
	}

	got := Merge(perService)
	want := []Pair{
		{Key: "ONE_KEY", Value: "a1"},
		{Key: "SHARED", Value: "from-a"},
		{Key: "TWO", Value: "b2"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Merge = %+v, want %+v", got, want)
	}
}

func TestMergePreservesFirstSeenOrder(t *testing.T) {
	perService := [][]Parameter{
		{
			{Name: "/env/a/zebra", Value: "z"},
			{Name: "/env/a/apple", Value: "a"},
		},
	}
	got := Merge(perService)
	if len(got) != 2 || got[0].Key != "ZEBRA" || got[1].Key != "APPLE" {
		t.Errorf("Merge did not preserve first-seen order: %+v", got)
	}
}

func TestMergeEmpty(t *testing.T) {
	if got := Merge(nil); len(got) != 0 {
		t.Errorf("Merge(nil) = %+v, want empty", got)
	}
}
