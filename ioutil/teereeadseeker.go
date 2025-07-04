// Package ioutil extends the stdlib ioutil
package ioutil

import (
	"fmt"
	"io"
)

// TeeReadSeeker returns a [ReadSeeker] that writes to w what it reads from r.
// All reads from r performed through it are matched with
// corresponding writes to w. There is no internal buffering -
// the write must complete before the read completes.
// Any error encountered while writing is reported as a read error.
// This is the same as io.TeeReader but also supports Seek.
func TeeReadSeeker(r io.ReadSeeker, w io.Writer) io.ReadSeeker {
	return &teeReadSeeker{r, w}
}

type teeReadSeeker struct {
	r io.ReadSeeker
	w io.Writer
}

func (t *teeReadSeeker) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		if n, err := t.w.Write(p[:n]); err != nil {
			return n, fmt.Errorf("w.Write failed %w", err)
		}
	}
	return
}

func (t *teeReadSeeker) Seek(offset int64, whence int) (int64, error) {
	seeked, err := t.r.Seek(offset, whence)

	if err != nil {
		return seeked, fmt.Errorf("r.Seek failed %w", err)
	}
	return seeked, nil
}
