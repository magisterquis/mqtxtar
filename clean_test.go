package main

/*
 * clean_test.go
 * Tests for clean.go
 * By J. Stuart McMurray
 * Created 20230516
 * Last Modified 20230516
 */

import "testing"

func TestClean(t *testing.T) {
	for _, c := range []struct {
		Have string
		Want string
	}{{
		Have: "a/b",
		Want: "a/b",
	}, {
		Have: "../../foo",
		Want: "foo",
	}, {
		Have: "/",
		Want: "/",
	}, {
		Have: "../foo/../bar",
		Want: "bar",
	}} {
		c := c /* :( */
		t.Run(c.Have, func(t *testing.T) {
			t.Parallel()
			got := Clean(c.Have)
			if got != c.Want {
				t.Errorf("got %q", got)
			}
		})
	}
}
