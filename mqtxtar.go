// Program mqtxtar: Tar-like txtar utility
package main

/*
 * mqtxtar.go
 * mqtxtar: Tar-like txtar utility
 * By J. Stuart McMurray
 * Created 20230516
 * Last Modified 20230813
 */

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
)

var (
	/* V*f wil be a no-op if -v isn't given. */
	VLogf   = log.Printf
	VPrintf = func(f string, a ...any) { fmt.Fprintf(os.Stderr, f, a...) }

	/* ArchiveFile is the name of the archive file on which to operate. */
	ArchiveFile = StdioFileName

	/* Unsafe controls whether we try to sanitize filenames. */
	Unsafe bool

	/* StopOnError controls whether or not we stop after the first error
	encountered. */
	StopOnError bool

	/* EncounteredError is true if we encountered an error during
	processing. */
	EncounteredError atomic.Bool

	/* Files is the list of paths on which to operate.  It's the union
	of flag.Args() and anything in the file given with -I. */
	Files = make([]string, 0)

	/* excludeRE indicates which files to not add or extract. */
	excludeRE *regexp.Regexp
)

// StdioFileName indicates we use stdin/out instead of a regular file.
const StdioFileName = "-"

func main() {
	/* Command-line flags. */
	var (
		verbOn = flag.Bool(
			"v",
			false,
			"Enable verbose output",
		)
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
		comment = flag.String(
			"comment",
			"",
			"Set archive `comment`, with -c",
		)
		relDir = flag.String(
			"C",
			"",
			"Relative `directory` for files for -c or -x",
		)
		listFile = flag.String(
			"I",
			"",
			"Optional `file` containing names of files to "+
				"add or extract",
		)
		exclude = flag.String(
			"exclude",
			"",
			"Do not add or extract files matching the `regex`",
		)
	)
	flag.BoolVar(
		&Unsafe,
		"P",
		Unsafe,
		"Do not sanitize filenames before extraction",
	)
	flag.StringVar(
		&ArchiveFile,
		"f",
		ArchiveFile,
		"Name of archive `file` or "+StdioFileName+
			" for stdin/out",
	)
	flag.BoolVar(
		&StopOnError,
		"e",
		StopOnError,
		"Stop after the first error",
	)
	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			`Usage: %s -c|-x|-t [options] [files...]

Tar-like txtar utility.  Creates, extracts, and lists the contents of archives
in txtar format.  For more details on txtar archives, please see

https://pkg.go.dev/golang.org/x/tools/txtar

Files to be added or extracted can be given as arguments or in a file specified
with -I or both.

Options:
`,
			os.Args[0],
		)
		flag.PrintDefaults()
	}
	flag.Parse()

	/* Work out verbose logging. */
	if !*verbOn {
		VLogf = func(string, ...any) {}
		VPrintf = func(string, ...any) {}
	}
	log.SetFlags(0)
	log.SetPrefix(filepath.Base(os.Args[0]) + ": ")

	/* If we're regexing, compile the regexen. */
	if "" != *exclude {
		var err error
		if excludeRE, err = regexp.Compile(*exclude); nil != err {
			log.Fatalf(
				"Error compiling exclude regex %q: %s",
				*exclude,
				err,
			)
		}
	}

	/* Make sure we have exactly one action. */
	var nDo int
	for _, p := range []*bool{doCreate, doExtract, doList} {
		if *p {
			nDo++
		}
	}
	if 1 != nDo {
		log.Fatalf("Need exactly one of -c, -x, or -t")
	}

	/* Work out the list of files on which to operate.  We do this now
	so as to not have to worry about chdiring. */
	mustGetFiles(*listFile)

	/* If we've got a relative path, get the absolute path to the archive
	file and chdir to the relative path, for simplicity. */
	if "" != *relDir {
		/* Make sure we have an absolute path for the archive file if
		we're not using stdio. */
		if StdioFileName != ArchiveFile {
			var err error
			if ArchiveFile, err = filepath.Abs(
				ArchiveFile,
			); nil != err {
				log.Fatalf(
					"Error ensuring archive file has "+
						"absolute path: %s",
					err,
				)
			}
		}
		/* Chdir to the -C directory, so we don't have to worry
		about filepath things. */
		if err := os.Chdir(*relDir); nil != err {
			log.Fatalf(
				"Unable to chdir to %s: %s",
				*relDir,
				err,
			)
		}
	}

	/* Do it! */
	switch {
	case *doCreate:
		if err := Create(*comment); nil != err {
			log.Fatalf("Error creating archive: %s", err)
		}
	case *doExtract:
		if err := Extract(false); nil != err {
			log.Fatalf("Error extracting archive: %s", err)
		}
	case *doList:
		if err := Extract(true); nil != err {
			log.Fatalf("Error listing archive: %s", err)
		}
	}

	/* If we tried and nothing worked, tell the shell. */
	if EncounteredError.Load() {
		os.Exit(1)
	}
}

// Errorf prints a message and sets EncounteredError.  If StopOnError is true
// Errorf terminates the program.
func Errorf(format string, args ...any) {
	/* Tell the user what went wrong. */
	log.Printf(format, args...)
	/* If all errors are fatal, this is fatal. */
	if StopOnError {
		os.Exit(1)
	}
	/* Not we've an error so we can exit non-zero. */
	EncounteredError.Store(true)
}

// mustGetFiles gets the list of paths on which to operate.  It terminates the
// program on error.  The paths are stored in Files.
func mustGetFiles(listFile string) {
	ps := make(map[string]struct{}) /* For deduping. */
	/* add adds s to paths if it's not in ps. */
	add := func(s string) {
		if _, ok := ps[s]; ok {
			return
		}
		ps[s] = struct{}{}
		Files = append(Files, s)
	}

	/* Add files from the listfile, if we have one. */
	if "" != listFile {
		lf, err := os.Open(listFile)
		if nil != err {
			log.Fatalf("Error opening file list: %s", err)
		}
		defer lf.Close()
		scanner := bufio.NewScanner(lf)
		for scanner.Scan() {
			l := strings.TrimSpace(scanner.Text())
			if "" == l {
				continue
			}
			add(l)
		}
		if err := scanner.Err(); nil != err {
			log.Fatalf("Error reading file list: %s", err)
		}
	}

	/* Add files from the commandline. */
	for _, p := range flag.Args() {
		add(p)
	}
}
