package logs

import (
	"io"
	"sync"
)

// MutexWriter contains writer with a mutex instance
type MutexWriter struct {
	Writer io.Writer
	sync.Mutex
}

// Write writes string to Writer
func (mw *MutexWriter) Write(s string) error {
	mw.Lock()
	if _, err := io.WriteString(mw.Writer, s); err != nil {
		return err
	}
	mw.Unlock()
	return nil
}

// CopyFrom copies from s to to Writer
func (mw *MutexWriter) CopyFrom(s io.ReadCloser) error {
	mw.Lock()
	if _, err := io.Copy(mw.Writer, s); err != nil {
		if err == io.EOF {
			return nil
		}

		return err
	}
	mw.Unlock()
	return nil
}
