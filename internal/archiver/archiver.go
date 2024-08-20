// Package archiver - mqtxtar's underlying archiver
package archiver

/*
 * archiver.go
 * mqtxtar's underlying archiver
 * By J. Stuart McMurray
 * Created 20240812
 * Last Modified 20240819
 */

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// CreatePerms ar the permissios with which we create files.
	CreatePerms = 0600
)

// Archiver is what actually does the mqtxtar things.
type Archiver struct {
	Comment  string /* Archive comment. */
	Filename string /* Archive filename, or - for stdio. */
	WithGzip bool   /* (De)compress with gzip. */

	Paths       []string /* Paths to add/extract, i.e. flag.Args(). */
	UnsafePaths bool     /* Don't strip leading /'s. */

	Verbose bool /* Verbose messages. */

	ExcludeGlobs []string         /* Blacklist of globs. */
	ExcludeREs   []*regexp.Regexp /* Blacklist of Regexen. */

	fs fs.FS /* For testing. */
}

// New returns a new Archiver, ready for use.
func New(
	comment string,
	filename string,
	withGzip bool,
	paths []string,
	unsafePaths bool,
	verbose bool,
	excludeGlobs []string,
	excludeREs []*regexp.Regexp,
) Archiver {
	a := Archiver{
		Comment:      comment,
		Filename:     filename,
		WithGzip:     withGzip,
		Paths:        paths,
		UnsafePaths:  unsafePaths,
		Verbose:      verbose,
		ExcludeGlobs: excludeGlobs,
		ExcludeREs:   excludeREs,
	}
	if nil == a.Paths {
		a.Paths = make([]string, 0)
	}
	return a
}

//// ToJSON returns a as an indented JSON object.  The program panics on error.
////
//// This was used in creating the JSON files in the testdata directory.
//func (a Archiver) ToJSON() string {
//	b, err := json.MarshalIndent(a, "", "\t")
//	if nil != err {
//		panic(fmt.Sprintf("failed to JSON: %s", err))
//	}
//	return string(b)
//}

// isExcluded returns true if any of a's exclude globs or regexes matches fpath.
func (a Archiver) isExcluded(fpath string) (bool, error) {
	for _, g := range a.ExcludeGlobs {
		ok, err := path.Match(g, fpath)
		if nil != err {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	for _, re := range a.ExcludeREs {
		if re.MatchString(fpath) {
			return true, nil
		}
	}
	return false, nil
}

// AddPathsFromFile adds paths from the file fn.  Each line in the file should
// be one path.  Duplicates aren't added.
func (a *Archiver) AddPathsFromFile(fn string) error {
	/* Mapify, for not adding dupes. */
	m := make(map[string]struct{})
	for _, p := range a.Paths {
		m[p] = struct{}{}
	}

	/* Open the file and prep for line-by-line reading. */
	f, err := os.Open(fn)
	if nil != err {
		return fmt.Errorf("opening: %w", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	/* Add each line from the file. */
	for scanner.Scan() {
		/* Get the next line, ignoring blanks. */
		l := strings.TrimSpace(scanner.Text())
		if "" == l {
			continue
		}
		/* Don't add dupes. */
		if _, ok := m[l]; ok {
			continue
		}
		/* Looks like we've a new one. */
		m[l] = struct{}{}
		a.Paths = append(a.Paths, l)
	}
	if err := scanner.Err(); nil != err {
		return fmt.Errorf("reading lines: %s", err)
	}

	return nil
}

// ToHostPath turns the txtar path p into an OS path, possibly safening it.
func (a *Archiver) ToHostPath(p string) string {
	return filepath.FromSlash(a.maybeSafenPath(p))
}

// FromHostPath turns the host path p into a txtar path, possibly safening it.
func (a *Archiver) FromHostPath(p string) string {
	return a.maybeSafenPath(filepath.ToSlash(p))
}

// maybeSafenPath safens the /-path p if a.UnsafePaths isn't set.
func (a *Archiver) maybeSafenPath(p string) string {
	if a.UnsafePaths {
		return p
	}
	/* Hack to remove Leading ../'s. */
	if strings.HasPrefix(p, "..") {
		p = "/" + p
	}
	/* Remove ALL the ..'s (and so on). */
	p = path.Clean(p)
	/* Don't be an absolute path. */
	p = strings.TrimLeft(p, "/")
	/* Normally clean would do this, but clean might also give us a /. */
	if "" == p {
		p = "."
	}

	return p
}
