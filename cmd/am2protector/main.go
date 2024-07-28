package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"

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
	hashes := make(map[string]struct{})
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
		hashes[am2.HashAm2()] = struct{}{}

		return nil
	})
	if err != nil {
		panic(err)
	}
	if len(hashes) == 0 {
		fmt.Println("-- No am2 or am2data files")
		return
	}

	for hash, _ := range hashes {
		fmt.Printf("INSERT INTO protected_am2 (am2_hash, ref) VALUES ('%s', '%s');\n", hash, ref)
	}
}
