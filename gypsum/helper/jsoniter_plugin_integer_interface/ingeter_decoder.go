package jsoniter_plugin_integer_interface

import (
	"fmt"
	"strconv"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

func init() {
	// make jsoniter decode integer as int64 rather than float64
	jsoniter.RegisterTypeDecoderFunc("interface {}", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		// TODO: register locally rather than globally
		switch iter.WhatIsNext() {
		case jsoniter.NumberValue:
			num := iter.ReadNumber()
			if integer, err := strconv.ParseInt(string(num), 10, 64); err == nil {
				*(*interface{})(ptr) = integer
			} else {
				*(*interface{})(ptr), err = num.Float64()
				if err != nil {
					iter.ReportError("custom number reader", fmt.Sprintf("invalid number %s", num))
				}
			}
		default:
			*(*interface{})(ptr) = iter.Read()
		}
	})
}
