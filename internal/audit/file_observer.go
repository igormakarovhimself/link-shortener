package audit

import (
	"encoding/json"
	"os"
	"sync"
)

type FileObserver struct {
	file *os.File
	mu   sync.Mutex
}

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

func (f *FileObserver) OnAudit(event AuditEvent) {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	data = append(data, '\n')
	f.file.Write(data)
}

func (f *FileObserver) Close() error {
	if f.file != nil {
		return f.file.Close()
	}
	return nil
}
