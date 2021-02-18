package gypsum

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path"
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

func ExtractWebAssets(extractPath string) error {
	s, err := os.Stat(extractPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(extractPath, 0644); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if s.IsDir() {
			// ok
		} else {
			return errors.New(extractPath + " is not a directory")
		}
	}
	return fs.WalkDir(publicAssets, "web", func(filePath string, d fs.DirEntry, err error) error {
		println("extracting: ", filePath)
		if err != nil {
			return err
		}
		if d.IsDir() {
			err = os.Mkdir(path.Join(extractPath, filePath), 0644)
			if err != nil {
				return err
			}
		} else {
			data, err := publicAssets.ReadFile(filePath)
			if err != nil {
				return err
			}
			err = os.WriteFile(path.Join(extractPath, filePath), data, 644)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
