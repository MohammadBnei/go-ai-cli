package service

import (
	"context"

	"github.com/blang/vfs"
)

type PromptConfig struct {
	ChatMessages   *ChatMessages
	PreviousPrompt string
	UserPrompt     string
	UpdateChan     chan *ChatMessage
	Contexts       IContextService
	Files          IFileService
}

type IContextService interface {
	AddContext(ctx context.Context, cancelFn func())
	AddContextWithId(ctx context.Context, cancelFn func(), id int64)
	CloseContext(ctx context.Context) error
	FindContextWithId(id int64) *ContextHolder
	CloseContextById(id int64) error
	Length() int
}

type IFileService interface {
	Append(fileType FileType, msgContent string, originalFileName string, msgId int64, data []byte) (*FileMetadata, error)
	Delete(id string) error
	Get(id string) (vfs.File, *FileMetadata, error)
	GetByMsgId(id int64, fileType FileType) (vfs.File, *FileMetadata, error)
	List(fileType FileType) ([]FileMetadata, error)
	SaveToOS(id string, destinationPath string) error
}
