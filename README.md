MagisterQuis' Text Archiver
===========================
[Tar](https://man.openbsd.org/tar)-like utility for operating on
[txtar](https://pkg.go.dev/golang.org/x/tools/txtar) archives.

Features
--------
- Create, Extract, and List the contents of txtar archives
- Works with files or Standard Input/Output
- Files to be extracted or listed are mmap'd, saving potential _kilobytes_ of
  memory
- Tar-like flags, but with just enough differece (`-cv` -> `-c -v`) to prevent
  mistakes due to overreliance on muscle memory

Quickstart
----------
```sh
go install github.com/magisterquis/mqtxtar@latest
mqtxtar -h

# Create an archive
mqtxtar -c -f src.txtar -comment "WIP Sources" *.go go.mod go.sum

# List an archive's contents
mqtxtar -t -f src.txtar

# Extract an archive
mqtxtar -x -f src.txtar
```

Usage
-----
```
Usage: ./mqtxtar -c|-x|-t [options] [files...]

Tar-like txtar utility.  Creates, extracts, and lists the contents of archives
in txtar format.  For more details on txtar archives, please see

https://pkg.go.dev/golang.org/x/tools/txtar

Files to be added or extracted can be given as arguments or in a file specified
with -I or both.

Options:
  -C directory
    	Relative directory for files for -c or -x
  -I file
    	Optional file containing names of files to add or extract
  -P	Do not sanitize filenames before extraction
  -c	Create an archive
  -comment comment
    	Set archive comment, with -c
  -e	Stop after the first error
  -exclude regex
    	Do not add or extract files matching the regex (may be repeated)
  -f file
    	Name of archive file or - for stdin/out (default "-")
  -t	List archive contents
  -v	Enable verbose output
  -x	Extract archive contents
```
