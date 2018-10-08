// Code generated by go-bindata. DO NOT EDIT.
// sources:
// templates/commands.tmpl
// templates/main.tmpl

package main


import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}


type asset struct {
	bytes []byte
	info  fileInfoEx
}

type fileInfoEx interface {
	os.FileInfo
	MD5Checksum() string
}

type bindataFileInfo struct {
	name        string
	size        int64
	mode        os.FileMode
	modTime     time.Time
	md5checksum string
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) MD5Checksum() string {
	return fi.md5checksum
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _bindataTemplatesCommandstmpl = []byte(
	"\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x56\x5b\x6f\xdb\x46\x13\x7d\x26\x7f\xc5\x7c\x44\xbe\x80\xac\x65\xaa" +
	"\x41\x8b\x3e\xa8\xd0\x43\x6a\xc7\x89\x81\xd8\x4e\x7d\xe9\x8b\x6b\xc0\x2b\x72\x44\x2d\xb2\xdc\xa5\x97\x2b\x37\x2e" +
	"\xc1\xff\x5e\xcc\xee\xf2\x22\x59\x6e\x2c\xa0\x7e\xb1\xb8\x33\x3b\xe7\xcc\x99\x0b\x39\x9d\xc2\x91\xca\x11\x0a\x94" +
	"\xa8\x99\xc1\x1c\x16\x4f\xa0\x2a\x94\xac\xe2\x87\x99\xe0\x87\xde\xa0\x74\x0a\xc7\x17\x70\x7e\x71\x0d\x1f\x8e\x4f" +
	"\xaf\xd3\x70\x3a\x85\x2b\x44\x58\x19\x53\xd5\xb3\xe9\xb4\xe0\x66\xb5\x5e\xa4\x99\x2a\xa7\x39\x93\x1c\x45\x61\xd8" +
	"\x93\x50\x7a\xba\x33\x56\x18\x56\x2c\xfb\xca\x0a\x84\x92\x71\x19\x86\xbc\xac\x94\x36\x10\x87\x41\xd3\x00\x5f\x42" +
	"\x7a\x6a\x0f\xea\xf4\xa4\x34\xd0\xb6\xd1\xb2\x34\x51\xd3\x00\xca\x1c\xda\xf6\x99\xd3\x95\xd1\x5c\x16\x35\x39\xd6" +
	"\xee\xe7\xc8\x39\x0c\xa2\xfd\xb8\x4d\x33\xc1\xa3\xcd\x5b\xba\x9e\xfe\x8d\x5a\x09\x55\x4c\x85\x2a\xb6\x8c\x75\xb5" +
	"\x7c\xf7\xd3\x34\x53\x0b\xcd\x76\x5a\x1e\x79\x85\x3a\x0a\x93\x30\x6c\x1a\x78\x23\x59\x89\x30\x9b\x43\x7a\x4e\x3f" +
	"\xda\xd6\x1e\xb2\x8a\xdb\xb3\x8f\xaa\x3b\x0d\x97\x6b\x99\x41\x67\x6b\xdb\x2b\xd4\x8f\xa8\xeb\x38\x81\xdb\xbb\x92" +
	"\x55\xb7\x2e\xcf\x3b\xf7\x0f\x9a\x30\xd0\x68\xd6\x5a\xee\xb2\x36\x61\x40\x82\x69\x26\x0b\x84\x37\xb5\x0d\x64\xd1" +
	"\x7c\x4c\xab\x68\x10\xec\xbc\x17\x04\x51\x8e\x75\xa6\x79\x65\xb8\x92\xd1\x0c\x48\x58\x1f\x23\x3d\x1e\x2c\x24\xfd" +
	"\xc4\xf9\xaf\xb5\xd8\xf2\xbb\xb9\xfc\xdc\xdb\xdb\x89\x63\xd3\x55\xb2\x0d\x9f\xe7\x7a\x89\x05\xaf\x0d\xea\xb8\x5e" +
	"\x2f\x32\x55\x96\x4c\xe6\xb0\x50\x4a\x24\x36\x4f\xa5\x0c\xb1\xcf\x04\x4f\x2f\x95\x32\x61\x18\xf0\x25\x8c\x3c\x89" +
	"\xb6\x75\x9a\xc3\x5b\x5b\x94\xf4\xc8\x59\x6c\x3e\x37\x35\x7a\x72\xd2\x29\xed\x68\x5d\xad\x94\x36\xce\x90\x5e\x73" +
	"\x23\x06\xcb\x67\x25\x8b\x99\x45\x3b\x63\xfa\x6b\xae\xfe\x92\xb1\xf5\xda\x4a\x3e\x21\x67\xca\x07\x50\xd4\x68\x49" +
	"\x74\x0c\x53\x1b\x1c\xe6\x5b\xd1\xc7\x1e\x04\x02\xf3\x57\xa0\x90\x60\xe1\xa8\x9a\xaa\xa2\x9e\x25\x2b\x15\xf4\xa2" +
	"\x7b\xf2\x35\xb5\xba\xc6\x56\x36\x12\xfd\xd0\xdf\x4a\x2f\x6c\x40\x26\xbe\x30\xcd\xca\xae\xfe\xc1\x23\xd3\x54\x84" +
	"\xa1\x09\xed\xd3\xf5\x53\x85\xde\x83\x22\xf4\x43\xe5\xfc\xf1\x1b\x2b\x2b\x81\x35\xb8\x96\x09\x9d\x5b\xc7\x0e\xbf" +
	"\x59\x5a\x1f\x3a\x27\x0f\xd4\x5f\x3a\x98\x43\x04\x10\xc1\x41\x5f\xcd\xae\x56\x5f\x98\x59\xc5\x09\x1c\x40\x64\xfb" +
	"\xa2\x4f\x33\xbd\xa9\x3b\x66\x14\xbd\x6d\xff\x94\x91\xc7\x1c\x31\xcb\xca\x9c\x80\x77\x94\x7f\xa8\xbf\x8f\xe4\xbb" +
	"\x96\x52\xa3\x8d\xf2\x5e\x70\x56\x0f\x4c\x03\xff\x3c\x83\xdb\x8d\xa1\xd8\x50\x73\xfb\x4e\x10\xd8\xf8\x43\xf0\x0d" +
	"\xe5\xec\x73\x3b\xa0\x8e\x8e\xc7\x4d\xe8\x7a\xa6\x0f\xf1\x52\x17\xda\xc6\xe9\xda\x2f\x08\xbc\xd2\xb3\xbe\x2e\xee" +
	"\xf8\xbd\x2e\xea\x19\x38\x35\xce\xb8\xe4\xe5\xba\x3c\xa7\xb3\xb8\x69\x40\xa0\x84\xf4\x12\x1f\xd6\x5c\x63\xde\x37" +
	"\x84\x8f\x77\xb9\x96\x33\xa0\x2e\x8a\x49\xd2\x1f\x36\xf4\x9c\x00\xd3\x45\xdd\x0b\xe3\xbb\x2c\x08\x86\x05\x63\x17" +
	"\x5f\xfa\x11\x8d\xdb\xcf\x71\xe4\x4c\xd4\xc5\xf4\x47\x63\xeb\x7c\xe7\x73\x88\xa2\xee\x7e\x17\x60\xbe\x6b\xfb\xdd" +
	"\xf6\x31\x4f\xa5\xe9\x02\x1e\x72\x99\xe3\xb7\x28\xb9\xbb\xb5\xcb\xe7\xce\x6b\x1c\xba\xff\x6b\x2d\x88\x8c\x73\x3d" +
	"\xb0\xaa\x51\x77\xd9\x11\x0c\x7d\x75\xba\x86\xe5\x13\x78\x53\x91\x06\xb6\x6f\x9f\xa9\xd2\x57\x93\x7a\x05\x1f\xbc" +
	"\x6f\x7a\x2a\x21\xaa\x98\x59\x45\xa3\x16\x20\xd4\xb9\x9f\x8a\x3a\xbd\xc4\x4a\xb0\x0c\xe3\xb5\x16\x13\xaa\xef\x7d" +
	"\x73\xdf\xb6\x94\x9e\x0b\xe0\xe7\xad\x69\xee\xdb\x7b\x2a\xb9\x55\xf6\x96\xec\x94\xfc\xdd\x04\xde\x25\x03\xf4\xb8" +
	"\x8f\x36\xdb\x3e\x08\x34\x3e\x74\xbb\xf1\x48\x70\x94\x26\xa5\x74\xcf\xd0\xac\x14\x79\xc5\x09\xed\x62\x62\x91\xfc" +
	"\x87\xa9\x3f\xac\x51\x3f\x8d\x73\x27\x16\x73\xd0\xf8\x90\xbe\xcf\xf3\xdf\xc9\xea\x9a\xf5\xbc\xdb\xb8\x5b\xf9\x8d" +
	"\x93\xa3\xe5\xf9\x0c\x61\x85\x2c\x47\xfd\x22\xc4\x27\x6b\x7e\x3d\xc6\xbf\x08\xf8\xdd\x25\x69\xdb\x76\x73\x4b\xfe" +
	"\x6f\xde\x2f\xca\x73\x2e\xec\x76\xea\x68\xfa\xb5\x82\x0f\xb0\x5b\xa9\x57\x48\xb5\x2c\x4d\x7a\x55\x69\x2e\xcd\x32" +
	"\x8e\xfe\xff\x18\x4d\x36\xd1\x93\x64\x8c\x35\x92\xef\x05\xe1\x5e\xa3\xdc\x7e\x90\x23\x35\x83\x36\xdc\x3e\x0f\x47" +
	"\x6d\x93\x1e\x31\xf9\x89\x3d\xe2\x6f\x2a\x7f\x1a\xee\x2c\x54\xfe\x34\x01\xd4\xba\xeb\xdc\x8f\x68\xc8\xc3\xb1\x3a" +
	"\xc3\x9c\x33\xff\x0e\x1a\x15\x75\xc7\xd6\x6a\xdb\x59\x5f\x65\x92\x40\x6b\xaa\x8c\xe4\x62\x28\x87\x50\x45\x7a\xc2" +
	"\x0c\x13\x71\x92\x7e\xd0\x3a\x46\xad\x93\xf4\xac\x2e\xe2\xe8\x46\xb2\x85\x40\x30\x0a\x0a\x34\x40\x94\xba\x2d\xd5" +
	"\xef\x11\x0a\x4a\x06\x8a\x3a\xda\x56\xbb\x05\x3d\x52\xd2\xa0\x34\x87\xc4\x3c\x9a\xc0\xf3\x54\x92\x94\x72\xf4\x8b" +
	"\x91\xc2\x26\xdb\x12\x6e\x4f\x76\x5d\xf5\x2a\x11\xd8\xb1\x8a\x87\x3d\xba\x23\xd9\x97\x73\x25\xd9\xb0\x36\xb0\x64" +
	"\x5c\x60\xde\xe5\xd9\x01\xd1\x5b\x3d\xc7\x4c\xe5\x98\x03\x97\x06\xf5\x92\x65\xd8\xb4\x1b\x50\xbe\x50\x37\xb2\x64" +
	"\xba\x5e\x31\x11\x3b\x76\x6f\xfd\xbd\xe4\xd7\xfd\x08\xf5\x71\x04\x7d\xcb\x52\x2c\x25\x6b\x7c\x81\x1f\x01\x9f\x28" +
	"\x5d\x32\x63\x50\xfb\x5f\x71\x07\x6c\x5d\xdc\xcb\xd5\x32\xa6\xef\x40\x2a\x8b\x7f\x65\xd1\x5b\x2c\x09\x5f\xf5\x31" +
	"\x34\x9a\x5c\x5b\xb2\x88\x3e\x40\x91\xc9\x61\x96\xb2\x32\x4f\x4f\x04\x2b\xea\x98\x6a\xa9\xc4\x1f\x4c\x7f\x89\xdf" +
	"\x6e\x4c\x8a\x2f\xfc\xd1\xe7\xd3\x61\xb8\x22\x1a\x30\x26\x6a\xf4\xc6\xe7\x5f\x78\x3b\x66\xd9\x51\xe0\xd2\xfc\xf2" +
	"\xf3\x6e\x02\xa7\x64\xda\x83\xc1\x8f\xfb\xa3\x2f\x85\x62\x2f\xe2\x9f\x38\xe3\x3e\x0c\xd2\xd7\x71\xd8\x85\xe6\xc6" +
	"\x66\x0f\xb0\x28\xfa\x2e\x56\xbf\xc6\x36\x9e\x5a\x1a\xb2\x61\x16\xdb\xf0\x9f\x00\x00\x00\xff\xff\xd5\xdd\x96\x04" +
	"\x2f\x0f\x00\x00")

func bindataTemplatesCommandstmplBytes() ([]byte, error) {
	return bindataRead(
		_bindataTemplatesCommandstmpl,
		"templates/commands.tmpl",
	)
}



func bindataTemplatesCommandstmpl() (*asset, error) {
	bytes, err := bindataTemplatesCommandstmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name: "templates/commands.tmpl",
		size: 3887,
		md5checksum: "",
		mode: os.FileMode(420),
		modTime: time.Unix(1538974665, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataTemplatesMaintmpl = []byte(
	"\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x44\x8e\x41\x4b\xc3\x40\x10\x85\xcf\x3b\xbf\x62\xc8\x41\x12\xb0\x9b\x7a" +
	"\xed\xad\x68\x0e\x5e\xac\x88\x78\x5f\x37\x93\xed\x60\x76\x66\xd9\x6c\x4a\x25\xe4\xbf\x4b\x2a\xe2\xed\xbd\xf7\xf1" +
	"\x1e\x2f\x39\xff\xe5\x02\x61\x74\x2c\x00\x1c\x93\xe6\x82\x35\x98\x2a\x70\x39\xcf\x9f\xd6\x6b\x6c\x7b\x27\x4c\x63" +
	"\x28\xee\x7b\xd4\xdc\x6a\x22\x71\x89\x77\x7e\xe4\x5d\x20\xa1\xec\x8a\xe6\xd6\x8f\x5c\x41\x03\x30\xcc\xe2\x6f\x63" +
	"\x75\x83\x0b\x18\x3f\xb2\x7d\x16\x2e\xf5\xdd\xa6\x1e\x55\x06\x0e\x0b\x18\x73\x4c\xe9\xc5\x45\x3a\x20\x62\xb5\x2c" +
	"\x68\x37\x83\xeb\x5a\xdd\x83\x31\x9d\x5c\x5e\x33\x0d\x7c\x3d\xfc\xb3\x4e\x2e\x7f\xf8\x83\xf2\xc4\x2a\xb7\xea\x83" +
	"\xdd\xdb\xfd\x96\xae\x0d\x80\x69\x5b\x7c\x3f\x3d\x9d\x0e\x78\xec\x7b\xcc\x14\x78\x2a\x94\xd1\x6b\x8c\x4e\xfa\x09" +
	"\xcf\x94\xc9\xc2\xef\xa7\x37\xd5\x62\xbb\x2b\xf9\xb9\x50\xdd\xc0\x0a\x3f\x01\x00\x00\xff\xff\xd7\x90\x9c\xb4\x08" +
	"\x01\x00\x00")

func bindataTemplatesMaintmplBytes() ([]byte, error) {
	return bindataRead(
		_bindataTemplatesMaintmpl,
		"templates/main.tmpl",
	)
}



func bindataTemplatesMaintmpl() (*asset, error) {
	bytes, err := bindataTemplatesMaintmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name: "templates/main.tmpl",
		size: 264,
		md5checksum: "",
		mode: os.FileMode(420),
		modTime: time.Unix(1538448749, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}


//
// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
//
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}

//
// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
// nolint: deadcode
//
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

//
// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or could not be loaded.
//
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}

//
// AssetNames returns the names of the assets.
// nolint: deadcode
//
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

//
// _bindata is a table, holding each asset generator, mapped to its name.
//
var _bindata = map[string]func() (*asset, error){
	"templates/commands.tmpl": bindataTemplatesCommandstmpl,
	"templates/main.tmpl":     bindataTemplatesMaintmpl,
}

//
// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
//
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, &os.PathError{
					Op: "open",
					Path: name,
					Err: os.ErrNotExist,
				}
			}
		}
	}
	if node.Func != nil {
		return nil, &os.PathError{
			Op: "open",
			Path: name,
			Err: os.ErrNotExist,
		}
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}


type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{Func: nil, Children: map[string]*bintree{
	"templates": {Func: nil, Children: map[string]*bintree{
		"commands.tmpl": {Func: bindataTemplatesCommandstmpl, Children: map[string]*bintree{}},
		"main.tmpl": {Func: bindataTemplatesMaintmpl, Children: map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	return os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
