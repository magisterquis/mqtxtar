package main

/*
 * create.go
 * Create an archive
 * By J. Stuart McMurray
 * Created 20230516
 * Last Modified 20231218
 */

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/tools/txtar"
)

// Create creates an archive.
func Create(comment string) error {
	/* archive to create. */
	var (
		archive = txtar.Archive{
			Comment: []byte(comment),
		}
		seen = make(map[string]struct{}) /* Deduping. */
	)

	/* The archive comment can't have a marker line or txtar.Format will
	be sad. */
	if hasMarkerLine(string(archive.Comment)) {
		return errors.New("comment cannot have a txtar marker line")
	}

	/* Add each named file. */
	for _, file := range Files {
		if err := addToArchive(&archive, file, seen); nil != err {
			return fmt.Errorf("adding %s: %w", file, err)
		}
	}

	/* Work out where to send the archive. */
	of := os.Stdout
	if StdioFileName != ArchiveFile {
		var err error
		if of, err = os.Create(ArchiveFile); nil != err {
			return fmt.Errorf("opening archive file: %w", err)
		}
		defer of.Close()
	}
	w := io.Writer(of)

	/* Optionally add a layer of compression. */
	if compression {
		zw := gzip.NewWriter(of)
		zw.ModTime = time.Now()
		defer zw.Close()
		w = zw
	}

	/* Emit the archive itself. */
	if _, err := w.Write(txtar.Format(&archive)); nil != err {
		return fmt.Errorf("writing archive: %w", err)
	}

	return nil
}

// hasMarkerLine returns true if s has a txtar marker line. */
func hasMarkerLine(s string) bool {
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "-- ") && strings.HasSuffix(l, " --") {
			return true
		}
	}
	if err := scanner.Err(); nil != err {
		panic(fmt.Sprintf("scanner.Scan on strings.Reader: %s", err))
	}
	return false

}

// addToArchive adds name to archive, deduping with seen.  If name is a file,
// it is added directly.  If name is a directory, files in it are recursively
// added.
func addToArchive(
	archive *txtar.Archive,
	file string,
	seen map[string]struct{},
) error {
	/* wdf is an fs.WalkDirFunc which does the actual adding. */
	wdf := func(path string, d fs.DirEntry, err error) error {
		/* Warn the user if something fails. */
		if nil != err {
			Errorf("Error accessing %s: %s", path, err)
			return nil
		}
		/* Only care about regular files. */
		if !d.Type().IsRegular() {
			return nil
		}

		/* Skip files we don't want. */
		if Excluded(path) {
			return nil
		}

		/* Try to not exploit anybody. */
		cpath := Clean(path)

		/* Don't add files more than once. */
		if _, ok := seen[cpath]; ok {
			return nil
		}
		seen[cpath] = struct{}{}
		if cpath != path {
			log.Printf("Cleaned %q to %q", path, cpath)
		}

		/* Add the file to the archive. */
		b, err := os.ReadFile(path)
		if nil != err {
			Errorf("Error reading %s: %s", path, err)
			return nil
		}
		archive.Files = append(archive.Files, txtar.File{
			Name: cpath,
			Data: b,
		})
		VPrintf("%s\n", path)

		return nil
	}

	/* Try to add the file. */
	return filepath.WalkDir(file, wdf)
}
