package branding

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"net/http"
)

// Logo image validation and normalization (G017 Plan r2 §1.4.1). The caller reads the request body once into a buffer
// of at most MaxLogoBytes+1 bytes; NormalizeLogo then sniffs the family, bounds the header dimensions, walks the whole
// buffer structurally (PNG chunks with CRCs and an ancillary allow-list; JPEG markers and entropy-coded scans), decodes
// it completely and re-encodes it with the standard library to a normalized 8-bit output whose size is bounded again.
// The caller's bytes never reach the store: only the encoder's output does, with the server-detected type.

// The logo bounds.
const (
	MaxLogoBytes  = 524288 // input and normalized output, in bytes
	MinLogoPixels = 16     // each dimension, inclusive
	MaxLogoPixels = 2048   // each dimension, inclusive
	jpegQuality   = 90
)

// The logo content types (the only families accepted).
const (
	TypePNG  = "image/png"
	TypeJPEG = "image/jpeg"
)

// The closed logo refusal fields (Error.Field with CodeInvalidInput).
const (
	FieldLogoTooLarge       = "logo_too_large"       // the input exceeds MaxLogoBytes
	FieldLogoType           = "logo_type"            // not a PNG or JPEG by content
	FieldLogoDimensions     = "logo_dimensions"      // a dimension outside 16..2048
	FieldLogoMalformed      = "logo_malformed"       // a structural or decode failure (truncated, appended, animated, unknown chunk, bad CRC)
	FieldNormalizedTooLarge = "normalized_too_large" // the re-encoded output exceeds MaxLogoBytes
)

const (
	pngSignature        = "\x89PNG\r\n\x1a\n"
	maxPNGChunk  uint32 = 1<<31 - 1
)

// Logo is a normalized logo: the encoder's bytes, the server-detected content type and the preserved dimensions.
type Logo struct {
	ContentType   string
	Data          []byte
	Width, Height int
}

// pngChunks is the PNG chunk allow-list: the critical chunks and the listed ancillary ones. Animation chunks (acTL,
// fcTL, fdAT) and every other type are refused.
var pngChunks = map[string]bool{
	"IHDR": true, "PLTE": true, "IDAT": true, "IEND": true, "tRNS": true, "gAMA": true, "cHRM": true, "sRGB": true,
	"iCCP": true, "pHYs": true, "tEXt": true, "zTXt": true, "iTXt": true, "bKGD": true, "sBIT": true, "hIST": true,
	"tIME": true,
}

var errMalformed = errors.New("branding: malformed image")

// NormalizeLogo validates in (at most MaxLogoBytes) and returns its normalized re-encoding, or a closed invalid_input
// error naming the refusal.
func NormalizeLogo(in []byte) (Logo, error) {
	if len(in) > MaxLogoBytes {
		return Logo{}, invalid(FieldLogoTooLarge)
	}
	if len(in) == 0 {
		return Logo{}, invalid(FieldLogoType)
	}
	head := in
	if len(head) > 512 {
		head = head[:512]
	}
	ct := http.DetectContentType(head)
	var format string
	switch ct {
	case TypePNG:
		format = "png"
	case TypeJPEG:
		format = "jpeg"
	default:
		return Logo{}, invalid(FieldLogoType)
	}
	cfg, got, err := image.DecodeConfig(bytes.NewReader(in))
	switch {
	case err != nil:
		return Logo{}, invalid(FieldLogoMalformed)
	case got != format:
		return Logo{}, invalid(FieldLogoType)
	case cfg.Width < MinLogoPixels || cfg.Height < MinLogoPixels || cfg.Width > MaxLogoPixels || cfg.Height > MaxLogoPixels:
		return Logo{}, invalid(FieldLogoDimensions)
	}
	walk := walkPNG
	if format == "jpeg" {
		walk = walkJPEG
	}
	if walk(in) != nil {
		return Logo{}, invalid(FieldLogoMalformed)
	}
	img, got, err := image.Decode(bytes.NewReader(in))
	if err != nil || got != format {
		return Logo{}, invalid(FieldLogoMalformed)
	}
	b := img.Bounds()
	if b.Dx() != cfg.Width || b.Dy() != cfg.Height {
		return Logo{}, invalid(FieldLogoMalformed)
	}
	canvas := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(canvas, canvas.Bounds(), img, b.Min, draw.Src)
	out := &boundedBuffer{limit: MaxLogoBytes}
	if format == "png" {
		err = (&png.Encoder{CompressionLevel: png.DefaultCompression}).Encode(out, canvas)
	} else {
		err = jpeg.Encode(out, canvas, &jpeg.Options{Quality: jpegQuality})
	}
	if errors.Is(err, errTooLarge) {
		return Logo{}, invalid(FieldNormalizedTooLarge)
	}
	if err != nil {
		return Logo{}, invalid(FieldLogoMalformed)
	}
	return Logo{ContentType: ct, Data: out.buf.Bytes(), Width: b.Dx(), Height: b.Dy()}, nil
}

var errTooLarge = errors.New("branding: normalized output too large")

// boundedBuffer is the encoder's writer: it refuses any write that would exceed limit.
type boundedBuffer struct {
	buf   bytes.Buffer
	limit int
}

func (w *boundedBuffer) Write(p []byte) (int, error) {
	if w.buf.Len()+len(p) > w.limit {
		return 0, errTooLarge
	}
	return w.buf.Write(p)
}

// walkPNG checks the whole buffer: the signature, then chunks whose length fits the remaining bytes, whose CRC matches
// and whose type is allow-listed, IHDR exactly once and first, IEND exactly once and ending the buffer.
func walkPNG(in []byte) error {
	if len(in) < len(pngSignature) || string(in[:len(pngSignature)]) != pngSignature {
		return errMalformed
	}
	pos, first := len(pngSignature), true
	for {
		if len(in)-pos < 12 {
			return errMalformed // a chunk header and CRC no longer fit: truncated or no IEND
		}
		n := binary.BigEndian.Uint32(in[pos:])
		if n > maxPNGChunk || uint64(n) > uint64(len(in)-pos-12) {
			return errMalformed
		}
		typ := string(in[pos+4 : pos+8])
		data := in[pos+4 : pos+8+int(n)] // type and data, the CRC's input
		if crc32.ChecksumIEEE(data) != binary.BigEndian.Uint32(in[pos+8+int(n):]) {
			return errMalformed
		}
		if !pngChunks[typ] || (typ == "IHDR") != first {
			return errMalformed
		}
		first = false
		pos += 12 + int(n)
		if typ == "IEND" {
			if n != 0 || pos != len(in) {
				return errMalformed // IEND carries no data and must end the buffer (no appended bytes)
			}
			return nil
		}
	}
}

// walkJPEG checks the whole buffer: SOI, then marker segments skipped by their 16-bit length (APPn, COM, DQT, DHT,
// SOFn, DRI, ...), after each SOS the entropy-coded data up to the next marker other than FF00 stuffing or RSTn, and
// EOI ending the buffer. An APP1 thumbnail is skipped by its segment length, so its embedded EOI cannot end the walk.
func walkJPEG(in []byte) error {
	if len(in) < 4 || in[0] != 0xFF || in[1] != 0xD8 {
		return errMalformed
	}
	pos := 2
	for {
		if pos >= len(in) || in[pos] != 0xFF {
			return errMalformed
		}
		for pos < len(in) && in[pos] == 0xFF {
			pos++ // fill bytes before a marker
		}
		if pos >= len(in) {
			return errMalformed
		}
		m := in[pos]
		pos++
		switch {
		case m == 0xD9: // EOI
			if pos != len(in) {
				return errMalformed // appended bytes after EOI
			}
			return nil
		case m == 0x00 || m == 0xD8:
			return errMalformed // stuffing outside a scan, or a second SOI
		case m == 0x01 || (m >= 0xD0 && m <= 0xD7):
			continue // TEM and RSTn carry no length
		}
		if len(in)-pos < 2 {
			return errMalformed
		}
		n := int(binary.BigEndian.Uint16(in[pos:]))
		if n < 2 || n > len(in)-pos {
			return errMalformed
		}
		pos += n
		if m != 0xDA { // not SOS
			continue
		}
		// Entropy-coded data: runs until an FF followed by a byte other than 00 (stuffing), D0-D7 (RSTn) or FF (fill).
		for {
			if pos >= len(in)-1 {
				return errMalformed // no marker after the scan: truncated, no EOI
			}
			if in[pos] != 0xFF {
				pos++
				continue
			}
			nx := in[pos+1]
			if nx == 0x00 || (nx >= 0xD0 && nx <= 0xD7) {
				pos += 2
				continue
			}
			if nx == 0xFF {
				pos++
				continue
			}
			break // a marker at pos
		}
	}
}
