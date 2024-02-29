package service

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"github.com/bwmarrin/snowflake"
)

type FileType string

const (
	Audio FileType = "audio"
	Image FileType = "image"
	Text  FileType = "text"
)

type FileMetadata struct {
	ID         string    `json:"id"`
	FileType   FileType  `json:"fileType"`
	FileName   string    `json:"fileName"`
	MsgID      int64     `json:"msgId"`
	MsgContent string    `json:"msgContent"`
	Timestamp  time.Time `json:"timestamp"`
}

type FileService struct {
	fs       vfs.Filesystem
	metadata map[string]FileMetadata
	node     *snowflake.Node
}

func NewFileService() (*FileService, error) {
	node, _ := snowflake.NewNode(1)
	fs := memfs.Create()
	err := fs.Mkdir("data", 0777)
	if err != nil {
		return nil, err
	}
	return &FileService{
		fs:       fs,
		metadata: make(map[string]FileMetadata),
		node:     node,
	}, nil
}

func (s *FileService) Append(fileType FileType, msgContent, originalFileName string, msgId int64, data []byte) (*FileMetadata, error) {

	fm := FileMetadata{
		ID:         s.node.Generate().String(),
		FileType:   fileType,
		FileName:   originalFileName,
		MsgID:      msgId,
		MsgContent: msgContent,
		Timestamp:  time.Now(),
	}

	filePath := filepath.Join("data", fm.ID)

	file, err := s.fs.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return nil, err
	}

	s.metadata[fm.ID] = fm

	return &fm, nil
}

func (s *FileService) List(fileType FileType) ([]FileMetadata, error) {
	var files []FileMetadata
	for _, meta := range s.metadata {
		if meta.FileType == fileType {
			files = append(files, meta)
		}
	}
	return files, nil
}

func (s *FileService) Get(id string) (vfs.File, *FileMetadata, error) {
	file, err := s.fs.OpenFile(filepath.Join("data", id), os.O_RDONLY, 0644)
	if err != nil {
		return nil, nil, err
	}

	meta, ok := s.metadata[id]
	if !ok {
		return nil, nil, errors.New("file metadata not found")
	}
	return file, &meta, nil
}

func (s *FileService) GetByMsgId(id int64, fileType FileType) (vfs.File, *FileMetadata, error) {
	for _, meta := range s.metadata {
		if meta.MsgID == id && meta.FileType == fileType {
			file, err := s.fs.OpenFile(filepath.Join("data", meta.ID), os.O_RDONLY, 0644)
			if err != nil {
				return nil, nil, err
			}
			return file, &meta, nil
		}
	}
	return nil, nil, errors.New("file not found")
}

func (s *FileService) Delete(id string) error {
	filePath := filepath.Join("data", id)
	if err := s.fs.Remove(filePath); err != nil {
		return err
	}
	delete(s.metadata, id)
	return nil
}

func (s *FileService) SaveToOS(id string, destinationPath string) error {
	filePath := filepath.Join("data", id)
	file, err := s.fs.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	return os.WriteFile(destinationPath, data, 0644)
}
