package token

import (
	"fmt"
)

// File represents a source file.
type File struct {
	name  string // file name as provided to AddFile
	base  int    // Pos value range for this file is [base...base+size]
	size  int    // file size as provided to AddFile
	lines []int  // lines contains the offset of the first character for each line (the first entry is always 0)
}

// Name returns the file name of file f as registered with AddFile.
func (f *File) Name() string {
	return f.name
}

// Base returns the base offset of file f as registered with AddFile.
func (f *File) Base() int {
	return f.base
}

// Size returns the size of file f as registered with AddFile.
func (f *File) Size() int {
	return f.size
}

// LineCount returns the number of lines in file f.
func (f *File) LineCount() int {
	return len(f.lines)
}

// Lines returns the effective line offset table of the form described by [File.SetLines].
// Callers must not mutate the result.
func (f *File) Lines() []int {
	return f.lines
}

// AddLine adds the line offset for a new line.
// The line offset must be larger than the offset for the previous line
// and smaller than the file size; otherwise the line offset is ignored.
func (f *File) AddLine(offset int) {
	i := len(f.lines)
	if (i == 0 || f.lines[i-1] < offset) && offset < f.size {
		f.lines = append(f.lines, offset)
	}
}

// LineStart returns the position of the first character in the line.
func (f *File) LineStart(line int) Pos {
	switch {
	case line < 1:
		panic(fmt.Sprintf("invalid line number %d (should be >= 1)", line))

	case line > len(f.lines):
		panic(fmt.Sprintf("invalid line number %d (should be < %d)", line, len(f.lines)))

	default:
		return Pos(f.base + f.lines[line-1])
	}
}

// FileSetPos returns the position in the file set.
func (f *File) Pos(offset int) Pos {
	return Pos(f.base + f.fixOffset(offset))
}

// Offset returns the offset for the given file position p.
func (f *File) Offset(p Pos) int {
	return f.fixOffset(int(p) - f.base)
}

// Line returns the line number for the given file position p.
func (f *File) Line(p Pos) int {
	return f.Position(p).Line
}

// Position returns the position value for the given file position p.
// If p is out of bounds, it is adjusted to match the File.Offset behavior.
func (f *File) Position(p Pos) Position {
	if p == NoPos {
		return Position{}
	}
	return f.position(p)
}

func (f *File) position(p Pos) Position {
	offset := f.fixOffset(int(p) - f.base)
	var pos Position
	pos.Offset = offset
	pos.Filename, pos.Line, pos.Column = f.unpack(offset)
	return pos
}

// fixOffset fixes an out-of-bounds offset such that 0 <= offset <= f.size.
func (f *File) fixOffset(offset int) int {
	switch {
	case offset < 0:
		return 0
	case offset > f.size:
		return f.size
	default:
		return offset
	}
}

// unpack returns the filename, line, column number for a file offset.
func (f *File) unpack(offset int) (filename string, line, column int) {
	filename = f.name
	if i := searchInts(f.lines, offset); i >= 0 {
		line, column = i+1, offset-f.lines[i]+1
	}
	return filename, line, column
}

func searchInts(a []int, x int) int {
	i, j := 0, len(a)
	for i < j {
		h := i + (j-i)/2 // avoid overflow when computing h
		if a[h] <= x {
			i = h + 1
		} else {
			j = h
		}
	}
	return i - 1
}
