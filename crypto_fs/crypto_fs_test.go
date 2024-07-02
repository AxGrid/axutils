package crypto_fs

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"path"
	"path/filepath"
	"testing"
)

var secret = []byte("super_puper_secret_super_puper_s")

// Тесты
// 1. Создание криптованного и не криптованного FS / отдельно файл, отдельно папка, файл + папка
func TestCryptoFSBuilder(t *testing.T) {
	err := NewCryptoFSBuilder().
		AddFile("/Users/lifeofhucci/axutils/crypto_fs/example.txt").
		WithAES(secret).
		WithSavePath("enc_fs.bin").
		Save()
	assert.Nil(t, err)

	err = NewCryptoFSBuilder().
		AddFile("/Users/lifeofhucci/axutils/crypto_fs/example.txt").
		WithSavePath("un_enc_fs.bin").
		Save()
	assert.Nil(t, err)

	err = NewCryptoFSBuilder().
		AddFolder("/Users/lifeofhucci/axutils/crypto_fs/example").
		WithSavePath("folder_fs.bin").
		Save()
	assert.Nil(t, err)

	err = NewCryptoFSBuilder().
		AddFolder("/Users/lifeofhucci/axutils/crypto_fs/example").
		AddFile("/Users/lifeofhucci/axutils/crypto_fs/example.txt").
		WithSavePath("folder_wFile_fs.bin").
		Save()
	assert.Nil(t, err)

}

// 2. Открытие папки, открытие файла, чтение директории и чтение каждого файла
func TestCryptoFs_Read(t *testing.T) {
	cfs, err := CryptoFSFromFile("enc_fs.bin", secret)
	assert.Nil(t, err)

	exampleTxt, err := cfs.Open("/Users/lifeofhucci/axutils/crypto_fs/example.txt")
	assert.Nil(t, err)

	stat, err := exampleTxt.Stat()
	assert.Nil(t, err)

	b := make([]byte, stat.Size())
	_, err = exampleTxt.Read(b)
	assert.Nil(t, err)
	assert.Equal(t, "text", string(b))

	cfs, err = CryptoFSFromFile("folder_fs.bin", nil)
	assert.Nil(t, err)

	var shouldExists []string
	err = filepath.WalkDir("/Users/lifeofhucci/axutils/crypto_fs/example", func(path string, d fs.DirEntry, err error) error {
		shouldExists = append(shouldExists, path)
		return nil
	})
	assert.Nil(t, err)

	for _, pth := range shouldExists {
		p, err := cfs.Open(pth)
		assert.Nil(t, err)
		assert.NotNil(t, p)
	}

	cfs, err = CryptoFSFromFile("folder_wFile_fs.bin", nil)
	assert.Nil(t, err)
	shouldExists = append(shouldExists, "/Users/lifeofhucci/axutils/crypto_fs/example.txt")

	for _, pth := range shouldExists {
		p, err := cfs.Open(pth)
		assert.Nil(t, err)
		assert.NotNil(t, p)
	}

	dir, err := cfs.ReadDir("/Users/lifeofhucci/axutils/crypto_fs/example")
	assert.Nil(t, err)
	for _, entry := range dir {
		f, err := cfs.Open(path.Join("/Users/lifeofhucci/axutils/crypto_fs/example", entry.Name()))
		assert.Nil(t, err)
		assert.NotNil(t, f)
	}
}

// 3. Walk

func TestCryptoFS_Walk(t *testing.T) {
	cfs, err := CryptoFSFromFile("folder_wFile_fs.bin", nil)
	assert.Nil(t, err)

	checkMap := make(map[string]bool)
	err = filepath.WalkDir("/Users/lifeofhucci/axutils/crypto_fs/example", func(path string, d fs.DirEntry, err error) error {
		fmt.Println(path)
		checkMap[path] = false
		return nil
	})
	assert.Nil(t, err)

	err = fs.WalkDir(cfs, "/Users/lifeofhucci/axutils/crypto_fs/example", func(path string, d fs.DirEntry, err error) error {
		checkMap[path] = true
		return nil
	})
	assert.Nil(t, err)

	for _, v := range checkMap {
		assert.True(t, v)
	}
}

func TestCryptoFS_Relative(t *testing.T) {
	err := NewCryptoFSBuilder().
		AddFolder("example").
		WithAES(secret).
		WithSavePath("enc_fs.bin").
		WithRootDir("./").
		Save()
	assert.Nil(t, err)

	cfs, err := CryptoFSFromFile("enc_fs.bin", secret)
	assert.Nil(t, err)
	fmt.Println(cfs)

	err = fs.WalkDir(cfs, "example", func(path string, d fs.DirEntry, err error) error {
		fmt.Println(path)
		return nil
	})

}
