package goexiv_test

import (
	"github.com/kolesa-team/goexiv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"runtime"
	"sync"
	"testing"
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
			"Failed to read input data",
			20,
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
	initializeImage("testdata/pixel.jpg", t)
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

	assert.Equal(t, map[string]string{
		"Exif.Image.ExifTag":                 "130",
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
	}, data.AllTags())

	//
	// IPTC
	//
	iptcData := img.GetIptcData()
	assert.Equal(t, map[string]string{
		"Iptc.Application2.Copyright":   "this is the copy, right?",
		"Iptc.Application2.CountryName": "Lancre",
		"Iptc.Application2.DateCreated": "2012-10-13",
		"Iptc.Application2.TimeCreated": "12:49:32+01:00",
	}, iptcData.AllTags())
}

func TestNoMetadata(t *testing.T) {
	img, err := goexiv.Open("testdata/stripped_pixel.jpg")
	require.NoError(t, err)
	err = img.ReadMetadata()
	require.NoError(t, err)
	assert.Nil(t, img.ICCProfile())
}

type MetadataTestCase struct {
	Format                 string // exif or iptc
	Key                    string
	Value                  string
	ImageFilename          string
	ExpectedErrorSubstring string
}

var metadataSetStringTestCases = []MetadataTestCase{
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

func Test_SetMetadataStringFromFile(t *testing.T) {
	var data goexiv.MetadataProvider

	for i, testcase := range metadataSetStringTestCases {
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
		}

		require.NoErrorf(t, err, "case #%d Cannot write image metadata", i)

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

var metadataSetShortIntTestCases = []MetadataTestCase{
	// valid exif key, jpeg
	{
		Format:                 "exif",
		Key:                    "Exif.Photo.ExposureProgram",
		Value:                  "1",
		ImageFilename:          "testdata/pixel.jpg",
		ExpectedErrorSubstring: "", // no error
	},
	// valid exif key, webp
	{
		Format:                 "exif",
		Key:                    "Exif.Photo.ExposureProgram",
		Value:                  "2",
		ImageFilename:          "testdata/pixel.webp",
		ExpectedErrorSubstring: "",
	},
	// valid iptc key, jpeg.
	// webp iptc is not supported (see libexiv2/src/webpimage.cpp WebPImage::setIptcData))
	{
		Format:                 "iptc",
		Key:                    "Iptc.Envelope.ModelVersion",
		Value:                  "3",
		ImageFilename:          "testdata/pixel.jpg",
		ExpectedErrorSubstring: "",
	},
	// invalid exif key, jpeg
	{
		Format:                 "exif",
		Key:                    "Exif.Invalid.Key",
		Value:                  "4",
		ImageFilename:          "testdata/pixel.jpg",
		ExpectedErrorSubstring: "Invalid key",
	},
	// invalid exif key, webp
	{
		Format:                 "exif",
		Key:                    "Exif.Invalid.Key",
		Value:                  "5",
		ImageFilename:          "testdata/pixel.webp",
		ExpectedErrorSubstring: "Invalid key",
	},
	// invalid iptc key, jpeg
	{
		Format:                 "iptc",
		Key:                    "Iptc.Invalid.Key",
		Value:                  "6",
		ImageFilename:          "testdata/pixel.jpg",
		ExpectedErrorSubstring: "Invalid record name",
	},
}

func Test_SetMetadataShortInt(t *testing.T) {
	var data goexiv.MetadataProvider

	for i, testcase := range metadataSetShortIntTestCases {
		img, err := goexiv.Open(testcase.ImageFilename)
		require.NoErrorf(t, err, "case #%d Error while opening image file", i)

		err = img.SetMetadataShort(testcase.Format, testcase.Key, testcase.Value)
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
		}

		require.NoErrorf(t, err, "case #%d Cannot write image metadata", i)

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

func Test_GetBytes(t *testing.T) {
	bytes, err := ioutil.ReadFile("testdata/stripped_pixel.jpg")
	require.NoError(t, err)

	img, err := goexiv.OpenBytes(bytes)
	require.NoError(t, err)

	require.Equal(
		t,
		len(bytes),
		len(img.GetBytes()),
		"Image size on disk and in memory must be equal",
	)

	bytesBeforeTag := img.GetBytes()
	assert.NoError(t, img.SetExifString("Exif.Photo.UserComment", "123"))
	bytesAfterTag := img.GetBytes()
	assert.True(t, len(bytesAfterTag) > len(bytesBeforeTag), "Image size must increase after adding an EXIF tag")
	assert.Equal(t, &bytesBeforeTag[0], &bytesAfterTag[0], "Every call to GetBytes must point to the same underlying array")

	assert.NoError(t, img.SetExifString("Exif.Photo.UserComment", "123"))
	bytesAfterTag2 := img.GetBytes()
	assert.Equal(
		t,
		len(bytesAfterTag),
		len(bytesAfterTag2),
		"Image size must not change after the same tag has been set",
	)
}

// Ensures image manipulation doesn't fail when running from multiple goroutines
func Test_GetBytes_Goroutine(t *testing.T) {
	var wg sync.WaitGroup
	iterations := 0

	bytes, err := ioutil.ReadFile("testdata/stripped_pixel.jpg")
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		iterations++
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			img, err := goexiv.OpenBytes(bytes)
			require.NoError(t, err)

			// trigger garbage collection to increase the chance that underlying img.img will be collected
			runtime.GC()

			bytesAfter := img.GetBytes()
			assert.NotEmpty(t, bytesAfter)

			// if this line is removed, then the test will likely fail
			// with segmentation violation.
			// so far we couldn't come up with a better solution.
			runtime.KeepAlive(img)
		}(i)
	}

	wg.Wait()
	runtime.GC()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	t.Logf("Allocated bytes after test:  %+v\n", memStats.HeapAlloc)
}

func TestExifStripKey(t *testing.T) {
	img, err := goexiv.Open("testdata/pixel.jpg")
	require.NoError(t, err)

	err = img.SetExifString("Exif.Photo.UserComment", "123")
	require.NoError(t, err)

	err = img.ExifStripKey("Exif.Photo.UserComment")
	require.NoError(t, err)

	err = img.ReadMetadata()
	require.NoError(t, err)

	data := img.GetExifData()

	_, err = data.GetString("Exif.Photo.UserComment")
	require.Error(t, err)
}

func TestIptcStripKey(t *testing.T) {
	img, err := goexiv.Open("testdata/pixel.jpg")
	require.NoError(t, err)

	err = img.SetIptcString("Iptc.Application2.Caption", "123")
	require.NoError(t, err)

	err = img.IptcStripKey("Iptc.Application2.Caption")
	require.NoError(t, err)

	err = img.ReadMetadata()
	require.NoError(t, err)

	data := img.GetIptcData()

	_, err = data.GetString("Iptc.Application2.Caption")
	require.Error(t, err)
}

func TestXmpStripKey(t *testing.T) {
	t.Skip("XMP SetXmpString and GetString is not implemented yet")
	//img, err := goexiv.Open("testdata/pixel.jpg")
	//require.NoError(t, err)
	//
	//err = img.SetXmpString("Xmp.dc.description", "123")
	//require.NoError(t, err)
	//
	//err = img.XmpStripKey("Xmp.dc.description")
	//require.NoError(t, err)
	//
	//err = img.ReadMetadata()
	//require.NoError(t, err)
	//
	//data := img.GetXmpData()
	//
	//_, err = data.GetString("Xmp.dc.description")
	//require.Error(t, err)
}

func BenchmarkImage_GetBytes_KeepAlive(b *testing.B) {
	bytes, err := ioutil.ReadFile("testdata/stripped_pixel.jpg")
	require.NoError(b, err)
	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			img, err := goexiv.OpenBytes(bytes)
			require.NoError(b, err)

			runtime.GC()

			require.NoError(b, img.SetExifString("Exif.Photo.UserComment", "123"))

			bytesAfter := img.GetBytes()
			assert.NotEmpty(b, bytesAfter)
			runtime.KeepAlive(img)
		}()
	}

	wg.Wait()
}

func BenchmarkImage_GetBytes_NoKeepAlive(b *testing.B) {
	bytes, err := ioutil.ReadFile("testdata/stripped_pixel.jpg")
	require.NoError(b, err)
	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			img, err := goexiv.OpenBytes(bytes)
			require.NoError(b, err)

			require.NoError(b, img.SetExifString("Exif.Photo.UserComment", "123"))

			bytesAfter := img.GetBytes()
			assert.NotEmpty(b, bytesAfter)
		}()
	}
}

// Fills the image with metadata
func initializeImage(path string, t *testing.T) {
	img, err := goexiv.Open(path)
	require.NoError(t, err)

	img.SetIptcString("Iptc.Application2.Copyright", "this is the copy, right?")
	img.SetIptcString("Iptc.Application2.CountryName", "Lancre")
	img.SetIptcString("Iptc.Application2.DateCreated", "20121013")
	img.SetIptcString("Iptc.Application2.TimeCreated", "124932:0100")

	exifTags := map[string]string{
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
	}

	for k, v := range exifTags {
		err = img.SetExifString(k, v)
		require.NoError(t, err, k, v)
	}
}
