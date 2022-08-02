package cabfile

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	cabfile "github.com/google/go-cabfile/cabfile"
)

func getArtifact(c *http.Client, url string) (io.ReadSeeker, error) {
	resp, err := http.Get(metadataURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	return bytes.NewReader(buf.Bytes()), err
}

func ExampleNext() {
	c := &http.Client{}
	f, err := getArtifact(c, "http://cns.utoronto.ca/test/archive/wsusscan.cab")
	if err != nil {
		log.Fatal("Error reading ususscan example", err)
	}

	buf := make([]byte, 4)
	cabinet, err := cabfile.New(f)
	for {
		r, finfo, err := cabinet.Next()

		if err != nil {
			break
		}

		// Read 4 bytes from file
		_, err = r.Read(buf)

		// Print out file info and bytes
		fmt.Println("r", r, "finfo", finfo, "buf", buf)
	}
}
