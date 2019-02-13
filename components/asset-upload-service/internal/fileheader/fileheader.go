package fileheader

import "mime/multipart"

//go:generate mockery -name=FileHeader -output=automock -outpkg=automock -case=underscore
type FileHeader interface {
	Filename() string
	Size() int64
	Open() (multipart.File, error)
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

func (h *multipartFileHeader) Open() (multipart.File, error) {
	return h.FileHeader.Open()
}

func FromMultipart(header *multipart.FileHeader) FileHeader {
	return &multipartFileHeader{header}
}
