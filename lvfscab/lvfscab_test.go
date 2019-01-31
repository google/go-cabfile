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

package lvfscab

import (
	"encoding/xml"
	"reflect"
	"testing"
)

func TestXMLParsing(t *testing.T) {
	const testData = `<?xml version="1.0" encoding="UTF-8"?>
<component type="firmware">
  <releases>
    <release urgency="low" version="1.2.6" timestamp="1480683870">
    </release>
  </releases>
</component>`
	want := component{
		[]release{release{Version: "1.2.6"}},
	}
	var md component
	if err := xml.Unmarshal([]byte(testData), &md); err != nil {
		t.Fatalf("Could not parse embedded XML data: %v", err)
	}
	if !reflect.DeepEqual(md, want) {
		t.Errorf("xml.Unmarshal = %#+v; want %#+v", md, want)
	}
}
