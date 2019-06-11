// +build !cgo

package compression

import (
	"fmt"
	"io"
)

// ZstdDecompressor is a DecompressorFunc for the zstd compression algorithm.
func ZstdDecompressor(r io.Reader) (io.ReadCloser, error) {
	return nil, fmt.Errorf("zstd not supported")
}

// zstdCompressor is a CompressorFunc for the zstd compression algorithm.
func zstdCompressor(r io.Writer, level *int) (io.WriteCloser, error) {
	return nil, fmt.Errorf("zstd not supported")
}
