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
