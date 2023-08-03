package pkg

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

// TODO: make sure it works for nested directories
func TarDir(src, dst string) (err error) {
	// the destFile must not be present in the srcDir
	if _, err := os.Stat(filepath.Join(src, dst)); !os.IsNotExist(err) {
		return fmt.Errorf("the destination file %s must not exist in the source directory %s", dst, src)
	}

	// open a tar.gz file
	file, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer file.Close()

	return Tar(src, file)
}

// Tar creates a tar file for the source directory and writes it to the destination file
func Tar(src string, writers ...io.Writer) (err error) {
	// make sure the srcDir exist before proceeding
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("the source directory %s does not exist - %v", src, err)
	}

	mw := io.MultiWriter(writers...)

	// create a gzip writer
	gw := gzip.NewWriter(mw)
	defer gw.Close()

	// create a tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	log.Println("Tar: src directory", src)
	// get a list of all files and folders in the srcDir
	// cmd := exec.Command("ls", "-lart", src)
	// stdoutStderr, err := cmd.CombinedOutput()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println(string(stdoutStderr))
	// walk through the source directory
	filepath.Walk(src, func(path string, fi fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		fmt.Printf("visited file or dir: %q\n", path)
		// return nil
		// log.Println("Tar: ", path)

		// // return any errors
		// if err != nil {
		// 	log.Printf("Tar: error walking the path %q: %v\n", src, err)
		// 	return err
		// }

		// get the path relative to the srcDir
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		log.Println("Tar: relPath ", relPath)

		// skip any non regular files
		if !fi.Mode().IsRegular() {
			fmt.Println("skipping non regular file: ", relPath)
			return nil
		}

		log.Println("Tar: File info: ", fi.Name(), " - ", fi.Size(), " - ", fi.Mode().IsRegular())
		header, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			log.Println("Tar: error getting header: ", err)
			return err
		}

		// log the header name
		fmt.Println("+ " + header.Name)
		// set the header name relative to path
		header.Name = relPath

		// log file info from header
		fmt.Println("Header info : name: ", header.FileInfo().Name())
		fmt.Println("Header info : size: ", header.FileInfo().Size())

		// write header
		err = tw.WriteHeader(header)
		if err != nil {
			log.Println("Tar: error writing header: ", err)
			return err
		}

		// write the file to tarfile writer
		file, err := os.Open(path)
		if err != nil {
			log.Println("Tar: error opening file: ", err)
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)

		if err != nil {
			log.Println("Tar: error copying file: ", err)
			return err
		}
		return nil
	})

	return nil
}

func Untar(dst string, reader io.Reader) (err error) {
	log.Println("Untar: ", dst)
	log.Println("Untar: creating gzip reader")
	gr, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gr.Close()

	log.Println("Untar: creating tar reader")
	tr := tar.NewReader(gr)

	for {
		log.Println("Untar: get header from tar reader")
		header, err := tr.Next()
		log.Println("Untar: header: ", header)
		switch {
		// if there are no more files left to be read
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		case header == nil:
			continue
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// if its a directory and it doesn't exist
			// create it with 0755 permission
			if _, err := os.Stat(target); os.IsNotExist(err) {
				if err := os.MkdirAll(target, 0o755); err != nil {
					return err
				}
			}

		case tar.TypeReg:
			// if it's a file create it
			// with 0666 permission

			// get the directory path and create if it doesn't exist
			dir := filepath.Dir(target)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return err
				}
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				log.Println("Untar: error opening file: ", err)
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()

		}
	}
}

func UntarTar(dst, src string) (err error) {
	log.Printf("UntarTar: %s -> %s", src, dst)
	// make sure the src file exist before proceeding
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("the source file %s does not exist - %v", src, err)
	}

	// open the tar.gz file
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	return Untar(dst, file)
}
