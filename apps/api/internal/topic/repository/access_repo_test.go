package repository

import "testing"

func TestSharedListPredicate(t *testing.T) {
	for _, tt := range []struct {
		table         string
		authenticated bool
		want          string
	}{
		{"", false, "access_scope = 'public'"},
		{"t", false, "t.access_scope = 'public'"},
		{"topic", true, "topic.access_scope IN ('public','login')"},
		{"parent", true, "parent.access_scope IN ('public','login')"},
	} {
		if got := SharedListPredicate(tt.table, tt.authenticated); got != tt.want {
			t.Errorf("got %q, want %q", got, tt.want)
		}
	}
}
