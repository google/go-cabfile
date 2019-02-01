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

// Package lvfscab allows to access LVFS metadata within a given Microsoft
// Cabinet file.
package lvfscab

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/blang/semver"

	"github.com/google/go-cabfile/cabfile"
)

// LVFSCabinet provides read-only access to Cabinet files shipped by the
// Linux Vendor Firmware Service (LVFS).
type LVFSCabinet struct {
	*cabfile.Cabinet

	ID      string
	Version string
}

type component struct {
	ID      string    `xml:"id"`
	Release []release `xml:"releases>release"`
}

type release struct {
	Version string `xml:"version,attr"`
}

// New returns a new LVFSCabinet with the metadata already parsed.
func New(r io.ReadSeeker) (*LVFSCabinet, error) {
	cab, err := cabfile.New(r)
	if err != nil {
		return nil, err
	}

	var mdfn string
	for _, fn := range cab.FileList() {
		if strings.HasSuffix(fn, ".metainfo.xml") {
			mdfn = fn
			break
		}
	}
	if mdfn == "" {
		return nil, errors.New("LVFS cabinet does not contain required metadata")
	}

	mdr, err := cab.Content(mdfn)
	if err != nil {
		return nil, fmt.Errorf("could not get content of metadata file %q: %v", mdfn, err)
	}
	var mdbuf bytes.Buffer
	if _, err := io.Copy(&mdbuf, mdr); err != nil {
		return nil, fmt.Errorf("could not read from metadata file %q: %v", mdfn, err)
	}
	var c component
	if err := xml.Unmarshal(mdbuf.Bytes(), &c); err != nil {
		return nil, fmt.Errorf("could not parse metadata file %q: %v", mdfn, err)
	}
	if c.ID == "" {
		return nil, fmt.Errorf("could not determine component ID from metadata file %q", mdfn)
	}
	// The AppStream specification encourages to list multiple releases
	// in the metadata to provide update descriptions. We make the
	// assumption here that the first release matches the release we
	// downloaded. This might not necessarily be true as it stands,
	// however.
	if len(c.Release) < 1 || c.Release[0].Version == "" {
		return nil, fmt.Errorf("could not extract release information from metadata file %q: %v", mdfn, err)
	}
	return &LVFSCabinet{
		Cabinet: cab,
		ID:      c.ID,
		Version: c.Release[0].Version,
	}, nil
}

// CompareVersions compares two versions used in LVFS. If both versions parse
// as semantic versions, compare them using semver. Otherwise fall back to a
// string comparison.
func CompareVersions(v1, v2 string) int {
	s1, err1 := semver.Make(v1)
	s2, err2 := semver.Make(v2)
	if err1 == nil && err2 == nil {
		return s1.Compare(s2)
	}
	return strings.Compare(v1, v2)
}
