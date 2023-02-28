package goexiv

// #cgo pkg-config: exiv2
// #include "helper.h"
// #include <stdlib.h>
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

type Error struct {
	code int
	what string
}

type Image struct {
	bytesArrayPtr unsafe.Pointer
	img           *C.Exiv2Image
}

type MetadataProvider interface {
	GetString(key string) (string, error)
}

type MetadataFormat int

const (
	EXIF MetadataFormat = iota
	IPTC
	XMP
)

var ErrMetadataKeyNotFound = errors.New("key not found")

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

func makeImage(cimg *C.Exiv2Image, bytesPtr unsafe.Pointer) *Image {
	img := &Image{
		bytesArrayPtr: bytesPtr,
		img:           cimg,
	}

	runtime.SetFinalizer(img, func(x *Image) {
		C.exiv2_image_free(x.img)

		if x.bytesArrayPtr != nil {
			C.free(x.bytesArrayPtr)
		}
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

	return makeImage(cimg, nil), nil
}

// OpenBytes opens a byte slice with image data and returns a pointer to
// the corresponding Image object, but does not read the Metadata.
// Start the parsing with a call to ReadMetadata()
func OpenBytes(input []byte) (*Image, error) {
	if len(input) == 0 {
		return nil, &Error{0, "input is empty"}
	}

	var cerr *C.Exiv2Error

	bytesArrayPtr := C.CBytes(input)
	cimg := C.exiv2_image_factory_open_bytes(
		(*C.uchar)(bytesArrayPtr),
		C.long(len(input)),
		&cerr,
	)

	if cerr != nil {
		err := makeError(cerr)
		C.exiv2_error_free(cerr)
		return nil, err
	}

	return makeImage(cimg, bytesArrayPtr), nil
}

type LogMsgLevel int

const (
	LogMsgDebug LogMsgLevel = 0
	LogMsgInfo              = 1
	LogMsgWarn              = 2
	LogMsgError             = 3
	LogMsgMute              = 4
)

// SetLogMsgLevel Set the log level (outputs to stderr)
func SetLogMsgLevel(level LogMsgLevel) {
	C.exiv2_log_msg_set_level(C.int(level))
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

// Returns an image contents.
// If its metadata has been changed, the changes are reflected here.
func (i *Image) GetBytes() []byte {
	size := C.exiv_image_get_size(i.img)
	ptr := C.exiv_image_get_bytes_ptr(i.img)

	return C.GoBytes(unsafe.Pointer(ptr), C.int(size))
}

// PixelWidth returns the width of the image in pixels
func (i *Image) PixelWidth() int64 {
	return int64(C.exiv2_image_get_pixel_width(i.img))
}

// PixelHeight returns the height of the image in pixels
func (i *Image) PixelHeight() int64 {
	return int64(C.exiv2_image_get_pixel_height(i.img))
}

// ICCProfile returns the ICC profile or nil if the image doesn't has one.
func (i *Image) ICCProfile() []byte {
	size := C.int(C.exiv2_image_icc_profile_size(i.img))
	if size <= 0 {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(C.exiv2_image_icc_profile(i.img)), size)
}

// Sets an exif or iptc key with a given string value
func (i *Image) SetMetadataString(format, key, value string) error {
	if format != "iptc" && format != "exif" {
		return errors.New("invalid metadata type: " + format)
	}

	cKey := C.CString(key)
	cValue := C.CString(value)

	defer func() {
		C.free(unsafe.Pointer(cKey))
		C.free(unsafe.Pointer(cValue))
	}()

	var cerr *C.Exiv2Error

	if format == "iptc" {
		C.exiv2_image_set_iptc_string(i.img, cKey, cValue, &cerr)
	} else {
		C.exiv2_image_set_exif_string(i.img, cKey, cValue, &cerr)
	}

	if cerr != nil {
		err := makeError(cerr)
		C.exiv2_error_free(cerr)
		return err
	}

	return nil
}

// Sets an exif or iptc key with a given short value
func (i *Image) SetMetadataShort(format, key, value string) error {
	if format != "iptc" && format != "exif" {
		return errors.New("invalid metadata type: " + format)
	}

	cKey := C.CString(key)
	cValue := C.CString(value)

	defer func() {
		C.free(unsafe.Pointer(cKey))
		C.free(unsafe.Pointer(cValue))
	}()

	var cerr *C.Exiv2Error

	if format == "iptc" {
		C.exiv2_image_set_iptc_short(i.img, cKey, cValue, &cerr)
	} else {
		C.exiv2_image_set_exif_short(i.img, cKey, cValue, &cerr)
	}

	if cerr != nil {
		err := makeError(cerr)
		C.exiv2_error_free(cerr)
		return err
	}

	return nil
}

func (i *Image) StripKey(f MetadataFormat, key string) error {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))

	var cErr *C.Exiv2Error

	switch f {
	case EXIF:
		C.exiv2_exif_strip_key(i.img, ckey, &cErr)
	case IPTC:
		C.exiv2_iptc_strip_key(i.img, ckey, &cErr)
	case XMP:
		C.exiv2_xmp_strip_key(i.img, ckey, &cErr)
	default:
		return errors.New("invalid metadata format")
	}

	if cErr != nil {
		err := makeError(cErr)
		C.exiv2_error_free(cErr)
		return err
	}

	return nil
}
