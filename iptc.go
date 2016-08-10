package goexiv

// #cgo pkg-config: exiv2
// #include "helper.h"
// #include <stdlib.h>
import "C"

import (
	"runtime"
	"unsafe"
)

type IptcData struct {
	img  *Image // We point to img to keep it alive
	data *C.Exiv2IptcData
}

type IptcDatum struct {
	data  *IptcData
	datum *C.Exiv2IptcDatum
}

func makeIptcData(img *Image, cdata *C.Exiv2IptcData) *IptcData {
	data := &IptcData{
		img,
		cdata,
	}

	runtime.SetFinalizer(data, func(x *IptcData) {
		C.exiv2_iptc_data_free(x.data)
	})

	return data
}

func makeIptcDatum(data *IptcData, cdatum *C.Exiv2IptcDatum) *IptcDatum {
	if cdatum == nil {
		return nil
	}

	datum := &IptcDatum{
		data,
		cdatum,
	}

	runtime.SetFinalizer(datum, func(x *IptcDatum) {
		C.exiv2_iptc_datum_free(x.datum)
	})

	return datum
}

func (i *Image) GetIptcData() *IptcData {
	return makeIptcData(i, C.exiv2_image_get_iptc_data(i.img))
}

func (d *IptcData) FindKey(key string) (*IptcDatum, error) {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))

	var cerr *C.Exiv2Error

	cdatum := C.exiv2_iptc_data_find_key(d.data, ckey, &cerr)

	if cerr != nil {
		err := makeError(cerr)
		C.exiv2_error_free(cerr)
		return nil, err
	}

	return makeIptcDatum(d, cdatum), nil
}

func (d *IptcDatum) String() string {
	cstr := C.exiv2_iptc_datum_to_string(d.datum)
	defer C.free(unsafe.Pointer(cstr))

	return C.GoString(cstr)
}
