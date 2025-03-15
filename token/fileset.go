package token

import (
	"cmp"
	"fmt"
	"slices"
)

// FileSet represents a set of source files.
type FileSet struct {
	base  int     // base offset for the next file
	files []*File // list of files in the order added to the set
	last  *File   // cache of last file looked up
}

// NewFileSet creates a new file set.
func NewFileSet() *FileSet {
	return &FileSet{
		base: 1, // 0 == NoPos
	}
}

// Base returns the minimum base offset that must be provided to
// [FileSet.AddFile] when adding the next file.
func (s *FileSet) Base() int {
	return s.base
}

// AddFile adds a new file in the file set.
func (s *FileSet) AddFile(filename string, base, size int) *File {
	if base < 0 {
		base = s.base
	}

	switch {
	case base < s.base:
		panic(fmt.Sprintf("invalid base %d (should be >= %d)", base, s.base))
	case size < 0:
		panic(fmt.Sprintf("invalid size %d (should be >= 0)", size))
	}

	f := &File{
		name:  filename,
		base:  base,
		size:  size,
		lines: []int{0},
	}

	// base >= s.base && size >= 0
	base += size + 1 // +1 because EOF also has a position
	if base < 0 {
		panic("token.Pos offset overflow (> 2G of source code in file set)")
	}

	// add the file to the file set
	s.base = base
	s.files = append(s.files, f)
	s.last = f
	return f
}

// File returns the file that contains the position p.
// If no such file is found the result is nil.
func (s *FileSet) File(p Pos) *File {
	if p == NoPos {
		return nil
	}
	return s.file(p)
}

// Position converts a [Pos] p in the fileset into a Position value.
func (s *FileSet) Position(p Pos) Position {
	if p == NoPos {
		return Position{}
	}
	if f := s.file(p); f != nil {
		return f.position(p)
	}
	return Position{}
}

func (s *FileSet) file(p Pos) *File {
	// common case: p is in last file.
	if f := s.last; f != nil && f.base <= int(p) && int(p) <= f.base+f.size {
		return f
	}

	// p is not in last file - search all files
	if i := searchFiles(s.files, int(p)); i >= 0 {
		f := s.files[i]
		// f.base <= int(p) by definition of searchFiles
		if int(p) <= f.base+f.size {
			s.last = f
			return f
		}
	}
	return nil
}

func searchFiles(a []*File, x int) int {
	i, found := slices.BinarySearchFunc(a, x, func(a *File, x int) int {
		return cmp.Compare(a.base, x)
	})
	if !found {
		// We want the File containing x, but if we didn't
		// find x then i is the next one.
		i--
	}
	return i
}
