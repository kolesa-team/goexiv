package goexiv

// #cgo pkg-config: exiv2
// #include "helper.h"
// #include <stdlib.h>
import "C"

import (
	"runtime"
	"unsafe"
)

type Error struct {
	code int
	what string
}

type Image struct {
	img *C.Exiv2Image
}

func (e *Error) Error() string {
	return e.what
}

func (e *Error) Code() int {
	return e.code
}

func makeError(cerr *C.Exiv2Error) *Error {
	return &Error{
		int(C.exiv2_error_code(cerr)),
		C.GoString(C.exiv2_error_what(cerr)),
	}
}

func makeImage(cimg *C.Exiv2Image) *Image {
	img := &Image{
		cimg,
	}

	runtime.SetFinalizer(img, func(x *Image) {
		C.exiv2_image_free(x.img)
	})

	return img
}

// Open opens an image file from the filesystem and returns a pointer to
// the corresponding Image object, but does not read the Metadata.
// Start the parsing with a call to ReadMetadata()
func Open(path string) (*Image, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	var cerr *C.Exiv2Error

	cimg := C.exiv2_image_factory_open(cpath, &cerr)

	if cerr != nil {
		err := makeError(cerr)
		C.exiv2_error_free(cerr)
		return nil, err
	}

	return makeImage(cimg), nil
}

// OpenBytes opens a byte slice with image data and returns a pointer to
// the corresponding Image object, but does not read the Metadata.
// Start the parsing with a call to ReadMetadata()
func OpenBytes(b []byte) (*Image, error) {
	var cerr *C.Exiv2Error
	cimg := C.exiv2_image_factory_open_bytes((*C.uchar)(unsafe.Pointer(&b[0])), C.long(len(b)), &cerr)

	if cerr != nil {
		err := makeError(cerr)
		C.exiv2_error_free(cerr)
		return nil, err
	}

	return makeImage(cimg), nil
}

// ReadMetadata reads the metadata of an Image
func (i *Image) ReadMetadata() error {
	var cerr *C.Exiv2Error

	C.exiv2_image_read_metadata(i.img, &cerr)

	if cerr != nil {
		err := makeError(cerr)
		C.exiv2_error_free(cerr)
		return err
	}

	return nil
}

// PixelWidth returns the width of the image in pixels
func (i *Image) PixelWidth() int64 {
	return int64(C.exiv2_image_get_pixel_width(i.img))
}

// PixelHeight returns the height of the image in pixels
func (i *Image) PixelHeight() int64 {
	return int64(C.exiv2_image_get_pixel_height(i.img))
}
