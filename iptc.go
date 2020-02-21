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

// IptcDatumIterator wraps the respective C++ structure.
type IptcDatumIterator struct {
	data *IptcData
	iter *C.Exiv2IptcDatumIterator
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

func (i *Image) SetIptcString(key, value string) error {
	return i.SetMetadataString("iptc", key, value)
}

func (d *IptcData) GetString(key string) (string, error) {
	datum, err := d.FindKey(key)
	if err != nil {
		return "", err
	}

	if datum == nil {
		return "", errMetadataKeyNotFound
	}

	return datum.String(), nil
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

// Key returns the IPTC key of the datum.
func (d *IptcDatum) Key() string {
	return C.GoString(C.exiv2_iptc_datum_key(d.datum))
}

func (d *IptcDatum) String() string {
	cstr := C.exiv2_iptc_datum_to_string(d.datum)
	defer C.free(unsafe.Pointer(cstr))

	return C.GoString(cstr)
}

// Iterator returns a new IptcDatumIterator to iterate over all IPTC data.
func (d *IptcData) Iterator() *IptcDatumIterator {
	return makeIptcDatumIterator(d, C.exiv2_iptc_data_iterator(d.data))
}

// HasNext returns true as long as the iterator has another datum to deliver.
func (i *IptcDatumIterator) HasNext() bool {
	return C.exiv2_iptc_data_iterator_has_next(i.iter) != 0
}

// Next returns the next IptcDatum of the iterator or nil if iterator has reached the end.
func (i *IptcDatumIterator) Next() *IptcDatum {
	return makeIptcDatum(i.data, C.exiv2_iptc_datum_iterator_next(i.iter))
}

func makeIptcDatumIterator(data *IptcData, cIter *C.Exiv2IptcDatumIterator) *IptcDatumIterator {
	datum := &IptcDatumIterator{data, cIter}

	runtime.SetFinalizer(datum, func(i *IptcDatumIterator) {
		C.exiv2_iptc_datum_iterator_free(i.iter)
	})

	return datum
}
