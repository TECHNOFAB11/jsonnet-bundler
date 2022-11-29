// Copyright 2018 jsonnet-bundler authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkg

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type HttpPackage struct {
	Source *deps.Http
}

func NewHttpPackage(source *deps.Http) Interface {
	return &HttpPackage{
		Source: source,
	}
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

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

func (h *HttpPackage) Install(ctx context.Context, name, dir, version string) (string, error) {
	destPath := path.Join(dir, name)

	pkgh := sha256.Sum256([]byte(fmt.Sprintf("jsonnetpkg-%s-%s", strings.Replace(name, "/", "-", -1), strings.Replace(version, "/", "-", -1))))
	// using 16 bytes should be a good middle ground between length and collision resistance
	tmpDir, err := ioutil.TempDir(filepath.Join(dir, ".tmp"), hex.EncodeToString(pkgh[:16]))
	if err != nil {
		return "", errors.Wrap(err, "failed to create tmp dir")
	}
	defer os.RemoveAll(tmpDir)

	err = os.Setenv("VERSION", version)
	if err != nil {
		return "", err
	}
	url := os.ExpandEnv(h.Source.Url)

	filename := filepath.Base(url)
	err = downloadFile(path.Join(tmpDir, filename), url)
	if err != nil {
		return "", errors.Wrap(err, "failed to download file")
	}

	var ar *os.File
	ar, err = os.Open(path.Join(tmpDir, filename))
	if err != nil {
		return "", errors.Wrap(err, "failed to open downloaded archive")
	}
	defer ar.Close()
	err = GzipUntar(destPath, ar, "")
	if err != nil {
		return "", errors.Wrap(err, "failed to unpack downloaded archive")
	}

	return version, nil
}
