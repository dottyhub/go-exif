package exif

import (
	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

var (
	parser *Parser
)

// ValueContext describes all of the parameters required to find and extract
// the actual tag value.
type ValueContext struct {
	unitCount       uint32
	valueOffset     uint32
	rawValueOffset  []byte
	addressableData []byte

	tagType   TagTypePrimitive
	byteOrder binary.ByteOrder

	// undefinedValueTagType is the effective type to use if this is an "undefined"
	// value.
	undefinedValueTagType TagTypePrimitive

	// undefinedValueUnitCount is the effective unit-count to use if this is an
	// "undefined" value.
	undefinedValueUnitCount uint32

	ifdPath string
	tagId   uint16
}

func newValueContext(ifdPath string, tagId uint16, unitCount, valueOffset uint32, rawValueOffset, addressableData []byte, tagType TagTypePrimitive, byteOrder binary.ByteOrder) *ValueContext {
	return &ValueContext{
		unitCount:       unitCount,
		valueOffset:     valueOffset,
		rawValueOffset:  rawValueOffset,
		addressableData: addressableData,

		tagType:   tagType,
		byteOrder: byteOrder,

		ifdPath: ifdPath,
		tagId:   tagId,
	}
}

func newValueContextFromTag(ite *IfdTagEntry, addressableData []byte, byteOrder binary.ByteOrder) *ValueContext {
	return newValueContext(
		ite.IfdPath,
		ite.TagId,
		ite.UnitCount,
		ite.ValueOffset,
		ite.RawValueOffset,
		addressableData,
		ite.TagType,
		byteOrder)
}

func (vc *ValueContext) SetUnknownValueParameters(tagType TagTypePrimitive, unitCount uint32) {
	vc.undefinedValueTagType = tagType
	vc.undefinedValueUnitCount = unitCount
}

func (vc *ValueContext) UnitCount() uint32 {
	return vc.unitCount
}

func (vc *ValueContext) ValueOffset() uint32 {
	return vc.valueOffset
}

func (vc *ValueContext) RawValueOffset() []byte {
	return vc.rawValueOffset
}

func (vc *ValueContext) AddressableData() []byte {
	return vc.addressableData
}

// isEmbedded returns whether the value is embedded or a reference. This can't
// be precalculated since the size is not defined for all types (namely the
// "undefined" types).
func (vc *ValueContext) isEmbedded() bool {
	tagType, unitCount := vc.effectiveValueParameters()

	return (tagType.Size() * int(unitCount)) <= 4
}

func (vc *ValueContext) effectiveValueParameters() (tagType TagTypePrimitive, unitCount uint32) {
	if vc.tagType == TypeUndefined {
		tagType = vc.undefinedValueTagType
		unitCount = vc.undefinedValueUnitCount

		if tagType == 0 {
			log.Panicf("undefined-value type not set")
		}
	} else {
		tagType = vc.tagType
		unitCount = vc.unitCount
	}

	return tagType, unitCount
}

func (vc *ValueContext) readRawEncoded() (rawBytes []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	tagType, unitCount := vc.effectiveValueParameters()

	unitSizeRaw := uint32(tagType.Size())

	if vc.isEmbedded() == true {
		byteLength := unitSizeRaw * unitCount
		return vc.rawValueOffset[:byteLength], nil
	} else {
		return vc.addressableData[vc.valueOffset : vc.valueOffset+unitCount*unitSizeRaw], nil
	}
}

// Format returns a string representation for the value.
//
// Where the type is not ASCII, `justFirst` indicates whether to just stringify
// the first item in the slice (or return an empty string if the slice is
// empty).
//
// Since this method lacks the information to process undefined-type tags (e.g.
// byte-order, tag-ID, IFD type), it will return an error if attempted. See
// `UndefinedValue()`.
func (vc *ValueContext) Format() (value string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	rawBytes, err := vc.readRawEncoded()
	log.PanicIf(err)

	phrase, err := Format(rawBytes, vc.tagType, false, vc.byteOrder)
	log.PanicIf(err)

	return phrase, nil
}

// FormatOne is similar to `Format` but only gets and stringifies the first
// item.
func (vc *ValueContext) FormatFirst() (value string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	rawBytes, err := vc.readRawEncoded()
	log.PanicIf(err)

	phrase, err := Format(rawBytes, vc.tagType, true, vc.byteOrder)
	log.PanicIf(err)

	return phrase, nil
}

func (vc *ValueContext) ReadBytes() (value []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	rawValue, err := vc.readRawEncoded()
	log.PanicIf(err)

	value, err = parser.ParseBytes(rawValue, vc.unitCount)
	log.PanicIf(err)

	return value, nil
}

func (vc *ValueContext) ReadAscii() (value string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	rawValue, err := vc.readRawEncoded()
	log.PanicIf(err)

	value, err = parser.ParseAscii(rawValue, vc.unitCount)
	log.PanicIf(err)

	return value, nil
}

func (vc *ValueContext) ReadAsciiNoNul() (value string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	rawValue, err := vc.readRawEncoded()
	log.PanicIf(err)

	value, err = parser.ParseAsciiNoNul(rawValue, vc.unitCount)
	log.PanicIf(err)

	return value, nil
}

func (vc *ValueContext) ReadShorts() (value []uint16, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	rawValue, err := vc.readRawEncoded()
	log.PanicIf(err)

	value, err = parser.ParseShorts(rawValue, vc.unitCount, vc.byteOrder)
	log.PanicIf(err)

	return value, nil
}

func (vc *ValueContext) ReadLongs() (value []uint32, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	rawValue, err := vc.readRawEncoded()
	log.PanicIf(err)

	value, err = parser.ParseLongs(rawValue, vc.unitCount, vc.byteOrder)
	log.PanicIf(err)

	return value, nil
}

func (vc *ValueContext) ReadRationals() (value []Rational, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	rawValue, err := vc.readRawEncoded()
	log.PanicIf(err)

	value, err = parser.ParseRationals(rawValue, vc.unitCount, vc.byteOrder)
	log.PanicIf(err)

	return value, nil
}

func (vc *ValueContext) ReadSignedLongs() (value []int32, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	rawValue, err := vc.readRawEncoded()
	log.PanicIf(err)

	value, err = parser.ParseSignedLongs(rawValue, vc.unitCount, vc.byteOrder)
	log.PanicIf(err)

	return value, nil
}

func (vc *ValueContext) ReadSignedRationals() (value []SignedRational, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	rawValue, err := vc.readRawEncoded()
	log.PanicIf(err)

	value, err = parser.ParseSignedRationals(rawValue, vc.unitCount, vc.byteOrder)
	log.PanicIf(err)

	return value, nil
}

// Values knows how to resolve the given value. This value is always a list
// (undefined-values aside), so we're named accordingly.
//
// Since this method lacks the information to process unknown-type tags (e.g.
// byte-order, tag-ID, IFD type), it will return an error if attempted. See
// `UndefinedValue()`.
func (vc *ValueContext) Values() (value interface{}, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if vc.tagType == TypeByte {
		value, err = vc.ReadBytes()
		log.PanicIf(err)
	} else if vc.tagType == TypeAscii {
		value, err = vc.ReadAscii()
		log.PanicIf(err)
	} else if vc.tagType == TypeAsciiNoNul {
		value, err = vc.ReadAsciiNoNul()
		log.PanicIf(err)
	} else if vc.tagType == TypeShort {
		value, err = vc.ReadShorts()
		log.PanicIf(err)
	} else if vc.tagType == TypeLong {
		value, err = vc.ReadLongs()
		log.PanicIf(err)
	} else if vc.tagType == TypeRational {
		value, err = vc.ReadRationals()
		log.PanicIf(err)
	} else if vc.tagType == TypeSignedLong {
		value, err = vc.ReadSignedLongs()
		log.PanicIf(err)
	} else if vc.tagType == TypeSignedRational {
		value, err = vc.ReadSignedRationals()
		log.PanicIf(err)
	} else if vc.tagType == TypeUndefined {
		log.Panicf("will not parse undefined-type value")

		// Never called.
		return nil, nil
	} else {
		log.Panicf("value of type [%s] is unparseable", vc.tagType)

		// Never called.
		return nil, nil
	}

	return value, nil
}

func init() {
	parser = &Parser{}
}