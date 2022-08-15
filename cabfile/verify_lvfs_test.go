// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cabfile

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
)

const repoURL = "https://cdn.fwupd.org/downloads"

// Set mirrorURL to a file:/// URL (without trailing slash) in case you do not
// want to fetch from the internet. lvfs-website's contrib/sync-pulp.py can
// help you sync.
var mirrorURL = ""

// Metadata just provides a full list of all artifacts currently published
// by LVFS and ignores all other data in the XML.
type Metadata struct {
	Location []string `xml:"component>releases>release>location"`
}

func artifacts(c *http.Client, url string) ([]string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	r := io.Reader(resp.Body)
	if resp.ContentLength > 0 || req.URL.Scheme == "file" {
		// Legacy support: In Go 1.7 or greater decompression is automatic for HTTP.
		gzr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gzr.Close()
		r = gzr
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, err
	}
	var md Metadata
	err = xml.Unmarshal(buf.Bytes(), &md)
	return md.Location, err
}

func parseFile(t *testing.T, c *http.Client, u string) {
	if mirrorURL != "" {
		ur, err := url.Parse(u)
		if err != nil {
			t.Fatalf("Could not parse location URL %q: %v", u, err)
		}
		fn := path.Base(ur.Path)
		// TODO: Use url.JoinPath
		u = path.Join(mirrorURL, fn)
	}

	t.Logf("Fetching %s...", u)
	req, err := http.NewRequest("GET", u, nil)
	req.Header.Add("User-Agent", "Go-cabfile")
	resp, err := c.Do(req)
	if err != nil {
		t.Errorf("Could not fetch URL %q: %v", u, err)
		return
	}
	defer resp.Body.Close()

	tmpf, err := os.CreateTemp("", "verify_lvfs_test")
	if err != nil {
		t.Fatalf("Could not create temporary file: %v", err)
	}
	defer func() {
		os.Remove(tmpf.Name())
		tmpf.Close()
	}()
	if _, err := io.Copy(tmpf, resp.Body); err != nil {
		t.Errorf("Could not copy file content to temporary file: %v", err)
		return
	}
	cab, err := New(tmpf)
	if err != nil {
		t.Errorf("Failed to parse cab file from URL %q: %v", u, err)
		return
	}
	const metainfoName = ".metainfo.xml"
	fns := cab.FileList()
	for _, fn := range fns {
		if strings.HasSuffix(fn, metainfoName) {
			// Found the required file.
			return
		}
	}
	t.Errorf("Cabinet file downloaded from %q misses *%s member; file list: %v", u, metainfoName, fns)
}

// TestLVFSFileParsing downloads all available cab files from LVFS and checks
// if their file list contains a *.metainfo.xml file.
func TestLVFSFileParsing(t *testing.T) {
	// TODO: Use url.JoinPath (Go 1.19+)
	u := repoURL + "/firmware.xml.gz"
	tr := &http.Transport{}
	if mirrorURL != "" {
		tr.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
		// TODO: Use url.JoinPath (Go 1.19+)
		u = mirrorURL + "/firmware.xml.gz"
	}
	c := &http.Client{Transport: tr}

	// TODO: Unpack the cabs using gcab and verify that the content is
	// hash-identical to cabfile's unpacking method.

	t.Logf("Fetching %s...", u)
	urls, err := artifacts(c, u)
	if err != nil {
		t.Fatalf("Could not ingest metadata from %q: %v", u, err)
	}
	t.Logf("Found %d artifacts.", len(urls))
	for _, u := range urls {
		parseFile(t, c, u)
	}
}
