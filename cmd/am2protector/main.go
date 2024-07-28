package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/walterwanderley/am2manager"
)

func main() {
	var ref string
	flag.StringVar(&ref, "ref", "protected", "Link to more informations about the captures")
	flag.Parse()

	if ref == "" {
		panic("-ref can't be empty")
	}
	fsys := os.DirFS(".")
	hashes := make(map[string][]string)
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}

		if !am2manager.IsAm2(data) && !am2manager.IsAm2Data(data) {
			return nil
		}
		var am2 am2manager.Am2Data
		err = am2.UnmarshalBinary(data)
		if err != nil {
			return err
		}

		am2hash := am2.HashAm2()
		if _, ok := hashes[am2hash]; !ok {
			hashes[am2hash] = make([]string, 0)
		}
		filename := filepath.Base(p)
		hashes[am2hash] = append(hashes[am2hash], filename)

		return nil
	})
	if err != nil {
		panic(err)
	}
	if len(hashes) == 0 {
		fmt.Println("-- No am2 or am2data files")
		return
	}

	for hash, files := range hashes {
		fmt.Printf("INSERT OR IGNORE INTO protected_am2(am2_hash, ref) VALUES ('%s', '%s'); -- %s\n", hash, ref, strings.Join(files, ", "))
	}
}
