// Package sexp parses R s expression trees into native go objects
package sexp

import (
	"encoding/binary"
	"errors"
	"log"
	"strconv"
)

// Parse converts a byte array containing R SEXP to a golang object.
// This can be converted to native golang types.
func Parse(buf []byte, offset int) (interface{}, error) {
	obj, _, err := parseReturningOffset(buf, offset)
	return obj, err
}

func getLength(buf []byte, offset int, isLong bool) (int, error) {
	throwError := func() (int, error) {
		return offset, errors.New("Abruptly reached end of buffer")
	}
	if isLong {
		if len(buf) <= offset+4 {
			return throwError()
		}
		return int(binary.LittleEndian.Uint32(buf[offset+1 : offset+5])), nil
	}
	if len(buf) <= offset+3 {
		return throwError()
	}
	return int(uint32(buf[offset+1]) | (uint32(buf[offset+2]) << 8) | (uint32(buf[offset+3]) << 16)), nil
}

func parseReturningOffset(buf []byte, offset int) (interface{}, int, error) {
	isLong := ((buf[offset] & 64) != 0)
	length, err := getLength(buf, offset, isLong)
	if err != nil {
		return nil, len(buf), err
	}
	xt := expressionType(buf[offset] & 63)

	hasAtt := ((buf[offset] & 128) != 0)

	offset = offset + 4
	if isLong {
		offset = offset + 4
	}
	end := offset + length

	log.Printf("parseREXP: type=%v, len=%v, hasAtt=%v, isLong=%v\n", xt, length, hasAtt, isLong)

	var attr interface{}
	if hasAtt {
		var err error
		attr, offset, err = parseReturningOffset(buf, offset)
		if err != nil {
			return nil, offset, err
		}
	}

	if xt == xtNull {
		return nil, offset, nil
	}
	if xt == xtInt {
		return parseInt(buf, offset, end)
	}
	if xt == xtString {
		return parseString(buf, offset, end)
	}
	/*
		if xt == xtS4 {
			//TODO
			return nil, offset, errors.New("Unimplemented expression type: XT_S4")
		}
		if xt == xtClos {
			//TODO
			return nil, offset, errors.New("Unimplemented expression type: XT_CLOS")
		}
	*/
	if xt == xtSymName {
		return parseSymName(buf, offset, end)
	}
	if xt == xtDoubleArray {
		return parseDoubleArray(buf, offset, end)
	}
	if xt == xtStringArray {
		return parseStringArray(buf, offset, end)
	}
	if xt == xtIntArray {
		return parseIntArray(buf, offset, end)
	}
	if xt == xtBoolArray {
		return parseBoolArray(buf, offset, end)
	}
	if xt == xtVector {
		return parseVector(attr, buf, offset, end)
	}
	/*
		if xt == xtListNoTag {
			return parseListNoTag(attr, buf, offset, end)
		}
	*/
	if xt == xtListTag {
		return parseListTag(buf, offset, end)
	}
	/*
		if xt == xtLangNoTag {
			return parseListNoTag(attr, buf, offset, end)
		}
	*/
	if xt == xtLangTag {
		return parseLangTag(buf, offset, end)
	}
	/*
		if xt == xtExpVector {
			return parseVector(attr, buf, offset, end)
		}
	*/
	if xt == xtRaw {
		return parseRaw(buf, offset, end)
	}
	if xt == xtComplexArray {
		return parseComplexArray(buf, offset, end)
	}
	/*
		if xt == xtUnknown {
			//TODO
			return nil, offset, errors.New("Unimplemented expression type: XT_UNKNOWN")
		}
	*/
	return nil, offset, errors.New("Unsupported expression type: " + strconv.Itoa(int(xt)))
}
