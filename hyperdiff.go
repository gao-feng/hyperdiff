package main

import (
	"os"
	"io"
	"fmt"
	"flag"
	"diff2archive/ioutils"
	"diff2archive/archive"
)

// Diff produces an archive of the changes between the specified
// layer and its parent layer which may be "".
func Diff(layerFs, parentFs string) (arch archive.Archive, err error) {

	if parentFs == "" {
		archive, err := archive.Tar(layerFs, archive.Uncompressed)
		if err != nil {
			return nil, err
		}
		return ioutils.NewReadCloserWrapper(archive, func() error {
			err := archive.Close()
			return err
		}), nil
	}

	changes, err := archive.ChangesDirs(layerFs, parentFs)
	if err != nil {
		return nil, err
	}

	archive, err := archive.ExportChanges(layerFs, changes)
	if err != nil {
		return nil, err
	}

	return ioutils.NewReadCloserWrapper(archive, func() error {
		err := archive.Close()
		return err
	}), nil

}

func main() {
	data := make([]byte, 1024)
	var n int

	layer := flag.String("layer", "", "Layer directroy")
	parent := flag.String("parent", "", "Parent directroy")
	tar := flag.String("tar", "", "Target tar file")

	flag.Parse()

	if *layer == "" {
		fmt.Printf("Please specify Layer directroy\n")
		return
	}

	if *tar == "" {
		fmt.Printf("Please specify Target tar file\n")
		return
	}

	reader, err := Diff(*layer, *parent)
	if err != nil {
		fmt.Printf("Diff result %s\n", err.Error())
		return
	}

	defer reader.Close()

	dst, err := os.Create(*tar)
	if err != nil {
		fmt.Printf("create file failed %s\n", err.Error())
		return
	}

	defer dst.Close()

	for (true) {
		n, err = reader.Read(data)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("read failed %s\n", err.Error())
			}
			fmt.Printf("finish read\n")
			break
		}
		fmt.Printf("read %d byte\n", n)

		n, err = dst.Write(data[:n])
		if err != nil {
			fmt.Printf("write failed %s\n", err.Error())
			break
		}

		fmt.Printf("write %d byte\n", n)
	}

}
