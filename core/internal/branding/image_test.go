package branding

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// U-BRD image (Plan r2 §1.4.1, §6.2): the PNG chunk walk, the JPEG marker walk and the normalization (family,
// dimensions, size cap, ancillary data dropped) on small synthetic files, plus the standard library's own progressive,
// CMYK and restart-marker JPEG test images.

func gradient(w, h int) *image.NRGBA {
	m := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			m.Set(x, y, color.NRGBA{uint8(x * 255 / w), uint8(y * 255 / h), 128, 255})
		}
	}
	return m
}

func encPNG(t *testing.T, m image.Image) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, m); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func encJPEG(t *testing.T, m image.Image, q int) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := jpeg.Encode(&b, m, &jpeg.Options{Quality: q}); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func chunk(typ string, data []byte) []byte {
	out := binary.BigEndian.AppendUint32(nil, uint32(len(data)))
	out = append(out, typ...)
	out = append(out, data...)
	return binary.BigEndian.AppendUint32(out, crc32.ChecksumIEEE(append([]byte(typ), data...)))
}

// afterIHDR inserts c after the IHDR chunk (signature 8 + IHDR 25 bytes); beforeIEND before the final 12-byte IEND.
func afterIHDR(p, c []byte) []byte {
	return append(append(append([]byte{}, p[:33]...), c...), p[33:]...)
}
func beforeIEND(p, c []byte) []byte {
	return append(append(append([]byte{}, p[:len(p)-12]...), c...), p[len(p)-12:]...)
}

// withAPP1 inserts an APP1 (EXIF) segment holding a complete embedded JPEG thumbnail (with its own SOI/EOI) after SOI.
func withAPP1(t *testing.T, j []byte) []byte {
	thumb := encJPEG(t, gradient(8, 8), 50)
	payload := append([]byte("Exif\x00\x00"), thumb...)
	seg := append([]byte{0xFF, 0xE1}, binary.BigEndian.AppendUint16(nil, uint16(len(payload)+2))...)
	seg = append(seg, payload...)
	return append(append(append([]byte{}, j[:2]...), seg...), j[2:]...)
}

// interlacedGray hand-builds an Adam7-interlaced 8-bit grayscale PNG (the Go encoder writes no interlaced PNG).
func interlacedGray(w, h int) []byte {
	passes := [][4]int{{0, 0, 8, 8}, {4, 0, 8, 8}, {0, 4, 4, 8}, {2, 0, 4, 4}, {0, 2, 2, 4}, {1, 0, 2, 2}, {0, 1, 1, 2}}
	var raw []byte
	for _, p := range passes {
		for y := p[1]; y < h; y += p[3] {
			if p[0] >= w {
				break
			}
			raw = append(raw, 0) // filter: none
			for x := p[0]; x < w; x += p[2] {
				raw = append(raw, uint8((x*7+y*13)%256))
			}
		}
	}
	var z bytes.Buffer
	zw := zlib.NewWriter(&z)
	_, _ = zw.Write(raw)
	_ = zw.Close()
	ihdr := binary.BigEndian.AppendUint32(nil, uint32(w))
	ihdr = binary.BigEndian.AppendUint32(ihdr, uint32(h))
	ihdr = append(ihdr, 8, 0, 0, 0, 1) // depth 8, grayscale, deflate, adaptive filtering, Adam7
	out := []byte(pngSignature)
	out = append(out, chunk("IHDR", ihdr)...)
	out = append(out, chunk("IDAT", z.Bytes())...)
	return append(out, chunk("IEND", nil)...)
}

// noisePalette is a palette PNG of random indexes into 256 random colours: about one byte per pixel as input, but
// three to four bytes per pixel once normalized to 8-bit RGBA, so its normalized output exceeds 512 KiB.
func noisePalette(t *testing.T, w, h int) []byte {
	r := rand.New(rand.NewSource(17))
	pal := make(color.Palette, 256)
	for i := range pal {
		pal[i] = color.NRGBA{uint8(r.Intn(256)), uint8(r.Intn(256)), uint8(r.Intn(256)), 255}
	}
	m := image.NewPaletted(image.Rect(0, 0, w, h), pal)
	for i := range m.Pix {
		m.Pix[i] = uint8(r.Intn(256))
	}
	return encPNG(t, m)
}

func stdlibImage(t *testing.T, name string) []byte {
	t.Helper()
	root := os.Getenv("GOROOT")
	if root == "" {
		out, err := exec.Command("go", "env", "GOROOT").Output() // the test binary may be built with -trimpath
		if err != nil {
			t.Fatalf("go env GOROOT: %v", err)
		}
		root = strings.TrimSpace(string(out))
	}
	b, err := os.ReadFile(filepath.Join(root, "src", "image", "testdata", name))
	if err != nil {
		t.Fatalf("standard library test image %s: %v", name, err)
	}
	return b
}

func TestNormalizeLogoAccepts(t *testing.T) {
	pngText := beforeIEND(encPNG(t, gradient(40, 24)), chunk("tEXt", []byte("Comment\x00hello")))
	n16 := image.NewNRGBA64(image.Rect(0, 0, 20, 18))
	for i := range n16.Pix {
		n16.Pix[i] = uint8(i * 31)
	}
	pal := image.NewPaletted(image.Rect(0, 0, 32, 16), color.Palette{color.Black, color.White, color.NRGBA{200, 10, 10, 128}})
	for i := range pal.Pix {
		pal.Pix[i] = uint8(i % 3)
	}
	cases := []struct {
		name, ct string
		in       []byte
		w, h     int
	}{
		{"png with tEXt", TypePNG, pngText, 40, 24},
		{"16-bit png", TypePNG, encPNG(t, n16), 20, 18},
		{"palette png with alpha", TypePNG, encPNG(t, pal), 32, 16},
		{"interlaced png", TypePNG, interlacedGray(19, 17), 19, 17},
		{"jpeg with APP1 thumbnail", TypeJPEG, withAPP1(t, encJPEG(t, gradient(64, 32), 95)), 64, 32},
		{"progressive jpeg", TypeJPEG, stdlibImage(t, "video-001.progressive.jpeg"), 150, 103},
		{"cmyk jpeg", TypeJPEG, stdlibImage(t, "video-001.cmyk.jpeg"), 150, 103},
		{"restart-marker jpeg", TypeJPEG, stdlibImage(t, "video-001.restart2.jpeg"), 150, 103},
		{"minimum 16 px", TypePNG, encPNG(t, image.NewGray(image.Rect(0, 0, 16, 16))), 16, 16},
		{"maximum 2048 px", TypePNG, encPNG(t, image.NewGray(image.Rect(0, 0, 2048, 16))), 2048, 16},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			l, err := NormalizeLogo(c.in)
			if err != nil {
				t.Fatalf("refused: %v", err)
			}
			if l.ContentType != c.ct || l.Width != c.w || l.Height != c.h || len(l.Data) > MaxLogoBytes {
				t.Fatalf("got %s %dx%d (%d bytes)", l.ContentType, l.Width, l.Height, len(l.Data))
			}
			if bytes.Equal(l.Data, c.in) {
				t.Fatal("stored bytes equal the input: not normalized")
			}
			m, format, err := image.Decode(bytes.NewReader(l.Data))
			if err != nil || "image/"+format != c.ct || m.Bounds().Dx() != c.w || m.Bounds().Dy() != c.h {
				t.Fatalf("normalized output does not decode to the same family and size: %v %s", err, format)
			}
			if c.ct == TypePNG {
				switch m.(type) {
				case *image.NRGBA, *image.RGBA: // 8-bit; the encoder writes an opaque canvas as 8-bit RGB
				default:
					t.Fatalf("normalized PNG decodes as %T, want 8-bit RGB(A)", m)
				}
				if bytes.Contains(l.Data, []byte("tEXt")) {
					t.Fatal("ancillary text chunk kept")
				}
				if walkPNG(l.Data) != nil {
					t.Fatal("normalized PNG fails the chunk walk")
				}
			} else {
				if bytes.Contains(l.Data, []byte("Exif")) || walkJPEG(l.Data) != nil {
					t.Fatal("normalized JPEG keeps EXIF or fails the marker walk")
				}
			}
			again, err := NormalizeLogo(l.Data)
			if err != nil || again.ContentType != c.ct {
				t.Fatalf("normalized output is not itself accepted: %v", err)
			}
		})
	}
}

func TestNormalizeLogoRefuses(t *testing.T) {
	p := encPNG(t, gradient(32, 32))
	j := encJPEG(t, gradient(32, 32), 90)
	prog := stdlibImage(t, "video-001.progressive.jpeg")
	badCRC := append([]byte{}, p...)
	badCRC[len(badCRC)-13] ^= 0xFF // the last byte of the final IDAT's CRC
	var g bytes.Buffer
	if err := gif.Encode(&g, gradient(32, 32), nil); err != nil {
		t.Fatal(err)
	}
	webp := append([]byte("RIFF\x24\x00\x00\x00WEBPVP8 "), make([]byte, 40)...)
	acTL := append(binary.BigEndian.AppendUint32(nil, 1), 0, 0, 0, 0)
	big := append(append([]byte{}, p...), make([]byte, MaxLogoBytes+1-len(p))...)
	cases := []struct {
		name, field string
		in          []byte
	}{
		{"empty", FieldLogoType, nil},
		{"header-only png", FieldLogoMalformed, p[:33]},
		{"truncated png", FieldLogoMalformed, p[:len(p)/2]},
		{"header-only jpeg", FieldLogoMalformed, j[:200]},
		{"truncated jpeg", FieldLogoMalformed, j[:len(j)-40]},
		{"truncated progressive jpeg", FieldLogoMalformed, prog[:len(prog)-100]},
		{"bytes after IEND", FieldLogoMalformed, append(append([]byte{}, p...), "<html>"...)},
		{"bytes after EOI", FieldLogoMalformed, append(append([]byte{}, j...), "<html>"...)},
		{"APP1 thumbnail then bytes after EOI", FieldLogoMalformed, append(withAPP1(t, j), 0xFF, 0xD9, 'x')},
		{"animated png (acTL)", FieldLogoMalformed, afterIHDR(p, chunk("acTL", acTL))},
		{"private chunk", FieldLogoMalformed, beforeIEND(p, chunk("prVt", []byte("x")))},
		{"unknown public chunk", FieldLogoMalformed, beforeIEND(p, chunk("eXIf", []byte("MM")))},
		{"bad crc", FieldLogoMalformed, badCRC},
		{"svg", FieldLogoType, []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32"><script>alert(1)</script></svg>`)},
		{"html", FieldLogoType, []byte("<!DOCTYPE html><html><body>logo</body></html>")},
		{"gif", FieldLogoType, g.Bytes()},
		{"webp", FieldLogoType, webp},
		{"15 px", FieldLogoDimensions, encPNG(t, gradient(15, 32))},
		{"2049 px", FieldLogoDimensions, encPNG(t, image.NewGray(image.Rect(0, 0, 2049, 16)))},
		{"524289 bytes", FieldLogoTooLarge, big},
		{"normalized too large", FieldNormalizedTooLarge, noisePalette(t, 700, 700)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.field == FieldNormalizedTooLarge && len(c.in) > MaxLogoBytes {
				t.Fatalf("fixture input is %d bytes, over the input bound", len(c.in))
			}
			_, err := NormalizeLogo(c.in)
			e, ok := AsError(err)
			if !ok || e.Code != CodeInvalidInput || e.Field != c.field {
				t.Fatalf("got %v, want invalid_input %s", err, c.field)
			}
		})
	}
}

func TestStructuralWalks(t *testing.T) {
	p := encPNG(t, gradient(20, 20))
	if walkPNG(p) != nil {
		t.Fatal("a plain PNG fails the walk")
	}
	if walkPNG(afterIHDR(p, chunk("IHDR", make([]byte, 13)))) == nil {
		t.Error("a second IHDR passes")
	}
	if walkPNG(append(append([]byte{}, p[:8]...), p[33:]...)) == nil {
		t.Error("a PNG without IHDR first passes")
	}
	j := encJPEG(t, gradient(20, 20), 80)
	if walkJPEG(j) != nil {
		t.Fatal("a plain JPEG fails the walk")
	}
	if walkJPEG(withAPP1(t, j)) != nil {
		t.Error("an APP1 thumbnail's embedded EOI ended the walk early")
	}
	if walkJPEG(append(append([]byte{}, j[:2]...), 0xFF, 0xE0, 0xFF, 0xFF)) == nil {
		t.Error("a segment length past the buffer passes")
	}
}
