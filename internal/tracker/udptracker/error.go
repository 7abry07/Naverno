package udptracker

import (
	"bytes"
	"fmt"
	"io"
)

type errorResponse struct {
	message string
}

func (res *errorResponse) decode(data *bytes.Buffer) error {
	message := []byte{}
	message, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("error while reading error response")
	}

	res.message = string(message)
	return nil
}

func (errorResponse) Action() action {
	return action_error
}
