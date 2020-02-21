[![Build Status](https://travis-ci.org/toaster/goexiv.svg)](https://travis-ci.org/toaster/goexiv.svg)

# Go bindings for exiv2 (http://www.exiv2.org)

Those bindings are at the moment (very) incomplete, but already allow you to
read the metadata of a file, get the EXIF fields out of it, and add EXIF/IPTC string fields. Binding coverage
will be extended as I start needing more methods (or receiving more pull
requests :) ).

## Requirements

You need to have libexiv2 installed at version 0.27 at least.

### Debian/Ubuntu

```bash
sudo apt install libexiv2-dev
```

Note: Because Exiv2 is a C++ library, you probably need Go 1.2 to benefit from
the improved C++ support in CGO.
