package main

import (
	"bytes"
	"encoding/json"
	"errors"
)

func unmarshalArray(value any, dest any) (err error) {
	if value == nil {
		return
	}

	data, ok := value.([]byte)

	if !ok {
		return errors.New("Type assertion to []byte failed")
	}

	data = bytes.ReplaceAll(data, []byte("\\"), []byte(""))
	data = bytes.ReplaceAll(data, []byte("}\",\"{"), []byte("},{"))
	dataLen := len(data)
	data[1] = '['
	data[dataLen-2] = ']'
	data = data[1 : dataLen-1]

	return json.Unmarshal(data, dest)
}
