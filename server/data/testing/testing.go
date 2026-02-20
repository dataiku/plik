package file

import (
	"bytes"
	"errors"
	"io"
	"sync"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/data"
)

// Ensure Testing Data Backend implements data.Backend interface
var _ data.Backend = (*Backend)(nil)

type BufferReadSeekCloser struct {
	buffer *bytes.Buffer
	off int64 // Forced to duplicate it since bytes.Buffer hides it
}

func NewBufferReadSeekCloser(buffer *bytes.Buffer) (io.ReadSeekCloser) {
	return &BufferReadSeekCloser { buffer: buffer, off: 0}
}

func (b *BufferReadSeekCloser) Read(data []byte) (n int, err error) {
	return b.buffer.Read(data)
}

func (b *BufferReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
    // Convert SeekStart and SeekEnd to SeekCurent equivalent
	switch whence {
    case io.SeekStart:
        offset = offset - b.off
	case io.SeekEnd:
		len := int64(b.buffer.Len()) + b.off
		offset = len - offset - b.off
	default:
	}

	// Seeking back is not supported by bytes.Buffer
	if offset < 0 {
		return b.off, io.ErrUnexpectedEOF
	}

    b.off += int64(len(b.buffer.Next(int(offset))))
	return b.off, nil
}


func (*BufferReadSeekCloser) Close() error {
	return nil
}


// Backend object
type Backend struct {
	files map[string][]byte
	err   error
	mu    sync.Mutex
}

// NewBackend instantiate a new Testing Data Backend
// from configuration passed as argument
func NewBackend() (b *Backend) {
	b = new(Backend)
	b.files = make(map[string][]byte)
	return
}

// GetFiles return the content of the backend for testing purposes
func (b *Backend) GetFiles() (files map[string][]byte) {
	return b.files
}

// GetFile implementation for testing data backend will search
// on filesystem the asked file and return its reading filehandle
func (b *Backend) GetFile(file *common.File) (reader io.ReadSeekCloser, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

	if content, ok := b.files[file.ID]; ok {
        return NewBufferReadSeekCloser(bytes.NewBuffer(content)), nil
	}

	return nil, errors.New("file not found")
}

// AddFile implementation for testing data backend will creates a new file for the given upload
// and save it on filesystem with the given file reader
func (b *Backend) AddFile(file *common.File, fileReader io.Reader) (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return b.err
	}

	if _, ok := b.files[file.ID]; ok {
		return errors.New("file exists")
	}

	content, err := io.ReadAll(fileReader)
	if err != nil {
		return err
	}

	b.files[file.ID] = content

	return nil
}

// RemoveFile implementation for testing data backend will delete the given
// file from filesystem
func (b *Backend) RemoveFile(file *common.File) (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return b.err
	}

	delete(b.files, file.ID)

	return nil
}

// SetError set the error that this backend will return on any subsequent method call
func (b *Backend) SetError(err error) {
	b.err = err
}
