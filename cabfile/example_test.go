package cabfile

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	cabfile "github.com/google/go-cabfile/cabfile"
)

// Pull a file down and return a reader for the contents
func getArtifact(c *http.Client, url string) (io.ReadSeeker, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, err
	}
	return bytes.NewReader(buf.Bytes()), err
}

func TestNextCall(t *testing.T) {
	//exampleURL := "https://www2.census.gov/econ1997/VIUS/CD-ROM/install/vi97.cab"
	exampleURL := "http://cns.utoronto.ca/test/archive/wsusscan.cab.0601013"

	c := &http.Client{}
	f, err := getArtifact(c, exampleURL)
	if err != nil {
		t.Fatalf("Could not read example cab file %q: %v", exampleURL, err)
	}

	buf := make([]byte, 8)
	cabinet, err := cabfile.New(f)
	if err != nil {
		t.Fatalf("Could not parse example cab file %q: %v", exampleURL, err)
	}
	for {
		fmt.Printf("calling next")
		r, finfo, err := cabinet.Next()
		fmt.Printf("called next")

		if err != nil {
			return
		}

		// Read 4 bytes from file
		_, err = r.Read(buf)

		// Print out file info and bytes
		fmt.Printf("r %+v\n finfo: %+v\n buf: %+v\n", r, finfo, buf)
	}
}
