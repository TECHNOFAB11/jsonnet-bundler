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
	"net/url"
	"os"
	"path"
)

type GitlabRegistryPackage struct {
	Source *deps.GitlabRegistry
}

func NewGitlabRegistryPackage(source *deps.GitlabRegistry) Interface {
	return &GitlabRegistryPackage{
		Source: source,
	}
}

func buildUrl(src *deps.GitlabRegistry, version string) url.URL {
	host := "gitlab.com"
	if src.Host != "" {
		host = src.Host
	}
	filename := "package.tar.gz"
	if src.Filename != "" {
		filename = src.Filename
	}

	packageUrl := url.URL{
		Scheme: "https",
		Host:   host,
	}
	// this way is needed so that net/url does not double encode '/' to '%252F'
	return *packageUrl.JoinPath(
		"api/v4/projects",
		url.PathEscape(src.Project),
		"packages/generic",
		src.Package,
		version,
		filename,
	)
}

func (h *GitlabRegistryPackage) Install(ctx context.Context, name, dir, version string) (string, error) {
	destPath := path.Join(dir, name)

	packageUrl := buildUrl(h.Source, version)

	tmpDir, err := CreateTempDir(name, dir, version)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	err = DownloadAndUntarTo(tmpDir, packageUrl.String(), destPath)
	if err != nil {
		return "", err
	}

	return version, nil
}
