package archiver

/*
 * archiver_test.go
 * Tests for archiver.go
 * By J. Stuart McMurray
 * Created 20240812
 * Last Modified 20240819
 */

import (
	"embed"
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"testing"
)

//go:embed testdata
var testdata embed.FS

// newTestArchiver returns an archiver corresponding to the JSON at path in tfs.
func newTestArchiver(t *testing.T, tfs fs.FS, path string) Archiver {
	/* Open the file with the archiver JSON. */
	f, err := tfs.Open(path)
	if nil != err {
		t.Fatalf("Error opening %s: %s", path, err)
	}
	defer f.Close()
	/* Turn into a proper archiver. */
	var a Archiver
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&a); nil != err {
		t.Fatalf(
			"Error unJSONing Archiver from %s: %s",
			path,
			err,
		)
	}
	a.fs = tfs
	return a
}

// subFS calls fs.Sub on testdata for the given path.
func subFS(t *testing.T, tdpath string) fs.FS {
	tfs, err := fs.Sub(testdata, path.Join("testdata", tdpath))
	if nil != err {
		t.Fatalf("Could not get testdata for %s: %s", tdpath, err)
	}
	return tfs
}

func TestArchiverAddPathsFromFile(t *testing.T) {
	type testC struct {
		have     []string
		haveFile string
		want     []string
	}
	cs := map[string]testC{
		"only_file": testC{
			haveFile: `foo
			bar
			
			tridge
			`,
			want: []string{"foo", "bar", "tridge"},
		},
		"no_file": testC{
			have: []string{"foo", "bar"},
			want: []string{"foo", "bar"},
		},
		"file_and_existing": testC{
			have: []string{"foo", "bar", "baaz"},
			haveFile: `foo
			bar
			
			tridge
			`,
			want: []string{"foo", "bar", "baaz", "tridge"},
		},
		"nothing": testC{},
	}
	for name, c := range cs {
		t.Run(name, func(t *testing.T) {
			a := Archiver{Paths: c.have}
			fn := filepath.Join(t.TempDir(), "paths")
			if err := os.WriteFile(
				fn,
				[]byte(c.haveFile),
				0600,
			); nil != err {
				t.Fatalf(
					"Error writing paths file %s: %s",
					fn,
					err,
				)
			}
			if err := a.AddPathsFromFile(fn); nil != err {
				t.Fatalf("Error adding paths: %s", err)
			}
			if !slices.Equal(a.Paths, c.want) {
				t.Fatalf(
					"Paths incorrect:\n got: %s\nwant: %s",
					a.Paths,
					c.want,
				)
			}
		})
	}
}

func TestArchiverToFromHostPath(t *testing.T) {
	type want struct {
		safe   string
		unsafe string
	}
	cases := map[string]want{
		"/p":          want{safe: "p", unsafe: "/p"},
		"//////////p": want{safe: "p", unsafe: "//////////p"},
		"/":           want{safe: ".", unsafe: "/"},
		"//":          want{safe: ".", unsafe: "//"},
		"p/foo":       want{safe: "p/foo", unsafe: "p/foo"},
		"/p/foo":      want{safe: "p/foo", unsafe: "/p/foo"},
		"/p/":         want{safe: "p", unsafe: "/p/"},
		"../../p":     want{safe: "p", unsafe: "../../p"},
		"/../../p":    want{safe: "p", unsafe: "/../../p"},
	}
	t.Run("ToHostPath/safePaths", func(t *testing.T) {
		a := Archiver{}
		for have, wants := range cases {
			got := a.ToHostPath(have)
			want := filepath.FromSlash(wants.safe)
			if got != want {
				t.Errorf(
					"Incorrect:\n"+
						"have: %s"+
						"\n got: %s\n"+
						"want: %s",
					have,
					got,
					want,
				)
			}
		}
	})
	t.Run("to/no_safePaths", func(t *testing.T) {
		a := Archiver{UnsafePaths: true}
		for have, wants := range cases {
			got := a.ToHostPath(have)
			want := filepath.FromSlash(wants.unsafe)
			if got != want {
				t.Errorf(
					"Incorrect:\n"+
						"have: %s"+
						"\n got: %s\n"+
						"want: %s",
					have,
					got,
					want,
				)
			}
		}
	})
	t.Run("from/safePaths", func(t *testing.T) {
		a := Archiver{}
		for have, wants := range cases {
			have := filepath.FromSlash(have)
			got := a.FromHostPath(have)
			want := wants.safe
			if got != want {
				t.Errorf(
					"Incorrect:\n"+
						"have: %s"+
						"\n got: %s\n"+
						"want: %s",
					have,
					got,
					want,
				)
			}
		}
	})
	t.Run("from/no_safePaths", func(t *testing.T) {
		a := Archiver{UnsafePaths: true}
		for have, wants := range cases {
			have := filepath.FromSlash(have)
			got := a.FromHostPath(have)
			want := wants.unsafe
			if got != want {
				t.Errorf(
					"Incorrect:\n"+
						"have: %s"+
						"\n got: %s\n"+
						"want: %s",
					have,
					got,
					want,
				)
			}
		}
	})
}
