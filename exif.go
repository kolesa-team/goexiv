package goexiv

// #cgo pkg-config: exiv2
// #include "helper.h"
// #include <stdlib.h>
import "C"

import (
	"runtime"
	"unsafe"
)

type ExifData struct {
	img  *Image // We point to img to keep it alive
	data *C.Exiv2ExifData
}

type ExifDatum struct {
	data  *ExifData
	datum *C.Exiv2ExifDatum
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
