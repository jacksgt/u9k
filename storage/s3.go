package storage

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/rhnvrm/simples3"

	"u9k/config"
)

var s3 *simples3.S3

func Init() {
	s3 = simples3.New(config.S3Region, config.S3AccessKey, config.S3SecretKey)
	if config.S3Endpoint != "" {
		s3.SetEndpoint(config.S3Endpoint)
	}

	log.Printf("Initialized S3 Storage Backend %s %s\n", config.S3Region, config.S3Endpoint)
}

// tests the connection by uploading, downloading and deleting a file
func Check() error {
	testFile := []byte("HelloWorld\nFooBar\nOneTwoThree\n")
	testKey := "connection-test.txt"
	err := StoreFile(testFile, testKey)
	if err != nil {
		return err
	}
	data, err := GetFile(testKey)
	if err != nil {
		return err
	}
	if !bytesAreEqual(testFile, data) {
		return fmt.Errorf("Uploaded and downloaded file differ, aborting.")
	}
	err = DeleteFile(testKey)
	if err != nil {
		return fmt.Errorf("Failed to delete file %s: %s", testKey, err)
	}

	return nil
}

// maybe this should just take models.File as an argument?
// or be a method on models.File? not sure ...
func FileKey(id string, filename string) string {
	return fmt.Sprintf("files/%s/%s", id, filename)
}

func StoreFileStream(fd io.ReadSeeker, key string, contentType string) error {
	// make sure we go to the beginning of the file
	fd.Seek(0, io.SeekStart)

	_, err := s3.FileUpload(simples3.UploadInput{
		Bucket:      config.S3Bucket,
		ObjectKey:   key,
		ContentType: contentType,
		FileName:    path.Base(key),
		Body:        fd,
	})
	if err != nil {
		log.Printf("Failed to upload file %s: %s\n", key, err)
		return err
	}
	return nil
}

func StoreFile(file []byte, key string) error {
	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(file)
	_, err := s3.FileUpload(simples3.UploadInput{
		Bucket:      config.S3Bucket,
		ObjectKey:   key,
		ContentType: contentType,
		FileName:    path.Base(key),
		Body:        bytes.NewReader(file),
	})
	if err != nil {
		log.Printf("Failed to upload file %s: %s\n", key, err)
		return err
	}

	return nil
}

func GetFile(key string) (buf []byte, err error) {
	file, err := GetFileStream(key)
	if err != nil {
		return buf, err
	}
	defer file.Close()

	// copy the entire file to memory
	buf, err = ioutil.ReadAll(file)
	return buf, err
}

func GetFileStream(key string) (io.ReadCloser, error) {
	var err error
	// workaround for https://github.com/rhnvrm/simples3/issues/7
	// need to URL-encode the filename, because it might contain special characters
	// only necessary for GET and DELETE
	key = encodeS3Uri(key)

	file, err := s3.FileDownload(simples3.DownloadInput{
		Bucket:    config.S3Bucket,
		ObjectKey: key,
	})
	if err != nil {
		log.Printf("Failed to download file %s: %s\n", key, err)
		return nil, err
	}

	return file, nil
}

func DeleteFile(key string) error {
	var err error
	// workaround for https://github.com/rhnvrm/simples3/issues/7
	// need to URL-encode the filename, because it might contain special characters
	// only necessary for GET and DELETE
	key = encodeS3Uri(key)

	err = s3.FileDelete(simples3.DeleteInput{
		Bucket:    config.S3Bucket,
		ObjectKey: key,
	})
	if err != nil {
		log.Printf("Failed to delete file %s: %s\n", key, err)
		return err
	}

	return nil
}

var encodings = map[string]string{
	`+`: `%2B`,
	`!`: `%21`,
	`"`: `%22`,
	`#`: `%23`,
	`$`: `%24`,
	`&`: `%26`,
	"`": `%27`,
	`(`: `%28`,
	`)`: `%29`,
	`*`: `%2A`,
	`,`: `%2C`,
	`:`: `%3A`,
	`;`: `%3B`,
	`=`: `%3D`,
	`?`: `%3F`,
	`@`: `%40`,
}

// workaround for https://github.com/rhnvrm/simples3/issues/7
// need to URL-encode the filename, because it might contain special characters
// only necessary for GET and DELETE
func encodeS3Uri(key string) string {
	regex := regexp.MustCompile(
		`(\+|!|"|#|\$|&|'|\(|\)|\*|\+|,|:|;|=|\?|@)`,
	)

	paths := strings.Split(key, "/")
	for i := range paths {
		paths[i] = regex.ReplaceAllStringFunc(
			url.PathEscape(paths[i]),
			func(match string) string {
				return encodings[match]
			},
		)
	}
	key = strings.Join(paths, "/")
	return key
}

// adapted from https://golangcode.com/get-the-content-type-of-file/
func getFileContentType(r io.Reader) string {
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := r.Read(buffer)
	if err != nil {
		log.Printf("Failed to detect Content-Type: %s\n", err)
		return "application/octet-stream"
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)
	return contentType
}

// checks whether two byte arrays have the same content
func bytesAreEqual(s1 []byte, s2 []byte) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}
