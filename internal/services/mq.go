package services

import (
	"context"
	"fmt"
)

type MqReader interface {
	Read() (string, error)
}

type MqService struct {
	reader MqReader
}

func NewMqService(reader MqReader) *MqService {
	return &MqService{
		reader: reader,
	}
}

func (s *MqService) Read(ctx context.Context) (string, error) {
	result, err := s.reader.Read()
	if err != nil {
		return "", fmt.Errorf("error reading from queue: %w", err)
	}
	return result, nil
}
