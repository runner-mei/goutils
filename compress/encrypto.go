package compress

import (
	"io"
	"log"
	"os"

	"github.com/yeka/zip"
)

type encryptoZipWriter struct {
	baseWriter

	password string
	file     *os.File
	gw       *zip.Writer
}

func (w *encryptoZipWriter) Close() error {
	e1 := w.gw.Close()
	e2 := w.file.Close()

	if e1 != nil {
		return e1
	}
	return e2
}

func (w *encryptoZipWriter) zipFile(name string, fr *os.File, fi os.FileInfo) error {
	log.Println("add file -", name)

	hdr, err := zip.FileInfoHeader(fi)
	if err != nil {
		return err
	}
	hdr.Name = name
	hdr.Flags |= 1 << 11 // 使用utf8编码
	hdr.Method = zip.Deflate
	hdr.SetPassword(w.password)
	hdr.SetEncryptionMethod(zip.AES256Encryption)

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
func EncryptZip(destFilePath, password string) (Writer, error) {
	fw, err := os.Create(destFilePath)
	if err != nil {
		return nil, err
	}

	// zip writer
	gw := zip.NewWriter(fw)

	tz := &encryptoZipWriter{
		password: password,
		file:     fw,
		gw:       gw,
	}
	tz.addFile = tz.zipFile
	return tz, nil
}

// func ExampleWriter_Encrypt() {
// 	contents := []byte("Hello World")

// 	// write a password zip
// 	raw := new(bytes.Buffer)
// 	zipw := zip.NewWriter(raw)
// 	w, err := zipw.Encrypt("hello.txt", "golang", zip.AES256Encryption)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	_, err = io.Copy(w, bytes.NewReader(contents))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	zipw.Close()

// 	// read the password zip
// 	zipr, err := zip.NewReader(bytes.NewReader(raw.Bytes()), int64(raw.Len()))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	for _, z := range zipr.File {
// 		z.SetPassword("golang")
// 		rr, err := z.Open()
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		_, err = io.Copy(os.Stdout, rr)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		rr.Close()
// 	}
// 	// Output:
// 	// Hello World
// }
