package compress

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
)

// func main() {
// 	os.Mkdir("/home/ty4z2008/tar", 0777)
// 	w, err := CopyFile("/home/ty4z2008/tar/1.pdf", "/home/ty4z2008/src/1.pdf")
// 	//targetfile,sourcefile
// 	if err != nil {
// 		fmt.Println(err.Error())
// 	}
// 	fmt.Println(w)

// 	TarGz("/home/ty4z2008/tar/1.pdf", "/home/ty4z2008/test.tar.gz") //压缩
// 	//UnTarGz("/home/ty4z2008/1.tar.gz", "/home/ty4z2008")     //解压
// 	os.RemoveAll("/home/ty4z2008/tar")

// 	fmt.Println("ok")
// }

// Writer 接口
type Writer interface {
	io.Closer
	Add(relPath, destPath string) error
	AddFile(relPath, destPath string) error
	AddDir(relPath, destPath string, skip func(fi os.FileInfo) bool) error
	AddPattern(relPath, destPath, pat string, isRel bool) error
}

type baseWriter struct {
	addFile func(name string, fr *os.File, fi os.FileInfo) error
}

func (w *baseWriter) Add(relPath, destPath string) error {
	return w.add(relPath, destPath, nil)
}

func (w *baseWriter) AddFile(relPath, destPath string) error {
	return w.add(relPath, destPath, nil)
}

func (w *baseWriter) AddPattern(relPath, destPath, pat string, isRel bool) error {
	pa := filepath.Join(destPath, pat)
	matches, err := filepath.Glob(pa)
	if err != nil {
		return err
	}
	for _, match := range matches {
		if !isRel {
			err = w.AddFile(relPath, match)
			if err != nil {
				return err
			}
			continue
		}

		relPa, err := filepath.Rel(destPath, match)
		if err != nil {
			return err
		}
		dir := filepath.Dir(relPa)
		if dir == "" || dir == "." {
			err = w.AddFile(relPath, match)
		} else {
			err = w.AddFile(filepath.Join(relPath, dir), match)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *baseWriter) AddDir(relPath, destPath string, skip func(fi os.FileInfo) bool) error {
	fr, err := os.Open(destPath)
	if err != nil {
		return err
	}
	defer fr.Close()

	// if fi := fr.Stat(); fi.IsDir() {
	// 	return errors.New("'" + destPath + "' isnot directory")
	// }

	return w.addDir(relPath, destPath, fr, skip)
}

func (w *baseWriter) add(relPath, destPath string, fi os.FileInfo) error {
	// Check if it's a file or a directory
	fr, err := os.Open(destPath)
	if err != nil {
		return err
	}
	defer fr.Close()

	if fi == nil {
		fi, err = fr.Stat()
		if err != nil {
			return err
		}
	}

	pa := filepath.ToSlash(filepath.Join(relPath, fi.Name()))
	if fi.IsDir() {
		return w.addDir(pa, destPath, fr, nil)
	}
	return w.addFile(pa, fr, fi)
}

func (w *baseWriter) addDir(relPath, destPath string, dir *os.File, skip func(fi os.FileInfo) bool) error {
	log.Println("add directory -", destPath)

	// Get file info slice
	fiList, err := dir.Readdir(0)
	if err != nil {
		return err
	}

	for _, fi := range fiList {
		if skip != nil && skip(fi) {
			continue
		}
		err = w.add(relPath, filepath.Join(destPath, fi.Name()), fi)
		if err != nil {
			return err
		}
	}
	return nil
}

type tarGzWriter struct {
	baseWriter

	file *os.File
	gw   *gzip.Writer
	tw   *tar.Writer
}

func (w *tarGzWriter) Close() error {
	e1 := w.tw.Close()
	e2 := w.gw.Close()
	e3 := w.file.Close()

	if e1 != nil {
		return e1
	}
	if e2 != nil {
		return e2
	}
	if e3 != nil {
		return e3
	}
	return nil
}

func (w *tarGzWriter) tarGzFile(name string, fr *os.File, fi os.FileInfo) error {
	log.Println("add file -", name)

	// Create tar header
	hdr := new(tar.Header)
	hdr.Name = name
	hdr.Size = fi.Size()
	hdr.Mode = int64(fi.Mode())
	hdr.ModTime = fi.ModTime()

	// Write hander
	err := w.tw.WriteHeader(hdr)
	if err != nil {
		return err
	}

	// Write file data
	_, err = io.Copy(w.tw, fr)
	return err
}

// TarGz 创建一个生成 targz 格式的
func TarGz(destFilePath string) (Writer, error) {
	fw, err := os.Create(destFilePath)
	if err != nil {
		return nil, err
	}

	// Gzip writer
	gw := gzip.NewWriter(fw)

	// Tar writer
	tw := tar.NewWriter(gw)

	tz := &tarGzWriter{
		file: fw,
		gw:   gw,
		tw:   tw,
	}
	tz.addFile = tz.tarGzFile
	return tz, nil
}

type zipWriter struct {
	baseWriter

	file *os.File
	gw   *zip.Writer
}

func (w *zipWriter) Close() error {
	e1 := w.gw.Close()
	e2 := w.file.Close()

	if e1 != nil {
		return e1
	}
	return e2
}

func (w *zipWriter) zipFile(name string, fr *os.File, fi os.FileInfo) error {
	log.Println("add file -", name)

	hdr, err := zip.FileInfoHeader(fi)
	if err != nil {
		return err
	}
	hdr.Name = name
	hdr.Flags = 1 << 11 // 使用utf8编码
	hdr.Method = zip.Deflate

	// Write hander
	zw, err := w.gw.CreateHeader(hdr)
	if err != nil {
		return err
	}

	// Write file data
	_, err = io.Copy(zw, fr)
	return err
}

// Zip 创建一个生成 zip 格式的
func Zip(destFilePath string) (Writer, error) {
	fw, err := os.Create(destFilePath)
	if err != nil {
		return nil, err
	}

	// zip writer
	gw := zip.NewWriter(fw)

	tz := &zipWriter{
		file: fw,
		gw:   gw,
	}
	tz.addFile = tz.zipFile
	return tz, nil
}

// // Ungzip and untar from source file to destination directory
// // you need check file exist before you call this function
// func UnTarGz(srcFilePath string, destDirPath string) {
// 	fmt.Println("UnTarGzing " + srcFilePath + "...")
// 	// Create destination directory
// 	os.Mkdir(destDirPath, os.ModePerm)

// 	fr, err := os.Open(srcFilePath)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer fr.Close()

// 	// Gzip reader
// 	gr, err := gzip.NewReader(fr)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer gr.Close()

// 	// Tar reader
// 	tr := tar.NewReader(gr)

// 	for {
// 		hdr, err := tr.Next()
// 		if err == io.EOF {
// 			// End of tar archive
// 			break
// 		}
// 		//handleError(err)
// 		fmt.Println("UnTarGzing file..." + hdr.Name)
// 		// Check if it is diretory or file
// 		if hdr.Typeflag != tar.TypeDir {
// 			// Get files from archive
// 			// Create diretory before create file
// 			os.MkdirAll(destDirPath+"/"+path.Dir(hdr.Name), os.ModePerm)
// 			// Write data to file
// 			fw, _ := os.Create(destDirPath + "/" + hdr.Name)
// 			if err != nil {
// 				panic(err)
// 			}
// 			_, err = io.Copy(fw, tr)
// 			if err != nil {
// 				panic(err)
// 			}
// 		}
// 	}
// }

// UnZip 解压
func UnZip(zipFile, dest string) error {
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		log.Println("unpacking", file.Name)
		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		filename := filepath.Join(dest, file.Name)
		err = os.MkdirAll(filepath.Dir(filename), 0777) //file.Mode())
		if err != nil {
			return err
		}
		w, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer w.Close()
		_, err = io.Copy(w, rc)
		if err != nil {
			return err
		}
		w.Close()
		rc.Close()
	}
	return nil
}
