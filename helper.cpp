#include "helper.h"

#include <exiv2/image.hpp>

#include <stdio.h>

#define DEFINE_STRUCT(name,wrapped_type,member_name) \
struct _##name { \
	_##name(wrapped_type member_name) \
		: member_name(member_name) {} \
	wrapped_type member_name; \
};

#define DEFINE_FREE_FUNCTION(name,type) \
void name##_free(type x) \
{ \
	delete x; \
}

DEFINE_STRUCT(Exiv2ImageFactory, Exiv2::ImageFactory*, factory);
DEFINE_STRUCT(Exiv2Image, Exiv2::Image::AutoPtr, image);

DEFINE_STRUCT(Exiv2XmpData, const Exiv2::XmpData&, data);
DEFINE_STRUCT(Exiv2XmpDatum, const Exiv2::Xmpdatum&, datum);

DEFINE_STRUCT(Exiv2ExifData, const Exiv2::ExifData&, data);
DEFINE_STRUCT(Exiv2ExifDatum, const Exiv2::Exifdatum&, datum);

DEFINE_STRUCT(Exiv2IptcData, const Exiv2::IptcData&, data);
DEFINE_STRUCT(Exiv2IptcDatum, const Exiv2::Iptcdatum&, datum);

struct _Exiv2Error {
	_Exiv2Error(const Exiv2::Error &error);

	int code;
	char *what;
};

_Exiv2Error::_Exiv2Error(const Exiv2::Error &error)
	: code(error.code())
	, what(strdup(error.what()))
{
}

Exiv2Image*
exiv2_image_factory_open(const char *path, Exiv2Error **error)
{
	Exiv2Image *p = 0;

	try {
		p = new Exiv2Image(Exiv2::ImageFactory::open(path));
		return p;
	} catch (Exiv2::Error &e) {
		delete p;

		if (error) {
			*error = new Exiv2Error(e);
		}
	}

	return 0;
}

void
exiv2_image_read_metadata(Exiv2Image *img, Exiv2Error **error)
{
	try {
		img->image->readMetadata();
	} catch (Exiv2::Error &e) {
		if (error) {
			*error = new Exiv2Error(e);
		}
	}
}

DEFINE_FREE_FUNCTION(exiv2_image, Exiv2Image*);

// XMP
Exiv2XmpData*
exiv2_image_get_xmp_data(const Exiv2Image *img)
{
	return new Exiv2XmpData(img->image->xmpData());
}

Exiv2XmpDatum*
exiv2_xmp_data_find_key(const Exiv2XmpData *data, const char *key, Exiv2Error **error)
{
	try {
		const Exiv2::XmpData::const_iterator it = data->data.findKey(Exiv2::XmpKey(key));
		if (it == data->data.end()) {
			return 0;
		}

		return new Exiv2XmpDatum(*it);
	} catch (Exiv2::Error &e) {
		if (error) {
			*error = new Exiv2Error(e);
		}

		return 0;
	}
}

DEFINE_FREE_FUNCTION(exiv2_xmp_data, Exiv2XmpData*);

char*
exiv2_xmp_datum_to_string(const Exiv2XmpDatum *datum)
{
    Exiv2::TypeId typeId = datum->datum.typeId();

    const std::string strval;

    if (typeId == Exiv2::xmpBag) {
        strval = datum->datum.toString();
    } else {
        strval = datum->datum.toString(0);
    }

	return strdup(strval.c_str());
}

DEFINE_FREE_FUNCTION(exiv2_xmp_datum, Exiv2XmpDatum*);

// IPTC

Exiv2IptcData*
exiv2_image_get_iptc_data(const Exiv2Image *img)
{
	return new Exiv2IptcData(img->image->iptcData());
}

Exiv2IptcDatum*
exiv2_iptc_data_find_key(const Exiv2IptcData *data, const char *key, Exiv2Error **error)
{
	try {
		const Exiv2::IptcData::const_iterator it = data->data.findKey(Exiv2::IptcKey(key));
		if (it == data->data.end()) {
			return 0;
		}

		return new Exiv2IptcDatum(*it);
	} catch (Exiv2::Error &e) {
		if (error) {
			*error = new Exiv2Error(e);
		}

		return 0;
	}
}

DEFINE_FREE_FUNCTION(exiv2_iptc_data, Exiv2IptcData*);

char*
exiv2_iptc_datum_to_string(const Exiv2IptcDatum *datum)
{
	const std::string strval = datum->datum.toString(0);
	return strdup(strval.c_str());
}

DEFINE_FREE_FUNCTION(exiv2_iptc_datum, Exiv2IptcDatum*);

// EXIF

Exiv2ExifData*
exiv2_image_get_exif_data(const Exiv2Image *img)
{
	return new Exiv2ExifData(img->image->exifData());
}

Exiv2ExifDatum*
exiv2_exif_data_find_key(const Exiv2ExifData *data, const char *key, Exiv2Error **error)
{
	try {
		const Exiv2::ExifData::const_iterator it = data->data.findKey(Exiv2::ExifKey(key));
		if (it == data->data.end()) {
			return 0;
		}

		return new Exiv2ExifDatum(*it);
	} catch (Exiv2::Error &e) {
		if (error) {
			*error = new Exiv2Error(e);
		}

		return 0;
	}
}

DEFINE_FREE_FUNCTION(exiv2_exif_data, Exiv2ExifData*);

char*
exiv2_exif_datum_to_string(const Exiv2ExifDatum *datum)
{
	const std::string strval = datum->datum.toString();
	return strdup(strval.c_str());
}

DEFINE_FREE_FUNCTION(exiv2_exif_datum, Exiv2ExifDatum*);

// ERRORS

int
exiv2_error_code(const Exiv2Error *error)
{
	return error->code;
}

const char*
exiv2_error_what(const Exiv2Error *error)
{
	return error->what;
}

void
exiv2_error_free(Exiv2Error *e)
{
	if (e == 0) {
		return;
	}

	free(e->what);
	delete e;
}
