package ioutil

import "io"

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
			return n, err
		}
	}
	return
}

func (t *teeReadSeeker) Seek(offset int64, whence int) (int64, error) {
	return t.r.Seek(offset, whence)
}
