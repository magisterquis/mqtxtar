package archiver

/*
 * create.go
 * Create a new archive
 * By J. Stuart McMurray
 * Created 20240812
 * Last Modified 20240819
 */

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"golang.org/x/tools/txtar"
)

// Create creates an archive.
func (a Archiver) Create() error {
	/* Archive around which this is one big wrapper. */
	ta := &txtar.Archive{Comment: []byte(a.Comment)}

	/* Add files to the archive, as we get them. */
	for _, path := range a.Paths {
		if err := a.addToArchive(ta, path); nil != err {
			return fmt.Errorf("adding %q: %w", path, err)
		}
	}

	/* Work out how to write this thing. */
	var w io.Writer = os.Stdout
	if "" != a.Filename {
		/* Write to a file if we have a filename. */
		f, err := os.OpenFile(
			a.Filename,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC,
			CreatePerms,
		)
		if nil != err {
			return fmt.Errorf(
				"creating archive file %s: %w",
				a.Filename,
				err,
			)
		}
		defer f.Close()
		w = f
	}
	/* Wrap in a zipper if we're zipping. */
	if a.WithGzip {
		z := gzip.NewWriter(w)
		defer z.Close()
		w = z
	}

	/* Finally, write out the archive. */
	if _, err := w.Write(txtar.Format(ta)); nil != err {
		return fmt.Errorf("writing archive: %w", err)
	}

	return nil
}

// addToArchive adds the files under path to ta.
func (a Archiver) addToArchive(ta *txtar.Archive, path string) error {
	wdf := func(
		path string,
		d fs.DirEntry,
		err error,
	) error {
		/* If we're skipping this path, do so before examining the
		error, to provide a nice way to not error out on unreadable
		directories. */
		if excl, err := a.isExcluded(path); nil != err {
			return fmt.Errorf(
				"checking if %s is excluded: %w",
				path,
				err,
			)
		} else if excl {
			return nil
		}
		/* If we couldn't read whatever this is, it's a problem. */
		if nil != err {
			return err
		}
		/* Don't really care about non-regular files. */
		if !d.Type().IsRegular() {
			return nil
		}
		/* Add this file, removing any previous ones with the same
		name first. */
		var b []byte
		if nil != a.fs { /* Slurp file. */
			b, err = fs.ReadFile(a.fs, path)
		} else {
			b, err = os.ReadFile(path)
		}
		if nil != err {
			return fmt.Errorf("reading %s: %w", path, err)
		}
		path = a.FromHostPath(path)   /* txtarify path. */
		ta.Files = slices.DeleteFunc( /* Dedupe. */
			ta.Files,
			func(f txtar.File) bool {
				return f.Name == path
			},
		)
		ta.Files = append(ta.Files, txtar.File{ /* Add. */
			Name: path,
			Data: b,
		})
		if a.Verbose { /* Log. */
			fmt.Fprintf(os.Stderr, "%s\n", path)
		}
		return nil
	}
	if nil != a.fs {
		return fs.WalkDir(a.fs, path, wdf)
	} else {
		return filepath.WalkDir(path, wdf)
	}
}
