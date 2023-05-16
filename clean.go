package main

/*
 * clean.go
 * Sanitize paths
 * By J. Stuart McMurray
 * Created 20230516
 * Last Modified 20230516
 */

import (
	"path/filepath"
	"strings"
)

/* upPrefix is the path prefix which moves us up one level.  We remove it. */
const upPrefix = "../"

// Clean cleans a path.  It calls filepath.Clean, and ensures paths don't start
// with ../.  If Unsafe (-P) is set, Clean is a no-op.
func Clean(p string) string {
	/* Don't do this if we're living dangerously. */
	if Unsafe {
		return p
	}

	/* Remove leading ../ */
	for strings.HasPrefix(p, upPrefix) {
		p = strings.TrimPrefix(p, upPrefix)
	}

	/* Clean stray ..'s and other silliness. */
	p = filepath.Clean(p)

	return p
}
