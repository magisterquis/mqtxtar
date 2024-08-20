package archiver

/*
 * listextract_test.go
 * Tests for listextract.go
 * By J. Stuart McMurray
 * Created 20240819
 * Last Modified 20240819
 */

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

// TestArchiverListExtract_List tests archive listing
func TestArchiverListExtract_List(t *testing.T) {
	tfs := subFS(t, "archiver/listextract/list")
	ajfns, err := fs.Glob(tfs, "*.json")
	if nil != err {
		t.Fatalf("Error getting test cases: %s", err)
	}
	for _, ajfn := range ajfns {
		t.Run(
			strings.TrimRight(ajfn, ".json"),
			archiverListTester{tfs, ajfn}.Do,
		)
	}
}

// archiverListTester tests an Archiver.List invocation.
type archiverListTester struct {
	tfs  fs.FS
	ajfn string
}

func (act archiverListTester) Do(t *testing.T) {
	/* Base test name. */
	name := strings.TrimSuffix(act.ajfn, ".json")

	/* Roll the Archiver to test with. */
	a := newTestArchiver(t, act.tfs, act.ajfn)
	a.Filename = name + ".txtar"

	/* Grab what we expect. */
	wantName := name + ".list"
	want, err := fs.ReadFile(act.tfs, wantName)
	if nil != err {
		t.Fatalf("Error reading want file %s: %s", wantName, err)
	}

	/* Did it work? */
	buf := new(bytes.Buffer)
	if err := a.ListOrExtract(buf, t.TempDir(), false); nil != err {
		t.Fatalf("Error: %s", err)
	}
	if got := buf.Bytes(); !bytes.Equal(got, want) {
		t.Fatalf("Incorrect listing:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// TestArchiverListExtract_Extract tests archive extraction
func TestArchiverListExtract_Extract(t *testing.T) {
	tfs := subFS(t, "archiver/listextract/extract")
	ajfns, err := fs.Glob(tfs, "*.json")
	if nil != err {
		t.Fatalf("Error getting test cases: %s", err)
	}
	for _, ajfn := range ajfns {
		t.Run(
			strings.TrimRight(ajfn, ".json"),
			archiverExtractTester{tfs, ajfn}.Do,
		)
	}
}

// archiverExtractTester tests an Archiver.Extract invocation.
type archiverExtractTester struct {
	tfs  fs.FS
	ajfn string
}

func (act archiverExtractTester) Do(t *testing.T) {
	/* Base test name. */
	name := strings.TrimSuffix(act.ajfn, ".json")

	/* Roll the Archiver to test with. */
	a := newTestArchiver(t, act.tfs, act.ajfn)
	a.Filename = name + ".txtar"

	/* Try the extraction. */
	td := t.TempDir()
	if err := a.ListOrExtract(io.Discard, td, true); nil != err {
		t.Fatalf("Extraction error: %s", err)
	}

	/* Mapify the txtar archive, for ease of deletion. */
	fb, err := fs.ReadFile(a.fs, a.Filename)
	if nil != err {
		t.Fatalf("Error reading %s: %s", a.Filename, err)
	}
	ta := txtar.Parse(fb)
	if 0 == len(ta.Files) {
		t.Fatalf("No files in archive %s", a.Filename)
	}
	tam := make(map[string][]byte, len(ta.Files))
	for _, f := range ta.Files {
		tam[f.Name] = f.Data
	}

	/* Make sure all the extracted files are in the archive. */
	if err := fs.WalkDir(os.DirFS(td), ".", func(
		path string,
		d fs.DirEntry,
		err error,
	) error {
		/* Make sure we do actually have something to look at. */
		if nil != err {
			return err
		}
		/* We should really only get files and directories. */
		ft := d.Type()
		switch {
		case ft.IsDir(): /* These are normal. */
			return nil
		case ft.IsRegular(): /* Should be one we expect. */
			/* Make sure we should have this file. */
			wantData, ok := tam[path]
			if !ok {
				t.Errorf("Unexpected file %s", path)
				return nil
			}
			/* Make sure the contents are what we expect. */
			fn := filepath.Join(td, path)
			gotData, err := os.ReadFile(fn)
			if nil != err {
				return fmt.Errorf(
					"reading extracted file %s: %s",
					fn,
					err,
				)
			}
			if !bytes.Equal(wantData, gotData) {
				t.Errorf(
					"Incorrect file contents for %s:\n"+
						"got:\n%s\n"+
						"want:\n%s",
					path,
					gotData,
					wantData,
				)
			}
			/* Note we've seen this one. */
			delete(tam, path)
		default: /* Shouldn't have any of these. */
			t.Errorf("Unexpected file type for %s: %s", path, ft)
		}
		return nil
	}); nil != err {
		t.Fatalf("Error walking extracted files directory: %s", err)
	}

	/* Make sure we extracted everything. */
	for _, n := range slices.Sorted(maps.Keys(tam)) {
		t.Errorf("Did not extract %s", n)
	}
}
