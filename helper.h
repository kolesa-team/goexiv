#ifdef __cplusplus
extern "C" {
#endif

#define DECLARE_STRUCT(name) typedef struct _##name name

DECLARE_STRUCT(Exiv2ImageFactory);
DECLARE_STRUCT(Exiv2Image);
DECLARE_STRUCT(Exiv2ExifData);
DECLARE_STRUCT(Exiv2ExifDatum);
DECLARE_STRUCT(Exiv2Error);

Exiv2Image* exiv2_image_factory_open(const char *path, Exiv2Error **error);

void exiv2_image_read_metadata(Exiv2Image *img, Exiv2Error **error);
Exiv2ExifData* exiv2_image_get_exif_data(const Exiv2Image *img);
void exiv2_image_free(Exiv2Image *img);

Exiv2ExifDatum* exiv2_exif_data_find_key(const Exiv2ExifData *data, const char *key, Exiv2Error **error);
void exiv2_exif_data_free(Exiv2ExifData *data);

char* exiv2_exif_datum_to_string(const Exiv2ExifDatum *datum);
void exiv2_exif_datum_free(Exiv2ExifDatum *datum);

int exiv2_error_code(const Exiv2Error *e);
const char *exiv2_error_what(const Exiv2Error *e);
void exiv2_error_free(Exiv2Error *e);

#ifdef __cplusplus
} // extern "C"
#endif
