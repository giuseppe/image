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
