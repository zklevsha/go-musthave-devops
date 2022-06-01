package structs

import (
	"fmt"

	"github.com/zklevsha/go-musthave-devops/internal/hash"
)

type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Hash    string `json:"hash,omitempty"`
}

func (s *Response) CalculateHash(key string) string {
	return hash.Sign(key, fmt.Sprintf("msg:%s;err:%s", s.Message, s.Error))
}

func (s *Response) SetHash(key string) {
	s.Hash = s.CalculateHash(key)
}

func (s *Response) AsText() string {
	var msg string
	if s.Message != "" {
		msg = fmt.Sprintf("meassage:%s;", s.Message)
	}
	if s.Error != "" {
		msg += fmt.Sprintf("error:%s;", s.Error)
	}
	if s.Hash != "" {
		msg += fmt.Sprintf("hash:%s;", s.Hash)
	}
	return msg
}
