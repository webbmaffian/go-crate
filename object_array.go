package main

import (
	"fmt"

	"github.com/jackc/pgtype"
)

type ObjectArray struct {
	data any
}

func (dst *ObjectArray) Get() interface{} {
	return nil
}

func (dst *ObjectArray) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	fmt.Println("Decode text:", src, *dst)

	return nil
}

func (src *ObjectArray) AssignTo(dst interface{}) error {
	return fmt.Errorf("cannot assign %v to %T", src, dst)
}

func (dst *ObjectArray) Set(src interface{}) error {
	return fmt.Errorf("cannot convert %v to Point", src)
}
