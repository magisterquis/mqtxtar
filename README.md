MagisterQuis' Text Archiver
===========================
[Tar](https://man.openbsd.org/tar)-like utility for operating on
[txtar](https://pkg.go.dev/golang.org/x/tools/txtar) archives.

Features
--------
- Create, Extract, and List the contents of txtar archives
- Works with files or Standard Input/Output
- Tar-like flags, but with just enough differece (`-cv` -> `-c -v`) to prevent
  mistakes due to overreliance on muscle memory
- Gzip compression and de-compression
- Exclude files based on globs or regex

Quickstart
----------
1. Have [Go installed](https://go.dev/doc/install).
2. Install the program itself.
   ```sh
   go install github.com/magisterquis/mqtxtar@latest
   ```
3. Create an archive.
   ```sh
   $ mqtxtar -c -f code.txtar *.go internal/*/*.go
   ```
4. List the contents of an archive.
   ```sh
   $ mqtxtar -t -f code.txtar
   mqtxtar.go
   internal/archiver/archiver.go
   internal/archiver/archiver_test.go
   internal/archiver/create.go
   internal/archiver/create_test.go
   internal/archiver/listextract.go
   internal/archiver/listextract_test.go
   ```
5. Extract an archive.
   ```sh
   $ mqtxtar -x -f code.txtar
   ```  

Usage
-----
```
Usage: mqtxtar -c|-t|-x [options] [paths...]

Tar-like txtar utility.  Creates, extracts, and lists the contents of archives
in txtar format.  For more details on txtar archives, please see

https://pkg.go.dev/golang.org/x/tools/txtar

Paths to be added or extracted can be given as arguments or in a file specified
with -I or both.  All paths within an archive use forward (Unix) slashes.

Options:
  -C directory
    	Set the working directory before doing anything else
  -I file
    	Optional file containing names of paths to add or extract, one per line
  -P	Do not strip leading slashes from pathnames
  -c	Create an archive
  -comment comment
    	Set archive comment, with -c and
  -exclude glob
    	Do not add or extract files matching theglob (may be repeated)
  -exclude-re regex
    	Do not add or extract files matching the regex (may be repeated)
  -f file
    	Optional archive file to use instead of standard input/output
  -t	List archive contents
  -v	Enable verbose output
  -x	Extract archive contents
  -z	(De)compress archive using gzip
```
