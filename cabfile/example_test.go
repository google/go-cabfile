package cabfile

import (
	"fmt"
	"log"
	"os"

	cabfile "github.com/google/go-cabfile/cabfile"
)

func ExampleNext() {
	f, err := os.Open("Data1.cab")
	if err != nil {
		log.Fatal("Error reading file", err)
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
