package compression

import (
	"bytes"
	"compress/bzip2"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/klauspost/pgzip"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/ulikunitz/xz"
)

// DecompressorFunc returns the decompressed stream, given a compressed stream.
// The caller must call Close() on the decompressed stream (even if the compressed input stream does not need closing!).
type DecompressorFunc func(io.Reader) (io.ReadCloser, error)

// GzipDecompressor is a DecompressorFunc for the gzip compression algorithm.
func GzipDecompressor(r io.Reader) (io.ReadCloser, error) {
	return pgzip.NewReader(r)
}

// Bzip2Decompressor is a DecompressorFunc for the bzip2 compression algorithm.
func Bzip2Decompressor(r io.Reader) (io.ReadCloser, error) {
	return ioutil.NopCloser(bzip2.NewReader(r)), nil
}

// XzDecompressor is a DecompressorFunc for the xz compression algorithm.
func XzDecompressor(r io.Reader) (io.ReadCloser, error) {
	r, err := xz.NewReader(r)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(r), nil
}

// compressorFunc writes the compressed stream to the given writer using the specified compression level.
// The caller must call Close() on the stream (even if the input stream does not need closing!).
type compressorFunc func(io.Writer, *int) (io.WriteCloser, error)

// gzipCompressor is a CompressorFunc for the gzip compression algorithm.
func gzipCompressor(r io.Writer, level *int) (io.WriteCloser, error) {
	if level != nil {
		return pgzip.NewWriterLevel(r, *level)
	}
	return pgzip.NewWriter(r), nil
}

// bzip2Compressor is a CompressorFunc for the bzip2 compression algorithm.
func bzip2Compressor(r io.Writer, level *int) (io.WriteCloser, error) {
	return nil, fmt.Errorf("bzip2 compression not supported")
}

// xzCompressor is a CompressorFunc for the xz compression algorithm.
func xzCompressor(r io.Writer, level *int) (io.WriteCloser, error) {
	return xz.NewWriter(r)
}

// compressionAlgos is an internal implementation detail of DetectCompression
var compressionAlgos = map[string]struct {
	prefix       []byte
	decompressor DecompressorFunc
}{
	"gzip":  {[]byte{0x1F, 0x8B, 0x08}, GzipDecompressor},                 // gzip (RFC 1952)
	"bzip2": {[]byte{0x42, 0x5A, 0x68}, Bzip2Decompressor},                // bzip2 (decompress.c:BZ2_decompress)
	"xz":    {[]byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}, XzDecompressor}, // xz (/usr/share/doc/xz/xz-file-format.txt)
	"zstd":  {[]byte{0x28, 0xb5, 0x2f, 0xfd}, ZstdDecompressor},           // zstd (http://www.zstd.net)
}

// compressors maps an algorithm to its compression function
var compressors = map[string]compressorFunc{
	"gzip":  gzipCompressor,
	"bzip2": bzip2Compressor,
	"xz":    xzCompressor,
	"zstd":  zstdCompressor,
}

// CompressStream returns the compressor by its name
func CompressStream(dest io.Writer, name string, level *int) (io.WriteCloser, error) {
	c, found := compressors[name]
	if !found {
		return nil, fmt.Errorf("cannot find compressor for '%s'", name)
	}
	return c(dest, level)

// DetectCompressionFormat returns a DecompressorFunc if the input is recognized as a compressed format, nil otherwise.
// Because it consumes the start of input, other consumers must use the returned io.Reader instead to also read from the beginning.
func DetectCompressionFormat(input io.Reader) (string, DecompressorFunc, io.Reader, error) {
	buffer := [8]byte{}

	n, err := io.ReadAtLeast(input, buffer[:], len(buffer))
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		// This is a “real” error. We could just ignore it this time, process the data we have, and hope that the source will report the same error again.
		// Instead, fail immediately with the original error cause instead of a possibly secondary/misleading error returned later.
		return "", nil, nil, err
	}

	name := ""
	var decompressor DecompressorFunc
	for algoname, algo := range compressionAlgos {
		if bytes.HasPrefix(buffer[:n], algo.prefix) {
			logrus.Debugf("Detected compression format %s", algoname)
			name = algoname
			decompressor = algo.decompressor
			break
		}
	}
	if decompressor == nil {
		logrus.Debugf("No compression detected")
	}

	return name, decompressor, io.MultiReader(bytes.NewReader(buffer[:n]), input), nil
}

// DetectCompression returns a DecompressorFunc if the input is recognized as a compressed format, nil otherwise.
// Because it consumes the start of input, other consumers must use the returned io.Reader instead to also read from the beginning.
func DetectCompression(input io.Reader) (DecompressorFunc, io.Reader, error) {
	_, d, r, e := DetectCompressionFormat(input)
	return d, r, e
}

// AutoDecompress takes a stream and returns an uncompressed version of the
// same stream.
// The caller must call Close() on the returned stream (even if the input does not need,
// or does not even support, closing!).
func AutoDecompress(stream io.Reader) (io.ReadCloser, bool, error) {
	decompressor, stream, err := DetectCompression(stream)
	if err != nil {
		return nil, false, errors.Wrapf(err, "Error detecting compression")
	}
	var res io.ReadCloser
	if decompressor != nil {
		res, err = decompressor(stream)
		if err != nil {
			return nil, false, errors.Wrapf(err, "Error initializing decompression")
		}
	} else {
		res = ioutil.NopCloser(stream)
	}
	return res, decompressor != nil, nil
}
