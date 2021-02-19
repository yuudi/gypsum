package gypsum

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	"errors"
	"io/fs"
	"math/rand"
	"os"
	"path"
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
