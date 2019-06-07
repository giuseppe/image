// +build !cgo

package compression

import (
	"fmt"
	"io"
)

// ZstdDecompressor is a DecompressorFunc for the zstd compression algorithm.
func ZstdDecompressor(r io.Reader) (io.ReadCloser, error) {
	return nil, fmt.Errorf("zstd not supported on this platform")
}
