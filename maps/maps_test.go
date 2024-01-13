package maps

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"math/bits"
	"math/rand"
	"testing"
)

func createMap(t *testing.T, size uint16) {
	m := New(size)

	requireBuckets := size>>3 + 1
	log2 := uint8(bits.Len16(requireBuckets))
	bucketNum := 1 << log2

	assert.Equal(t, bucketNum, len(m.Buckets), "buckets num")
	assert.Equal(t, log2, m.LogBucketsCount, "log2 buckets count")
}

func TestNew(t *testing.T) {
	var size = uint16(rand.Intn(math.MaxInt8))
	createMap(t, size)
}

// go test ./maps -fuzz=Fuzz -fuzztime=10s
func FuzzNew(f *testing.F) {
	testcases := []uint16{0, 1 << 5}
	for _, tc := range testcases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}
	f.Fuzz(createMap)
}

func TestNew_BucketNum(t *testing.T) {
	var size uint16
	for ; size < math.MaxUint8; size++ {
		m := New(size)
		log2 := uint8(bits.Len16(size>>3 + 1))
		require.Equal(t, log2, m.LogBucketsCount, size)
	}

}

func TestMap_Add(t *testing.T) {
	m := New(1)
	m.Add("a", "b")
	assert.Equal(t, 1, m.Length)
	assert.Equal(t, uint8(1), m.LogBucketsCount)
}

func TestMap_Get(t *testing.T) {
	m := New(1)
	m.Add("a", "b")
	res, ok := m.Get("a")

	assert.True(t, ok)
	assert.Equal(t, "b", res)
}
