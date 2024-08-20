package archiver

/*
 * create_test.go
 * Tests for create.go
 * By J. Stuart McMurray
 * Created 20240812
 * Last Modified 20240812
 */

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestArchiverCreate tests archive creation
func TestArchiverCreate(t *testing.T) {
	tfs := subFS(t, "archiver/create")
	ajfns, err := fs.Glob(tfs, "*.json")
	if nil != err {
		t.Fatalf("Error getting test cases: %s", err)
	}
	for _, ajfn := range ajfns {
		t.Run(
			strings.TrimSuffix(ajfn, ".json"),
			archiverCreateTester{tfs, ajfn}.Do,
		)
	}
}

// archiverCreateTester tests an Archiver.Create invocation.
type archiverCreateTester struct {
	tfs  fs.FS
	ajfn string
}

func (act archiverCreateTester) Do(t *testing.T) {
	/* Roll the Archiver to test with. */
	a := newTestArchiver(t, act.tfs, act.ajfn)

	/* Base test name. */
	name := strings.TrimSuffix(act.ajfn, ".json")

	/* Grab the file we're meant to get. */
	wantName := name + ".txtar"
	want, err := fs.ReadFile(act.tfs, wantName)
	if nil != err {
		t.Fatalf("Error reading want file %s: %s", wantName, err)
	}

	/* Try without compression. */
	t.Run("NoGzip", func(t *testing.T) {
		a := a /* Test-local copy. */
		a.Filename = filepath.Join(t.TempDir(), "got.txtar")
		if err := a.Create(); nil != err {
			t.Fatalf("Create failed: %s", err)
		}
		got, err := os.ReadFile(a.Filename)
		if nil != err {
			t.Fatalf("Error reading created file: %s", err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf(
				"Incorrect created archive:\n"+
					"got:\n%s\n"+
					"want:\n%s\n",
				got,
				want,
			)
		}
	})

	/* Try with compression. */
	t.Run("WithGzip", func(t *testing.T) {
		a := a /* Test-local copy. */
		a.Filename = filepath.Join(t.TempDir(), "got.txtar")
		a.WithGzip = true
		if err := a.Create(); nil != err {
			t.Fatalf("Create failed: %s", err)
		}
		gotZ, err := os.ReadFile(a.Filename)
		if nil != err {
			t.Fatalf("Error reading created file: %s", err)
		}

		zr, err := gzip.NewReader(bytes.NewReader(gotZ))
		if nil != err {
			t.Fatalf("Decompressor creation error: %s", err)
		}
		got, err := io.ReadAll(zr)
		if nil != err {
			t.Fatalf("Decompression error: %s", err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf(
				"Incorrect decompressed archive:\n"+
					"got:\n%s\n"+
					"want:\n%s\n",
				got,
				want,
			)
		}
	})
}
