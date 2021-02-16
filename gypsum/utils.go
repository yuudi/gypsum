package gypsum

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
)

var hotSalt []byte  // refresh at every start
var coldSalt []byte //

func init() {
	seed := make([]byte, 8)
	_, _ = cryptoRand.Read(seed)
	rand.Seed(int64(binary.LittleEndian.Uint64(seed)))

	hotSalt = make([]byte, 8)
	rand.Read(hotSalt)
}

func U64ToBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	return b
}

func ToUint(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}

func AnyToFloat(unk interface{}) (float64, error) {
	switch i := unk.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int16:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint16:
		return float64(i), nil
	case uint:
		return float64(i), nil
	case string:
		return strconv.ParseFloat(i, 64)
	default:
		return 0, errors.New(fmt.Sprintf("error: cannot accept %#v as float", unk))
	}
}

func AnyToInt64(unk interface{}) (int64, error) {
	switch i := unk.(type) {
	case int64:
		return int64(i), nil
	case int32:
		return int64(i), nil
	case float64:
		return int64(i), nil
	case float32:
		return int64(i), nil
	case int16:
		return int64(i), nil
	case int:
		return int64(i), nil
	case uint64:
		return int64(i), nil
	case uint32:
		return int64(i), nil
	case uint16:
		return int64(i), nil
	case uint:
		return int64(i), nil
	case string:
		return strconv.ParseInt(i, 10, 64)
	default:
		return 0, errors.New(fmt.Sprintf("error: cannot accept %#v as int", unk))
	}
}

var nonAsciiPattern = regexp.MustCompile("[^\\w\\-]")

func ReplaceFilename(s, r string) string {
	return nonAsciiPattern.ReplaceAllLiteralString(s, r)
}
