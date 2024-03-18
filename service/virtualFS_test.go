package service_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/goleak"

	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/service/godcontext"
)

func TestMain(m *testing.M) {
	godcontext.GodContext = context.Background()
	goleak.VerifyTestMain(m)

}

func TestAppend(t *testing.T) {
	fs, _ := service.NewFileService()

	_, err := fs.Append(service.Audio, "test_audio.wav", "test_audio.wav", 1, []byte("Sample audio content"))
	if err != nil {
		t.Errorf("Append failed: %v", err)
	}

	_, err = fs.Append(service.Text, "test_text.txt", "Sample text content", 2, []byte("Sample text content"))
	if err != nil {
		t.Errorf("Append failed: %v", err)
	}
}

func TestList(t *testing.T) {
	fs, _ := service.NewFileService()

	// Prepopulate with test data
	fs.Append(service.Audio, "test_audio.wav", "Sample audio content", 1, []byte("Sample audio content"))
	fs.Append(service.Text, "test_text.txt", "Sample text content", 2, []byte("Sample text content"))

	audioFiles, err := fs.List(service.Audio)
	if err != nil || len(audioFiles) != 1 {
		t.Errorf("List failed for Audio: %v", err)
	}

	textFiles, err := fs.List(service.Text)
	if err != nil || len(textFiles) != 1 {
		t.Errorf("List failed for Text: %v", err)
	}
}

func TestSaveToOS(t *testing.T) {
	fs, _ := service.NewFileService()

	// Prepopulate with test data
	fs.Append(service.Text, "test_text.txt", "Sample text content", 2, []byte("Sample text content"))
	textFiles, _ := fs.List(service.Text)

	tempDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = fs.SaveToOS(textFiles[0].ID, filepath.Join(tempDir, "saved_test_text.txt"))
	if err != nil {
		t.Errorf("SaveToOS failed: %v", err)
	}

	_, err = os.Stat(filepath.Join(tempDir, "saved_test_text.txt"))
	if os.IsNotExist(err) {
		t.Errorf("File was not saved to OS filesystem")
	}
}

func TestDelete(t *testing.T) {
	fs, _ := service.NewFileService()

	// Prepopulate with test data
	fs.Append(service.Audio, "test_audio.wav", "Sample audio content", 1, []byte("Sample audio content"))
	audioFiles, _ := fs.List(service.Audio)

	err := fs.Delete(audioFiles[0].ID)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	audioFilesAfterDelete, err := fs.List(service.Audio)
	if err != nil || len(audioFilesAfterDelete) != 0 {
		t.Errorf("File was not deleted properly")
	}
}
