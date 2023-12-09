package slices

import "io"

// ReadAllForByte Rewrite io ReadAll removes [] bytes and automatically expands due to insufficient capacity
func ReadAllForByte(r io.Reader, b []byte) error {
	b = b[:0]
	for {
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
	}
}
