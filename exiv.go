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

type ExifData struct {
	img  *Image // We point to img to keep it alive
	data *C.Exiv2ExifData
}

type ExifDatum struct {
	data  *ExifData
	datum *C.Exiv2ExifDatum
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

func makeExifData(img *Image, cdata *C.Exiv2ExifData) *ExifData {
	data := &ExifData{
		img,
		cdata,
	}

	runtime.SetFinalizer(data, func(x *ExifData) {
		C.exiv2_exif_data_free(x.data)
	})

	return data
}

func makeExifDatum(data *ExifData, cdatum *C.Exiv2ExifDatum) *ExifDatum {
	if cdatum == nil {
		return nil
	}

	datum := &ExifDatum{
		data,
		cdatum,
	}

	runtime.SetFinalizer(datum, func(x *ExifDatum) {
		C.exiv2_exif_datum_free(x.datum)
	})

	return datum
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

func (i *Image) GetExifData() *ExifData {
	return makeExifData(i, C.exiv2_image_get_exif_data(i.img))
}

func (d *ExifData) FindKey(key string) (*ExifDatum, error) {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))

	var cerr *C.Exiv2Error

	cdatum := C.exiv2_exif_data_find_key(d.data, ckey, &cerr)

	if cerr != nil {
		err := makeError(cerr)
		C.exiv2_error_free(cerr)
		return nil, err
	}

	return makeExifDatum(d, cdatum), nil
}

func (d *ExifDatum) String() string {
	cstr := C.exiv2_exif_datum_to_string(d.datum)
	defer C.free(unsafe.Pointer(cstr))

	return C.GoString(cstr)
}
