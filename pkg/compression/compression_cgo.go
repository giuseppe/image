// +build cgo

package compression

import (
	"io"

	"github.com/DataDog/zstd"
)

// ZstdDecompressor is a DecompressorFunc for the zstd compression algorithm.
func ZstdDecompressor(r io.Reader) (io.ReadCloser, error) {
	return zstd.NewReader(r), nil
}

// zstdCompressor is a CompressorFunc for the zstd compression algorithm.
func zstdCompressor(r io.Writer, level *int) (io.WriteCloser, error) {
	if level == nil {
		return zstd.NewWriter(r), nil
	}
	return zstd.NewWriterLevel(r, *level), nil
}
