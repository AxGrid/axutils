package crypto_fs

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/axgrid/axutils/crypto"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"
)

//var f embed.FS

type CryptoFSBuilder struct {
	savePath  string
	rootDir   string
	files     map[string]file
	aesSecret []byte
}

func NewCryptoFSBuilder() *CryptoFSBuilder {
	return &CryptoFSBuilder{
		files: make(map[string]file),
	}
}

func (b *CryptoFSBuilder) AddFile(path string) *CryptoFSBuilder {
	of, err := os.Open(path)
	if err != nil {
		return b
	}
	defer of.Close()
	data, err := os.ReadFile(path)
	if err != nil {
		return b
	}
	f := file{
		FileName: of.Name(),
		Path:     path,
		Data:     data,
	}
	b.files[path] = f
	return b
}

func (b *CryptoFSBuilder) AddFolder(folderPath string) *CryptoFSBuilder {
	folderPath = path.Join(b.rootDir, folderPath)
	err := filepath.Walk(folderPath, func(root string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			entries, err := os.ReadDir(root)
			if err != nil {
				return err
			}
			var files []file
			for _, e := range entries {
				files = append(files, file{
					FileName:  e.Name(),
					Path:      filepath.Join(root, e.Name()),
					IsDirFlag: e.IsDir(),
				})
			}
			b.files[root] = file{
				FileName:  info.Name(),
				Path:      root,
				IsDirFlag: info.IsDir(),
				Entries:   files,
			}
			return nil
		}
		content, err := os.ReadFile(root)
		if err != nil {
			return err
		}
		b.files[root] = file{
			FileName:  info.Name(),
			Path:      root,
			IsDirFlag: info.IsDir(),
			Data:      content,
		}

		return nil
	})
	if err != nil {
		return b
	}
	return b
}

func (b *CryptoFSBuilder) WithSavePath(path string) *CryptoFSBuilder {
	b.savePath = path
	return b
}

func (b *CryptoFSBuilder) WithRootDir(rootDir string) *CryptoFSBuilder {
	b.rootDir = rootDir
	return b
}

func (b *CryptoFSBuilder) WithAES(secret []byte) *CryptoFSBuilder {
	b.aesSecret = secret
	return b
}

func (b *CryptoFSBuilder) Save() error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(b.files)
	if err != nil {
		return err
	}
	data := buf.Bytes()
	if b.aesSecret != nil {
		aes := crypto.NewAES(b.aesSecret)
		data, err = aes.Encrypt(buf.Bytes())
		if err != nil {
			return err
		}
	}
	target, err := os.Create(b.savePath)
	if err != nil {
		return err
	}
	defer target.Close()
	_, err = target.Write(data)
	if err != nil {
		return err
	}
	return nil
}

type CryptoFS interface {
	Open(name string) (fs.File, error)
	ReadDir(name string) ([]fs.DirEntry, error)
}

type cryptoFS struct {
	files map[string]file
}

func CryptoFSFromFile(path string, aesSecret []byte) (CryptoFS, error) {
	cfs := &cryptoFS{}
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if aesSecret != nil {
		aes := crypto.NewAES(aesSecret)
		decrypted, err := aes.Decrypt(f)
		if err != nil {
			return nil, err
		}
		buf.Write(decrypted)
	} else {
		buf.Write(f)
	}
	dec := gob.NewDecoder(&buf)
	err = dec.Decode(&cfs.files)
	if err != nil {
		return nil, err
	}
	return cfs, nil
}

func (c *cryptoFS) Open(path string) (fs.File, error) {
	f, ok := c.files[path]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return f, nil
}

func (c *cryptoFS) ReadDir(name string) ([]fs.DirEntry, error) {
	f, ok := c.files[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	if !f.IsDir() {
		return nil, errors.New("not a directory")
	}
	entries := make([]fs.DirEntry, len(f.Entries))
	for i := 0; i < len(f.Entries); i++ {
		entries[i] = f.Entries[i]
	}
	return entries, nil
}

type file struct {
	FileName  string
	Path      string
	Data      []byte
	IsDirFlag bool
	Entries   []file
}

func (f file) Type() fs.FileMode {
	return fs.FileMode(0755)
}

func (f file) Info() (fs.FileInfo, error) {
	return f, nil
}

func (f file) Name() string {
	return f.FileName
}

func (f file) Size() int64 {
	return int64(len(f.Data))
}

func (f file) Mode() fs.FileMode {
	return fs.FileMode(0)
}

func (f file) ModTime() time.Time {
	return time.Time{}
}

func (f file) IsDir() bool {
	return f.IsDirFlag
}

func (f file) Sys() any {
	return nil
}

func (f file) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (f file) Read(i []byte) (int, error) {
	buf := bytes.NewBuffer(f.Data)
	return buf.Read(i)
}

func (f file) Close() error { return nil }
