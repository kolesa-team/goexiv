package goexiv

// #cgo pkg-config: exiv2
// #include "helper.h"
// #include <stdlib.h>
import "C"

import (
	"runtime"
	"unsafe"
)


type XmpData struct {
	img  *Image // We point to img to keep it alive
	data *C.Exiv2XmpData
}

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

func (i *Image) GetXmpData() *XmpData {
	return makeXmpData(i, C.exiv2_image_get_xmp_data(i.img))
}

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

	return makeXmpDatum(d, cdatum), nil
}

func (d *XmpDatum) String() string {
	cstr := C.exiv2_xmp_datum_to_string(d.datum)
	defer C.free(unsafe.Pointer(cstr))

	return C.GoString(cstr)
}