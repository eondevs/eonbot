package file

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var (
	ErrFileNotFound      = errors.New("file not found")
	ErrDirectoryNotFound = errors.New("directory not found")
)

var (
	jsonIndent  = "    " // 4 spaces
	jsonFileExt = "json"
)

// Exists checks if file in specified path exists.
// Returns true if exists.
func Exists(p string) bool {
	_, err := os.Stat(p)
	return !os.IsNotExist(err)
}

// RequiredExists checks if file in specified path exists.
// Returns nil if exists.
func RequiredExists(p string, dir bool) error {
	if !Exists(p) {
		// if directory is required, return
		// directory-specific error.
		if dir {
			return ErrDirectoryNotFound
		}
		return ErrFileNotFound
	}
	return nil
}

// ExecDir returns executable file execution directory.
func ExecDir() (dir string, err error) {
	p, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}

	dir = filepath.Dir(absPath)

	return dir, nil
}

// JSONPath joins all provided elements into one valid file path and
// adds JSON file extension to the last element.
// Returns empty string if no elements are provided.
func JSONPath(elem ...string) string {
	if len(elem) == 0 {
		return ""
	}

	f := elem[len(elem)-1]
	if path.Ext(f) == "" {
		elem[len(elem)-1] = f + "." + jsonFileExt
	}

	return path.Join(elem...)
}

// RemoveExt removes file extension.
func RemoveExt(p string) string {
	return strings.TrimSuffix(path.Base(p), filepath.Ext(p))
}

// IsJSONExt returns true if file name has JSON file extension.
func IsJSONExt(name string) bool {
	return path.Ext(name) == "."+jsonFileExt
}

// HasSuffix returns true if specified file has a specified suffix
// after the delimiter symbol and before the file extension.
func HasSuffix(p, delim, suffix string) bool {
	return strings.HasSuffix(RemoveExt(p), delim+suffix)
}

// RemoveSuffix returns file name without its suffix
// and extension.
func RemoveSuffix(p, delim, suffix string) string {
	return strings.TrimSuffix(RemoveExt(p), delim+suffix)
}

// Load reads file contents and returns it in a byte
// slice format.
func Load(p string) ([]byte, error) {
	if err := RequiredExists(p, false); err != nil {
		return nil, err
	}
	return ioutil.ReadFile(p)
}

// LoadJSON reads JSON file contents and unmarshals it into
// specified target pointer.
func LoadJSON(p string, target interface{}) error {
	d, err := Load(p)
	if err != nil {
		return err
	}

	return json.Unmarshal(d, target)
}

// Save writes byte slice into specified file.
// If file or its parent directories do not exist, they will be created.
// If file already exists, it will be replaced.
func Save(p string, d []byte) error {
	// create directory if it does not exist
	if err := PrepDir(p); err != nil {
		return err
	}

	return ioutil.WriteFile(p, d, 0644)
}

// SaveJSON marshals and indents provided structure and writes it to file.
func SaveJSON(p string, v interface{}) error {
	d, err := json.MarshalIndent(v, "", jsonIndent)
	if err != nil {
		return err
	}

	return Save(p, d)
}

// SaveJSONBytes indents provided byte slice and writes it to file.
func SaveJSONBytes(p string, d []byte) error {
	var out bytes.Buffer
	if err := json.Indent(&out, d, "", jsonIndent); err != nil {
		return err
	}

	return Save(p, out.Bytes())
}

// PrepDir creates all parent directories of the path.
// If directories already exist, no changes will be made.
func PrepDir(p string) error {
	return os.MkdirAll(path.Dir(p), os.ModePerm)
}

// Remove deletes specified file.
func Remove(p string) error {
	if err := RequiredExists(p, false); err != nil {
		return err
	}

	return os.Remove(p)
}
