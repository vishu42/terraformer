package targz

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Tardir creates a tar file for the source directory and writes it to the destination file
func Tardir(srcDir, destFile string) (err error) {
	// make sure the srcDir exist before proceeding
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("the source directory %s does not exist - %v", srcDir, err)
	}

	// the destFile must not be present in the srcDir
	if _, err := os.Stat(filepath.Join(srcDir, destFile)); !os.IsNotExist(err) {
		return fmt.Errorf("the destination file %s must not exist in the source directory %s", destFile, srcDir)
	}

	// open a tar.gz file
	file, err := os.Create(destFile)
	if err != nil {
		return err
	}

	defer file.Close()

	// // create a gzip writer
	// gw := gzip.NewWriter(file)
	// defer gw.Close()

	// create a tar writer
	tw := tar.NewWriter(file)
	defer tw.Close()

	// walk through the source directory
	filepath.Walk(srcDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// get the path relative to the srcDir
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		// log the header name
		fmt.Println("+ " + header.Name)
		// set the header name relative to path
		header.Name = relPath

		// write header
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}
		// write the file to tarfile writer
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)

		if err != nil {
			return err
		}
		return nil
	})

	return nil
}
