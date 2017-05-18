#ifdef __cplusplus
extern "C" {
#endif

#define DECLARE_STRUCT(name) typedef struct _##name name

DECLARE_STRUCT(Exiv2ImageFactory);
DECLARE_STRUCT(Exiv2Image);
DECLARE_STRUCT(Exiv2XmpData);
DECLARE_STRUCT(Exiv2XmpDatum);
DECLARE_STRUCT(Exiv2IptcData);
DECLARE_STRUCT(Exiv2IptcDatum);
DECLARE_STRUCT(Exiv2ExifData);
DECLARE_STRUCT(Exiv2ExifDatum);
DECLARE_STRUCT(Exiv2Error);

Exiv2Image* exiv2_image_factory_open(const char *path, Exiv2Error **error);

void exiv2_image_read_metadata(Exiv2Image *img, Exiv2Error **error);
void exiv2_image_free(Exiv2Image *img);

int exiv2_image_get_pixel_width(Exiv2Image *img);
int exiv2_image_get_pixel_height(Exiv2Image *img);

Exiv2XmpData* exiv2_image_get_xmp_data(const Exiv2Image *img);
void exiv2_xmp_data_free(Exiv2XmpData *data);
char* exiv2_xmp_datum_to_string(const Exiv2XmpDatum *datum);
void exiv2_xmp_datum_free(Exiv2XmpDatum *datum);
Exiv2XmpDatum* exiv2_xmp_data_find_key(const Exiv2XmpData *data, const char *key, Exiv2Error **error);

Exiv2IptcData* exiv2_image_get_iptc_data(const Exiv2Image *img);
void exiv2_iptc_data_free(Exiv2IptcData *data);
char* exiv2_iptc_datum_to_string(const Exiv2IptcDatum *datum);
void exiv2_iptc_datum_free(Exiv2IptcDatum *datum);
Exiv2IptcDatum* exiv2_iptc_data_find_key(const Exiv2IptcData *data, const char *key, Exiv2Error **error);

Exiv2ExifData* exiv2_image_get_exif_data(const Exiv2Image *img);
char* exiv2_exif_datum_to_string(const Exiv2ExifDatum *datum);
void exiv2_exif_datum_free(Exiv2ExifDatum *datum);
void exiv2_exif_data_free(Exiv2ExifData *data);
Exiv2ExifDatum* exiv2_exif_data_find_key(const Exiv2ExifData *data, const char *key, Exiv2Error **error);

int exiv2_error_code(const Exiv2Error *e);
const char *exiv2_error_what(const Exiv2Error *e);
void exiv2_error_free(Exiv2Error *e);

#ifdef __cplusplus
} // extern "C"
#endif
