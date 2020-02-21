package goexiv_test

import (
	"io/ioutil"
	"testing"

	"github.com/gitschneider/goexiv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenImage(t *testing.T) {
	// Open valid file
	img, err := goexiv.Open("testdata/pixel.jpg")

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
	bytes, err := ioutil.ReadFile("testdata/pixel.jpg")
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
	img, err := goexiv.Open("testdata/pixel.jpg")
	require.NoError(t, err)

	err = img.ReadMetadata()

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

	//
	// ICC profile
	//
	iccProfile := img.ICCProfile()
	assert.Equal(t,
		// 128 bytes header
		"\x00\x00\x02\x30"+
			"ADBE\x02\x10\x00\x00"+
			"mntrRGB XYZ \x07\xcf\x00\x06\x00\x03\x00\x00\x00\x00\x00\x00"+
			"acspAPPL\x00\x00\x00\x00"+
			"none\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"+
			"\x00\x00\xf6\xd6\x00\x01\x00\x00\x00\x00\xd3-"+
			"ADBE\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"+
			"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"+
			"\x00\x00\x00\x00\x00\x00"+
			// tag table (124 bytes)
			// tag count (10)
			"\x00\x00\x00\x0a"+
			// tag references (4 bytes signature, 4 bytes position from start of profile, 4 bytes length)
			"cprt\x00\x00\x00\xfc\x00\x00\x00\x32"+
			"desc\x00\x00\x01\x30\x00\x00\x00k"+
			"wtpt\x00\x00\x01\x9c\x00\x00\x00\x14"+
			"bkpt\x00\x00\x01\xb0\x00\x00\x00\x14"+
			"rTRC\x00\x00\x01\xc4\x00\x00\x00\x0e"+
			"gTRC\x00\x00\x01\xd4\x00\x00\x00\x0e"+
			"bTRC\x00\x00\x01\xe4\x00\x00\x00\x0e"+
			"rXYZ\x00\x00\x01\xf4\x00\x00\x00\x14"+
			"gXYZ\x00\x00\x02\b\x00\x00\x00\x14"+
			"bXYZ\x00\x00\x02\x1c\x00\x00\x00\x14"+
			// tagged element data (308 bytes; sum of the length of the ten tags)
			"text\x00\x00\x00\x00Copyright 1999 Adobe Systems Incorporated\x00\x00\x00"+
			"desc\x00\x00\x00\x00\x00\x00\x00\x11Adobe RGB (1998)"+
			"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"+
			"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"+
			"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"+
			"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"+
			"XYZ \x00\x00\x00\x00\x00\x00\xf3Q\x00\x01\x00\x00\x00\x01\x16\xcc"+
			"XYZ \x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"+
			"curv\x00\x00\x00\x00\x00\x00\x00\x01\x023\x00\x00"+
			"curv\x00\x00\x00\x00\x00\x00\x00\x01\x023\x00\x00"+
			"curv\x00\x00\x00\x00\x00\x00\x00\x01\x023\x00\x00"+
			"XYZ \x00\x00\x00\x00\x00\x00\x9c\x18\x00\x00O\xa5\x00\x00\x04\xfc"+
			"XYZ \x00\x00\x00\x00\x00\x004\x8d\x00\x00\xa0,\x00\x00\x0f\x95"+
			"XYZ \x00\x00\x00\x00\x00\x00&1\x00\x00\x10/\x00\x00\xbe\x9c",
		string(iccProfile),
	)
}

func TestNoMetadata(t *testing.T) {
	img, err := goexiv.Open("testdata/stripped_pixel.jpg")
	require.NoError(t, err)
	err = img.ReadMetadata()
	require.NoError(t, err)

	// no ICC profile

	assert.Nil(t, img.ICCProfile())
}

type MetadataTestCase struct {
	Format                 string // exif or iptc
	Key                    string
	Value                  string
	ImageFilename          string
	ExpectedErrorSubstring string
}

func TestSetMetadataString(t *testing.T) {
	cases := []MetadataTestCase{
		// valid exif key, jpeg
		{
			Format:                 "exif",
			Key:                    "Exif.Photo.UserComment",
			Value:                  "Hello, world! Привет, мир!",
			ImageFilename:          "testdata/pixel.jpg",
			ExpectedErrorSubstring: "", // no error
		},
		// valid exif key, webp
		{
			Format:                 "exif",
			Key:                    "Exif.Photo.UserComment",
			Value:                  "Hello, world! Привет, мир!",
			ImageFilename:          "testdata/pixel.webp",
			ExpectedErrorSubstring: "",
		},
		// valid iptc key, jpeg.
		// webp iptc is not supported (see libexiv2/src/webpimage.cpp WebPImage::setIptcData))
		{
			Format:                 "iptc",
			Key:                    "Iptc.Application2.Caption",
			Value:                  "Hello, world! Привет, мир!",
			ImageFilename:          "testdata/pixel.jpg",
			ExpectedErrorSubstring: "",
		},
		// invalid exif key, jpeg
		{
			Format:                 "exif",
			Key:                    "Exif.Invalid.Key",
			Value:                  "this value should not be written",
			ImageFilename:          "testdata/pixel.jpg",
			ExpectedErrorSubstring: "Invalid key",
		},
		// invalid exif key, webp
		{
			Format:                 "exif",
			Key:                    "Exif.Invalid.Key",
			Value:                  "this value should not be written",
			ImageFilename:          "testdata/pixel.webp",
			ExpectedErrorSubstring: "Invalid key",
		},
		// invalid iptc key, jpeg
		{
			Format:                 "iptc",
			Key:                    "Iptc.Invalid.Key",
			Value:                  "this value should not be written",
			ImageFilename:          "testdata/pixel.jpg",
			ExpectedErrorSubstring: "Invalid record name",
		},
	}

	var data goexiv.MetadataProvider

	for i, testcase := range cases {
		img, err := goexiv.Open(testcase.ImageFilename)
		require.NoErrorf(t, err, "case #%d Error while opening image file", i)

		err = img.SetMetadataString(testcase.Format, testcase.Key, testcase.Value)
		if testcase.ExpectedErrorSubstring != "" {
			require.Errorf(t, err, "case #%d Error was expected", i)
			require.Containsf(
				t,
				err.Error(),
				testcase.ExpectedErrorSubstring,
				"case #%d Error text must contain a given substring",
				i,
			)
			continue
		} else {
			require.NoErrorf(t, err, "case #%d Cannot write image metadata", i)
		}

		err = img.ReadMetadata()
		require.NoErrorf(t, err, "case #%d Cannot read image metadata", i)

		if testcase.Format == "iptc" {
			data = img.GetIptcData()
		} else {
			data = img.GetExifData()
		}

		receivedValue, err := data.GetString(testcase.Key)
		require.Equalf(
			t,
			testcase.Value,
			receivedValue,
			"case #%d Value written must be equal to the value read",
			i,
		)
	}
}
