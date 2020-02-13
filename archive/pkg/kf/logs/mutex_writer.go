package logs

import (
	"fmt"
	"io"
	"sync"
)

// MutexWriter contains writer with a mutex instance
type MutexWriter struct {
	Writer io.Writer
	sync.Mutex
}

// Write implements io.Writer
func (mw *MutexWriter) Write(data []byte) (int, error) {
	mw.Lock()
	defer mw.Unlock()
	return mw.Writer.Write(data)
}

// WriteStringf writes a string to the underlying Writer. It has the same
// symantics as fmt.Sprintf.
func (mw *MutexWriter) WriteStringf(s string, args ...interface{}) error {
	if _, err := io.WriteString(mw, fmt.Sprintf(s, args...)); err != nil {
		return err
	}

	return nil
}

// CopyFrom copies from s to to Writer
func (mw *MutexWriter) CopyFrom(s io.ReadCloser) error {
	mw.Lock()
	defer mw.Unlock()
	if _, err := io.Copy(mw.Writer, s); err != nil {
		if err == io.EOF {
			return nil
		}

		return err
	}

	return nil
}
