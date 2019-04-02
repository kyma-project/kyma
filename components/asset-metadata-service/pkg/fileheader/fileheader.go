package fileheader

import (
	"io"
	"mime/multipart"
)

//go:generate mockery -name=FileHeader -output=automock -outpkg=automock -case=underscore
type FileHeader interface {
	Filename() string
	Size() int64
	Open() (File, error)
}

//go:generate mockery -name=File -output=automock -outpkg=automock -case=underscore
type File interface {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
}

type multipartFileHeader struct {
	*multipart.FileHeader
}

func (h *multipartFileHeader) Filename() string {
	return h.FileHeader.Filename
}

func (h *multipartFileHeader) Size() int64 {
	return h.FileHeader.Size
}

func (h *multipartFileHeader) Open() (File, error) {
	return h.FileHeader.Open()
}

func FromMultipart(header *multipart.FileHeader) FileHeader {
	return &multipartFileHeader{header}
}
