package goexiv

// #cgo pkg-config: exiv2
// #include "helper.h"
// #include <stdlib.h>
import "C"

import (
	"runtime"
	"unsafe"
)

// XmpData contains all Xmp Data of an image.
type XmpData struct {
	img  *Image // We point to img to keep it alive
	data *C.Exiv2XmpData
}

// XmpDatum stores the info of one xmp datum.
type XmpDatum struct {
	data  *XmpData
	datum *C.Exiv2XmpDatum
}

func makeXmpData(img *Image, cdata *C.Exiv2XmpData) *XmpData {
	data := &XmpData{
		img,
		cdata,
	}

	runtime.SetFinalizer(data, func(x *XmpData) {
		C.exiv2_xmp_data_free(x.data)
	})

	return data
}

func makeXmpDatum(data *XmpData, cdatum *C.Exiv2XmpDatum) *XmpDatum {
	if cdatum == nil {
		return nil
	}

	datum := &XmpDatum{
		data,
		cdatum,
	}

	runtime.SetFinalizer(datum, func(x *XmpDatum) {
		C.exiv2_xmp_datum_free(x.datum)
	})

	return datum
}

// GetXmpData returns the XmpData of an Image.
func (i *Image) GetXmpData() *XmpData {
	return makeXmpData(i, C.exiv2_image_get_xmp_data(i.img))
}

// FindKey tries to find the specified key and returns its data.
// It returns an error if the key is invalid. If the key is not found, a
// nil pointer will be returned
func (d *XmpData) FindKey(key string) (*XmpDatum, error) {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))

	var cerr *C.Exiv2Error

	cdatum := C.exiv2_xmp_data_find_key(d.data, ckey, &cerr)

	if cerr != nil {
		err := makeError(cerr)
		C.exiv2_error_free(cerr)
		return nil, err
	}

	runtime.KeepAlive(d)
	return makeXmpDatum(d, cdatum), nil
}

func (d *XmpDatum) String() string {
	cstr := C.exiv2_xmp_datum_to_string(d.datum)
	defer C.free(unsafe.Pointer(cstr))

	return C.GoString(cstr)
}

func (i *Image) XmpStripKey(key string) error {
	return i.StripKey(XMP, key)
}
