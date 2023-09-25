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
	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
	"os"
	"path"
)

type HttpPackage struct {
	Source *deps.Http
}

func NewHttpPackage(source *deps.Http) Interface {
	return &HttpPackage{
		Source: source,
	}
}

func (h *HttpPackage) Install(ctx context.Context, name, dir, version string) (string, error) {
	destPath := path.Join(dir, name)

	err := os.Setenv("VERSION", version)
	if err != nil {
		return "", err
	}
	packageUrl := os.ExpandEnv(h.Source.Url)

	tmpDir, err := CreateTempDir(name, dir, version)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	err = DownloadAndUntarTo(tmpDir, packageUrl, destPath)
	if err != nil {
		return "", err
	}

	return version, nil
}
