package pkg

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func GzipUntar(dst string, r io.Reader, subDir string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}

		// strip the two first components of the path
		parts := strings.SplitAfterN(header.Name, "/", 2)
		if len(parts) < 2 {
			continue
		}
		suffix := parts[1]
		prefix := dst

		// reconstruct the target parh for the archive entry
		target := filepath.Join(prefix, suffix)

		// if subdir is provided and target is not under it, skip it
		subDirPath := filepath.Join(prefix, subDir)
		if subDir != "" && !strings.HasPrefix(target, subDirPath) {
			continue
		}

		// check the file type
		switch header.Typeflag {

		// create directories as needed
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}

		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
				return err
			}

			err := func() error {
				f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
				if err != nil {
					return err
				}
				defer f.Close()

				// copy over contents
				if _, err := io.Copy(f, tr); err != nil {
					return err
				}
				return nil
			}()

			if err != nil {
				return err
			}

		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
				return err
			}

			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}
		}
	}
}

func CreateTempDir(name, dir, version string) (string, error) {
	pkgh := sha256.Sum256([]byte(fmt.Sprintf("jsonnetpkg-%s-%s", strings.Replace(name, "/", "-", -1), strings.Replace(version, "/", "-", -1))))
	// using 16 bytes should be a good middle ground between length and collision resistance
	tmpDir, err := ioutil.TempDir(filepath.Join(dir, ".tmp"), hex.EncodeToString(pkgh[:16]))
	if err != nil {
		return "", errors.Wrap(err, "failed to create tmp dir")
	}

	return tmpDir, nil
}

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	color.Cyan("GET %s %d", url, resp.StatusCode)

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func DownloadAndUntarTo(tmpDir, url, destPath string) error {
	filename := filepath.Base(url)
	err := DownloadFile(path.Join(tmpDir, filename), url)
	if err != nil {
		return errors.Wrap(err, "failed to download file")
	}

	var ar *os.File
	ar, err = os.Open(path.Join(tmpDir, filename))
	if err != nil {
		return errors.Wrap(err, "failed to open downloaded archive")
	}
	defer ar.Close()
	err = GzipUntar(destPath, ar, "")
	if err != nil {
		return errors.Wrap(err, "failed to unpack downloaded archive")
	}
	return nil
}
