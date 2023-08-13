package main

/*
 * extract.go
 * Extract or list an archive
 * By J. Stuart McMurray
 * Created 20230516
 * Last Modified 20230813
 */

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
	"golang.org/x/tools/txtar"
)

const (
	// dirPerms are the permissions newly-created directories get.
	dirPerms = 0770
	// filePerms are the permissions newly-created files get.
	filePerms = 0660
)

// Extract walks through an archive and extracts the files in it.  If onlyList
// is true, files aren't actually extracted, the filenames are just printed.
func Extract(onlyList bool) error {
	/* Work out the files to list or extract, and keep track of tthe ones
	we've seen. */
	var rFiles map[string]bool
	if 0 != len(Files) {
		rFiles = make(map[string]bool)
		for _, f := range Files {
			rFiles[f] = false
		}
	}

	/* Get the archive into memory. */
	b, err := getArchive()
	if nil != err {
		return fmt.Errorf("loading archive: %w", err)
	}

	/* Parse it. */
	archive := txtar.Parse(b) /* The lack of error isn't great. */

	/* If we've got a comment, print it (with -v). */
	if 0 != len(archive.Comment) {
		VPrintf("Comment:\n%s", archive.Comment) /* Library adds \n. */
	}

	/* Extract or list the files. */
	for _, file := range archive.Files {
		/* If we're only worried about certain files, make sure this
		is one. */
		if nil != rFiles {
			if _, ok := rFiles[string(file.Name)]; !ok {
				continue
			}
			rFiles[string(file.Name)] = true
		}
		/* Skip files we don't want. */
		if nil != excludeRE && excludeRE.MatchString(file.Name) {
			continue
		}
		/* Handle this file. */
		if onlyList {
			listFile(file)
		} else {
			extractFile(file)
		}
	}

	/* If we didn't find all the files, note the missing ones. */
	var missing strings.Builder
	for _, f := range Files {
		if !rFiles[f] {
			fmt.Fprintf(&missing, "\t%s\n", f)
		}
	}
	if 0 != missing.Len() {
		Errorf(
			"The following files were not found:\n%s",
			missing.String(),
		)
	}

	return nil
}

// getArchive gets the archive as a []byte.  If it's a file, the archive is
// mapped into memory.  Stdin is slurped.
func getArchive() ([]byte, error) {
	/* Stdin is easy, if a bit more memory-intensive. */
	if StdioFileName == ArchiveFile {
		return io.ReadAll(os.Stdin)
	}
	/* Map the archive into memory. */
	f, err := os.Open(ArchiveFile)
	if nil != err {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()
	pos, err := f.Seek(0, io.SeekEnd)
	if nil != err {
		return nil, fmt.Errorf("determining size: %w", err)
	}
	if uint64(pos) > math.MaxInt {
		return nil, fmt.Errorf(
			"size %d too large (max %d)",
			uint64(pos),
			math.MaxInt,
		)
	}
	mb, err := unix.Mmap(
		int(f.Fd()),
		0,
		int(pos),
		unix.PROT_READ,
		unix.MAP_SHARED,
	)
	if nil != err {
		return nil, fmt.Errorf("mapping into memory: %s", err)
	}
	return mb, nil
}

// listFile prints info about the file.
func listFile(f txtar.File) {
	VPrintf("%d\t", len(f.Data))
	fmt.Printf("%s\n", f.Name)
}

// extractFile extracts f to disk.
func extractFile(tf txtar.File) {
	/* Work out the filename we'll actually use. */
	cpath := Clean(tf.Name)
	if cpath != tf.Name {
		log.Printf("Cleaned %q to %q", tf.Name, cpath)
	}
	/* Make sure we have the directory. */
	dir := filepath.Dir(cpath)
	if err := os.MkdirAll(dir, dirPerms); nil != err {
		Errorf("Error creating directory %s: %s", dir, err)
		return
	}
	/* Extract the file itself. */
	f, err := os.OpenFile(
		cpath,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		filePerms,
	)
	if nil != err {
		Errorf("Error opening %s: %s", cpath, err)
	}
	defer f.Close()
	if _, err := f.Write(tf.Data); nil != err {
		Errorf("Error writing to %s: %s", cpath, err)
	}
	VPrintf("%s\n", cpath)
}
