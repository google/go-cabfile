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
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
)

const metadataURL = "https://cdn.fwupd.org/downloads/firmware.xml.gz"

// Metadata just provides a full list of all artifacts currently published
// by LVFS and ignores all other data in the XML.
type Metadata struct {
	Location []string `xml:"component>releases>release>location"`
}

func artifacts(c *http.Client, url string) ([]string, error) {
	resp, err := http.Get(metadataURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	r, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, err
	}
	var md Metadata
	err = xml.Unmarshal(buf.Bytes(), &md)
	return md.Location, err
}

// TestLVFSFileParsing downloads all available cab files from LVFS and checks
// if their file list contains a *.metainfo.xml file.
func TestLVFSFileParsing(t *testing.T) {
	c := &http.Client{}

	// TODO: Unpack the cabs using gcab and verify that the content is
	// hash-identical to cabfile's unpacking method.

	urls, err := artifacts(c, metadataURL)
	if err != nil {
		t.Fatalf("Could not ingest metadata from %q: %v", metadataURL, err)
	}
cabFile:
	for _, url := range urls {
		req, err := http.NewRequest("GET", url, nil)
		req.Header.Add("User-Agent", "Go-cabfile")
		resp, err := c.Do(req)
		if err != nil {
			t.Errorf("Could not fetch URL %q: %v", url, err)
			continue
		}
		defer resp.Body.Close()
		tmpf, err := ioutil.TempFile("", "verify_lvfs_test")
		if err != nil {
			t.Fatalf("Could not create temporary file: %v", err)
		}
		defer os.Remove(tmpf.Name())
		if _, err := io.Copy(tmpf, resp.Body); err != nil {
			t.Errorf("Could not copy file content to temporary file: %v", err)
			tmpf.Close()
			continue
		}
		c, err := New(tmpf)
		if err != nil {
			t.Errorf("Failed to parse cab file from URL %q: %v", url, err)
			tmpf.Close()
			continue
		}
		const metainfoName = ".metainfo.xml"
		fns := c.FileList()
		for _, fn := range fns {
			if strings.HasSuffix(fn, metainfoName) {
				// Found the required file.
				continue cabFile
			}
		}
		t.Errorf("Cabinet file downloaded from %q misses *%s member; file list: %v", url, metainfoName, fns)
	}
}
