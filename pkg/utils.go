package pkg

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
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
