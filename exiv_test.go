package goexiv_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toaster/goexiv"
)

func TestOpenImage(t *testing.T) {
	wd, _ := os.Getwd()
	testImage := path.Join(wd, "pixel.jpg")

	// Open valid file

	img, err := goexiv.Open(testImage)

	if err != nil {
		t.Fatalf("Cannot open image: %s", err)
	}

	if img == nil {
		t.Fatalf("img is nil after successful open")
	}

	// Open non existing file

	img, err = goexiv.Open("thisimagedoesnotexist")

	if err == nil {
		t.Fatalf("No error set after opening a non existing image")
	}

	exivErr, ok := err.(*goexiv.Error)

	if !ok {
		t.Fatalf("Returned error is not of type Error")
	}

	if exivErr.Code() != 9 {
		t.Fatalf("Unexpected error code (expected 9, got %d)", exivErr.Code())
	}
}

func Test_OpenBytes(t *testing.T) {
	wd, _ := os.Getwd()
	testImage := path.Join(wd, "pixel.jpg")
	bytes, err := ioutil.ReadFile(testImage)
	require.NoError(t, err)

	img, err := goexiv.OpenBytes(bytes)
	if assert.NoError(t, err) {
		assert.NotNil(t, img)
	}
}

func Test_OpenBytesFailures(t *testing.T) {
	tests := []struct {
		name        string
		bytes       []byte
		wantErr     string
		wantErrCode int
	}{
		{
			"no image",
			[]byte("no image"),
			"The memory contains data of an unknown image type",
			12,
		},
		{
			"empty byte slice",
			[]byte{},
			"input is empty",
			0,
		},
		{
			"nil byte slice",
			nil,
			"input is empty",
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := goexiv.OpenBytes(tt.bytes)
			if assert.EqualError(t, err, tt.wantErr) {
				exivErr, ok := err.(*goexiv.Error)
				if assert.True(t, ok, "occurred error is not of Type goexiv.Error") {
					assert.Equal(t, tt.wantErrCode, exivErr.Code(), "unexpected error code")
				}
			}
		})
	}
}

func TestMetadata(t *testing.T) {
	wd, _ := os.Getwd()
	testImage := path.Join(wd, "pixel.jpg")

	img, _ := goexiv.Open(testImage)

	err := img.ReadMetadata()

	if err != nil {
		t.Fatalf("Cannot read image metadata: %s", err)
	}

	width := img.PixelWidth()
	height := img.PixelHeight()
	if width != 1 || height != 1 {
		t.Errorf("Cannot read image size (expected 1x1, got %dx%d)", width, height)
	}

	data := img.GetExifData()

	// Invalid key
	datum, err := data.FindKey("NotARealKey")

	if err == nil {
		t.Fatalf("FindKey returns a nil error for an invalid key")
	}

	if datum != nil {
		t.Fatalf("FindKey does not return nil for an invalid key")
	}

	// Valid, existing key

	datum, err = data.FindKey("Exif.Image.Make")

	if err != nil {
		t.Fatalf("FindKey returns an error for a valid, existing key: %s", err)
	}

	if datum == nil {
		t.Fatalf("FindKey returns nil for a valid, existing key")
	}

	if datum.String() != "FakeMake" {
		t.Fatalf("Unexpected value for EXIF datum Exif.Image.Make (expected 'FakeMake', got '%s')", datum.String())
	}

	// Valid, non existing key

	datum, err = data.FindKey("Exif.Photo.Flash")

	if err != nil {
		t.Fatalf("FindKey returns an error for a valid, non existing key: %s", err)
	}

	if datum != nil {
		t.Fatalf("FindKey returns a non null datum for a valid, non existing key")
	}

	// Iterate over all Exif data accessing Key() and String()
	{
		keyValues := map[string]string{}
		for i := data.Iterator(); i.HasNext(); {
			d := i.Next()
			keyValues[d.Key()] = d.String()
		}
		assert.Equal(t, keyValues, map[string]string{
			"Exif.Image.ExifTag":                 "134",
			"Exif.Image.Make":                    "FakeMake",
			"Exif.Image.Model":                   "FakeModel",
			"Exif.Image.ResolutionUnit":          "2",
			"Exif.Image.XResolution":             "72/1",
			"Exif.Image.YCbCrPositioning":        "1",
			"Exif.Image.YResolution":             "72/1",
			"Exif.Photo.ColorSpace":              "65535",
			"Exif.Photo.ComponentsConfiguration": "1 2 3 0",
			"Exif.Photo.DateTimeDigitized":       "2013:12:08 21:06:10",
			"Exif.Photo.ExifVersion":             "48 50 51 48",
			"Exif.Photo.FlashpixVersion":         "48 49 48 48",
		})
	}

	//
	// IPTC
	//
	iptcData := img.GetIptcData()

	// Iterate over all IPCT data accessing Key() and String()
	{
		keyValues := map[string]string{}
		for i := iptcData.Iterator(); i.HasNext(); {
			d := i.Next()
			keyValues[d.Key()] = d.String()
		}
		assert.Equal(t, keyValues, map[string]string{
			"Iptc.Application2.Copyright":   "this is the copy, right?",
			"Iptc.Application2.CountryName": "Lancre",
			"Iptc.Application2.DateCreated": "1848-10-13",
			"Iptc.Application2.TimeCreated": "12:49:32+01:00",
		})
	}
}
