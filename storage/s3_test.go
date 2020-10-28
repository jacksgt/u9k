package storage

import (
	"bytes"
	"io/ioutil"
	"testing"

	"u9k/config"
)

func TestFileKey(t *testing.T) {
	var got, want string

	got = FileKey("foobar", "example.pdf")
	want = "files/foobar/example.pdf"
	if got != want {
		t.Errorf("FileKey incorrect, got: %s, want: %s.", got, want)
	}
}

// tests StoreFileStream, GetFileStream and DeleteFile
func TestStoreGetDeleteFileStream(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config.Init()
	Init()

	inData := []byte("HelloWorld\nFooBar\nOneTwoThree\n")
	testKey := "test/TestStoreGetDeleteFileStream (!@&$).txt"

	err := StoreFileStream(bytes.NewReader(inData), testKey, "text/plain")
	if err != nil {
		t.Fatalf("Failed to store file %s: %s", testKey, err)
		return
	}
	outStream, err := GetFileStream(testKey)
	if err != nil {
		t.Fatalf("Failed to get file %s: %s", testKey, err)
	}

	outData, _ := ioutil.ReadAll(outStream)
	if !bytesAreEqual(inData, outData) {
		t.Fatalf("Uploaded and downloaded file differ, aborting.")
		return
	}
	err = DeleteFile(testKey)
	if err != nil {
		t.Fatalf("Failed to delete file %s: %s\n", testKey, err)
	}
}

// tests StoreFile, GetFile and DeleteFile
func TestStoreGetDeleteFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config.Init()
	Init()

	testFile := []byte("HelloWorld\nFooBar\nOneTwoThree\n")
	testKey := "test/TestStoreGetDeleteFile (!@&$).txt"

	err := StoreFile(testFile, testKey)
	if err != nil {
		t.Fatalf("Failed to store file %s: %s", testKey, err)
		return
	}
	data, err := GetFile(testKey)
	if err != nil {
		t.Fatalf("Failed to get file %s: %s", testKey, err)
	}
	if !bytesAreEqual(testFile, data) {
		t.Fatalf("Uploaded and downloaded file differ, aborting.")
		return
	}
	err = DeleteFile(testKey)
	if err != nil {
		t.Fatalf("Failed to delete file %s: %s\n", testKey, err)
	}
}
