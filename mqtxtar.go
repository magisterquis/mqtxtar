// Program mqtxtar: Tar-like txtar utility
package main

/*
 * mqtxtar.go
 * mqtxtar: Tar-like txtar utility
 * By J. Stuart McMurray
 * Created 20230516
 * Last Modified 20240819
 */

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"

	"github.com/magisterquis/mqtxtar/internal/archiver"
)

func main() {
	var (
		excludeGlobs []string
		excludeREs   []*regexp.Regexp
	)
	/* Actions, of which only one at a time may be used. */
	var (
		doCreate = flag.Bool(
			"c",
			false,
			"Create an archive",
		)
		doExtract = flag.Bool(
			"x",
			false,
			"Extract archive contents",
		)
		doList = flag.Bool(
			"t",
			false,
			"List archive contents",
		)
	)
	/* Other flags. */
	var (
		wDir = flag.String(
			"C",
			"",
			"Set the working `directory` before doing "+
				"anything else",
		)
		archiveName = flag.String(
			"f",
			"",
			"Optional archive `file` to use instead of standard "+
				"input/output",
		)
		listFile = flag.String(
			"I",
			"",
			"Optional `file` containing names of paths to "+
				"add or extract, one per line",
		)
		unsafePaths = flag.Bool(
			"P",
			false,
			"Do not strip leading slashes from pathnames",
		)
		verbose = flag.Bool(
			"v",
			false,
			"Enable verbose output",
		)
		comment = flag.String(
			"comment",
			"",
			"Set archive `comment`, with -c and",
		)
		withGzip = flag.Bool(
			"z",
			false,
			"(De)compress archive using gzip",
		)
	)
	flag.Func(
		"exclude",
		"Do not add or extract files matching the`glob` "+
			"(may be repeated)",
		func(s string) error {
			excludeGlobs = append(excludeGlobs, s)
			return nil
		},
	)
	flag.Func(
		"exclude-re",
		"Do not add or extract files matching the `regex` "+
			"(may be repeated)",
		func(s string) error {
			re, err := regexp.Compile(s)
			if nil != err {
				return err
			}
			excludeREs = append(excludeREs, re)
			return nil
		},
	)
	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			`Usage: %s -c|-t|-x [options] [paths...]

Tar-like txtar utility.  Creates, extracts, and lists the contents of archives
in txtar format.  For more details on txtar archives, please see

https://pkg.go.dev/golang.org/x/tools/txtar

Paths to be added or extracted can be given as arguments or in a file specified
with -I or both.  All paths within an archive use forward (Unix) slashes.

Options:
`,
			os.Args[0],
		)
		flag.PrintDefaults()
	}
	flag.Parse()

	/* If we have a directory to be in, do that first like we said we
	would. */
	if "" != *wDir {
		if err := os.Chdir(*wDir); nil != err {
			log.Fatalf("Cannot chdir to %s: %s", *wDir, err)
		}
	}

	/* Roll the archiver with options. */
	a := archiver.New(
		*comment,
		*archiveName,
		*withGzip,
		flag.Args(),
		*unsafePaths,
		*verbose,
		excludeGlobs,
		excludeREs,
	)
	if "" != *listFile {
		if err := a.AddPathsFromFile(*listFile); nil != err {
			log.Fatalf(
				"Error adding paths from %s: %s",
				*listFile,
				err,
			)
		}
	}

	/* Make sure we only have one action. */
	if 1 != len(slices.DeleteFunc(
		[]bool{*doCreate, *doExtract, *doList},
		func(b bool) bool { return !b },
	)) {
		log.Fatalf("Need exactly one of -c, -t, or -x")
	}

	/* Figure out what to do. */
	var err error
	switch {
	case *doCreate:
		err = a.Create()
	case *doExtract:
		err = a.ListOrExtract(os.Stdout, "", true)
	case *doList:
		err = a.ListOrExtract(os.Stdout, "", false)
	default:
		panic("no action given")
	}
	if nil != err {
		log.Fatalf("Fatal error: %s", err)
	}
}
