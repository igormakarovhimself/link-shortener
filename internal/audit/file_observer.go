package audit

import (
	"encoding/json"
	"os"
	"sync"
)

// FileObserver записывает аудит-события в файл в формате JSON.
type FileObserver struct {
	file *os.File
	mu   sync.Mutex
}

// NewFileObserver создает FileObserver, который пишет в файл по пути filePath.
func NewFileObserver(filePath string) (*FileObserver, error) {
	if filePath == "" {
		return nil, nil
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileObserver{
		file: file,
	}, nil
}

// OnAudit записывает событие в файл в виде одной JSON-строки.
func (f *FileObserver) OnAudit(event AuditEvent) {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	data = append(data, '\n')
	_, _ = f.file.Write(data)
}

// Close закрывает файл.
func (f *FileObserver) Close() error {
	if f.file != nil {
		return f.file.Close()
	}
	return nil
}
