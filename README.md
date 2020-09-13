[![Build Status](https://api.travis-ci.org/kolesa-team/goexiv.svg)](https://api.travis-ci.org/kolesa-team/goexiv.svg)

# Go bindings for exiv2 (http://www.exiv2.org)

The library allows reading and writing EXIF and IPTC metadata to/from JPG, WEBP, and PNG images.

It is based on https://github.com/abustany/goexiv and https://github.com/gitschneider/goexiv with support added for writing the metadata and various bugfixes.

Библиотека для записи и чтения метаданных EXIF и IPTC в изображениях формата JPG, WEBP и PNG.
Основана на https://github.com/abustany/goexiv и https://github.com/gitschneider/goexiv с добавленной фукнциональностью для записи метаданных и исправлением ошибок.

## Requirements

A [libexiv2](http://www.exiv2.org) library v0.27 is required (this might also work with newer versions, but it hasn't been tested).

On Ubuntu, libexiv2 can be installed from the package manager (`sudo apt install libexiv2-dev`), but there is no guarantee it comes with the version needed.
So it is safer to install it manually:

* Download and unpack the library from `https://github.com/Exiv2/exiv2/releases/tag/v0.27.2`
* Install the library (the steps are taken from libexiv2 README):
    ```
    mkdir build && cd build
    cmake .. -DCMAKE_BUILD_TYPE=Release
    cmake --build .
    sudo make install
    sudo ldconfig
    ```
* It may be needed to set the following variable: `export PKG_CONFIG_PATH=/usr/local/lib64/pkgconfig`

Now the Go code in this project can interface with the libexiv2 library.

The installation process for other operating systems should be similar.
Also, this library is tested with `golang:1.13-alpine` docker image, where the correct version of libexiv2 is installed with `apk --update add exiv2-dev`.

## Usage

Basic usage:

```
import "github.com/kolesa-team/goexiv"

// Open an image from disk
goexivImg, err := goexiv.Open("/path/to/image.jpg")
if err != nil {
    return err
}

// Write an EXIF comment
err = goexivImg.SetMetadataString("exif", "Exif.Photo.UserComment", "A comment. Might be a JSON string. Можно писать и по-русски!")
if err != nil {
    return err
}

// Read metadata
err = goexivImg.ReadMetadata()
if err != nil {
    return err
}

// Read an EXIF comment
userComment, err := goexivImg.GetExifData().GetString("Exif.Photo.UserComment")
if err != nil {
    return err
}

fmt.Println(userComment)
// "A comment. Might be a JSON string. Можно писать и по-русски!"
```

Changing the image metadata in memory and returning the updated image (an approach fit for a web service):

```
// Say we have an image in memory: var img []byte
goexivImg, err := goexiv.OpenBytes(img)
if err != nil {
    return err
}

// Write an IPTC comment
err = goexivImg.SetMetadataString("iptc", "Iptc.Application2.Caption", "A comment. Might be a JSON string.")
if err != nil {
    return err
}

// Get back the modified image, so it can now be further processed (e.g. sent over the network)
img = goexivImg.GetBytes()
```

Retrieving all metadata keys and values:

```
img.ReadMetadata()
// map[string]string
exif := img.GetExifData().AllTags()

// map[string]string
iptc := img.GetIptcData().AllTags()
```

A complete image processing workflow in Go can be organized with the following additional libraries:

* https://github.com/kolesa-team/go-webp - Go bindings for libwebp to process WEBP images
* https://github.com/lEx0/go-libjpeg-nrgba - Go bindings for libjpeg-turbo, a fast JPEG processing library.
* https://github.com/disintegration/imaging - a generic Go library for working with images (covers many formats, but is not as fast as the libraries above, so it can be used as a fallback for PNG and GIF)
