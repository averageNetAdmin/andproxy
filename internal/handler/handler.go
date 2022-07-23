package handler

import (
	"fmt"
	"strings"

	"github.com/averageNetAdmin/andproxy/internal/handler/def"
	myhttp "github.com/averageNetAdmin/andproxy/internal/handler/http"
)

type Handler interface {
	Listen()
}

func NewHandler(filePath string) (Handler, error) {
	ind := strings.LastIndex(filePath, "/")
	if ind == -1 {
		return nil, fmt.Errorf("error parsing file path: invalid file path %s", filePath)
	}
	fileParts := strings.Split(filePath[ind+1:], "_")

	switch fileParts[0] {
	case "tcp4", "tcp6", "udp4", "udp6":
		h, err := def.NewHandler(filePath, fileParts[0], fileParts[1])
		if err != nil {
			return nil, err
		}
		return h, nil
	case "http":
		h, err := myhttp.NewHandler(filePath, fileParts[1], false)
		if err != nil {
			return nil, err
		}
		return h, nil
	case "https":
		h, err := myhttp.NewHandler(filePath, fileParts[1], true)
		if err != nil {
			return nil, err
		}
		return h, nil
	}
	return nil, nil
}
