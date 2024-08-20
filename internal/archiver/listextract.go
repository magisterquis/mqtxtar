package archiver

/*
 * listextract.go
 * List and/or extract archive contents
 * By J. Stuart McMurray
 * Created 20240813
 * Last Modified 20240819
 */

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/tools/txtar"
)

const (
	// CreateFilePerms are the permissions with which we create files.
	CreateFilePerms = 0o0644
	// CreateDirPerms are the permissions which which we create directories.
	CreateDirPerms = 0o0755
)

// ListOrExtract lists and/or extracts the contents of a's archive file,
// subject to globbing and file list globbing.  Listing output goes to w.
// files will be extracted to where, which may be "".
func (a Archiver) ListOrExtract(
	w io.Writer,
	where string,
	doExtract bool,
) error {
	/* Slurp file or stdin. */
	var (
		b   []byte
		err error
	)
	if "" == a.Filename { /* Just stdin. */
		if b, err = io.ReadAll(os.Stdin); nil != err {
			return fmt.Errorf("reading archive: %w", err)
		}
	} else if nil == a.fs {
		if b, err = os.ReadFile(a.Filename); nil != err {
			return fmt.Errorf("reading %s: %w", a.Filename, err)
		}
	} else {
		if b, err = fs.ReadFile(a.fs, a.Filename); nil != err {
			return fmt.Errorf("reading %s: %w", a.Filename, err)
		}
	}
	/* Decompress, if we're doing that. */
	if a.WithGzip {
		zr, err := gzip.NewReader(bytes.NewReader(b))
		if nil != err {
			return fmt.Errorf(
				"initializing gunzipper: %w",
				err,
			)
		}
		if b, err = io.ReadAll(zr); nil != err {
			return fmt.Errorf("gunzipping: %w", err)
		}
	}

	/* Parse into an archive. */
	ar := txtar.Parse(b)

	/* Print the comment, if we're verbose. */
	if a.Verbose && 0 == len(ar.Comment) {
		if _, err := fmt.Fprintf(w, "-No Comment-\n\n"); nil != err {
			return err
		}
	} else if a.Verbose {
		if _, err := fmt.Fprintf(w, "%s\n", ar.Comment); nil != err {
			return err
		}
	}

	/* Print and/or extract each allowed file plus maybe its size. */
	for _, f := range ar.Files {
		if err := a.extractFromArchive(
			w,
			f,
			where,
			doExtract,
		); nil != err {
			return fmt.Errorf("processing %s: %w", f.Name, err)
		}
	}

	return nil
}

// listOrExtractFromArchive lists or extracts f.  Listing output is written to
// w.
func (a Archiver) extractFromArchive(
	w io.Writer,
	f txtar.File,
	where string,
	doExtract bool,
) error {
	/* Work out what we'll call this file locally. */
	hn := a.ToHostPath(f.Name)

	/* Skip excluded files and files not on our list, if we have one. */
	if excl, err := a.isExcluded(hn); nil != err {
		return fmt.Errorf("checking if %s is excluded: %w", hn, err)
	} else if excl {
		return nil
	}

	/* And, if we have a file list, only those. */
	var found bool
	for _, g := range a.Paths {
		if ok, err := filepath.Match(g, hn); nil != err {
			return fmt.Errorf("invalid glob %s: %s", g, err)
		} else if ok {
			found = true
			break
		}
	}
	if 0 != len(a.Paths) && !found {
		return nil
	}

	/* If we're extracting, do it. */
	if doExtract {
		fn := filepath.Join(where, hn)
		/* Make sure parent directories exist. */
		dn := filepath.Dir(fn)
		if err := os.MkdirAll(dn, CreateDirPerms); nil != err {
			return fmt.Errorf("creating directory %s: %w", dn, err)
		}
		/* Write the file itself. */
		if err := os.WriteFile(
			fn,
			f.Data,
			CreateFilePerms,
		); nil != err {
			return fmt.Errorf("writing %s: %w", hn, err)
		}
	}

	/* Work out what to print, if anything. */
	var err error
	switch {
	case a.Verbose && doExtract, !a.Verbose && !doExtract: /* Filename. */
		_, err = fmt.Fprintf(w, "%s\n", hn)
	case a.Verbose && !doExtract: /* Filename and size. */
		_, err = fmt.Fprintf(w, "%d\t%s\n", len(f.Data), hn)
	case !a.Verbose && doExtract: /* Nothing. */
	default:
		panic("BUG: Unpossible combination of verbose and doExtract")
	}
	if nil != err {
		return fmt.Errorf("writing file info: %w", err)
	}

	return nil
}
