// Copyright 2018 Jsgenesis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grandpa_types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"reflect"
	"strings"

	"golang.org/x/crypto/blake2b"
)

// Implementation for Parity codec in Go.
// Derived from https://github.com/paritytech/parity-codec/
// While Rust implementation uses Rust type system and is highly optimized, this one
// has to rely on Go's reflection and thus is notably slower.
// Feature parity is almost full, apart from the lack of support for u128 (which are missing in Go).

const maxUint = ^uint(0)
const maxInt = int(maxUint >> 1)

// Encoder is a wrapper around a Writer that allows encoding data items to a stream.
// Allows passing encoding options
type Encoder struct {
	Writer io.Writer
}

func NewEncoder(writer io.Writer) *Encoder {
	return &Encoder{Writer: writer}
}

// Write several bytes to the encoder.
func (pe Encoder) Write(bytes []byte) error {
	c, err := pe.Writer.Write(bytes)
	if err != nil {
		return err
	}
	if c < len(bytes) {
		return fmt.Errorf("Could not write %d bytes to writer", len(bytes))
	}
	return nil
}

// PushByte writes a single byte to an encoder.
func (pe Encoder) PushByte(b byte) error {
	return pe.Write([]byte{b})
}

// EncodeUintCompact writes an unsigned integer to the stream using the compact encoding.
// A typical usage is storing the length of a collection.
// Definition of compact encoding:
// 0b00 00 00 00 / 00 00 00 00 / 00 00 00 00 / 00 00 00 00
//   xx xx xx 00															(0 ... 2**6 - 1)		(u8)
//   yL yL yL 01 / yH yH yH yL												(2**6 ... 2**14 - 1)	(u8, u16)  low LH high
//   zL zL zL 10 / zM zM zM zL / zM zM zM zM / zH zH zH zM					(2**14 ... 2**30 - 1)	(u16, u32)  low LMMH high
//   nn nn nn 11 [ / zz zz zz zz ]{4 + n}									(2**30 ... 2**536 - 1)	(u32, u64, u128, U256, U512, U520) straight LE-encoded
// Rust implementation: see impl<'a> Encode for CompactRef<'a, u64>
func (pe Encoder) EncodeUintCompact(v big.Int) error {
	if v.Sign() == -1 {
		return errors.New("Assertion error: EncodeUintCompact cannot process negative numbers")
	}

	if v.IsUint64() {
		if v.Uint64() < 1<<30 {
			if v.Uint64() < 1<<6 {
				err := pe.PushByte(byte(v.Uint64()) << 2)
				if err != nil {
					return err
				}
			} else if v.Uint64() < 1<<14 {
				err := binary.Write(pe.Writer, binary.LittleEndian, uint16(v.Uint64()<<2)+1)
				if err != nil {
					return err
				}
			} else {
				err := binary.Write(pe.Writer, binary.LittleEndian, uint32(v.Uint64()<<2)+2)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	numBytes := len(v.Bytes())
	if numBytes > 255 {
		return errors.New("Assertion error: numBytes>255 exeeds allowed for length prefix")
	}
	topSixBits := uint8(numBytes - 4)
	lengthByte := topSixBits<<2 + 3

	if topSixBits > 63 {
		return errors.New("Assertion error: n<=63 needed to compact-encode substrate unsigned big integer")
	}
	err := pe.PushByte(lengthByte)
	if err != nil {
		return err
	}
	buf := v.Bytes()
	Reverse(buf)
	err = pe.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

// Encode a value to the stream.
func (pe Encoder) Encode(value interface{}) error {
	t := reflect.TypeOf(value)

	// If the type implements encodeable, use that implementation
	encodeable := reflect.TypeOf((*Encodeable)(nil)).Elem()
	if t.Implements(encodeable) {
		err := value.(Encodeable).Encode(pe)
		if err != nil {
			return err
		}
		return nil
	}

	tk := t.Kind()
	switch tk {

	// Boolean and numbers are trivially encoded via binary.Write
	// It will use reflection again and take a performance hit
	// TODO: consider handling every case directly
	case reflect.Bool:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Int:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uintptr:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		err := binary.Write(pe.Writer, binary.LittleEndian, value)
		if err != nil {
			return err
		}
	case reflect.Ptr:
		rv := reflect.ValueOf(value)
		if rv.IsNil() {
			return errors.New("Encoding null pointers not supported; consider using Option type")
		} else {
			dereferenced := rv.Elem()
			err := pe.Encode(dereferenced.Interface())
			if err != nil {
				return err
			}
		}

	// Arrays: no compact-encoded length prefix
	case reflect.Array:
		rv := reflect.ValueOf(value)
		l := rv.Len()
		for i := 0; i < l; i++ {
			err := pe.Encode(rv.Index(i).Interface())
			if err != nil {
				return err
			}
		}

	// Slices: first compact-encode length, then each item individually
	case reflect.Slice:
		rv := reflect.ValueOf(value)
		l := rv.Len()
		len64 := uint64(l)
		if len64 > math.MaxUint32 {
			return errors.New("Attempted to serialize a collection with too many elements.")
		}
		err := pe.EncodeUintCompact(*big.NewInt(0).SetUint64(len64))
		if err != nil {
			return err
		}
		for i := 0; i < l; i++ {
			err = pe.Encode(rv.Index(i).Interface())
			if err != nil {
				return err
			}
		}

	// Strings are encoded as UTF-8 byte slices, just as in Rust
	case reflect.String:
		s := reflect.ValueOf(value).String()
		err := pe.Encode([]byte(s))
		if err != nil {
			return err
		}

	case reflect.Struct:
		rv := reflect.ValueOf(value)
		for i := 0; i < rv.NumField(); i++ {
			ft := rv.Type().Field(i)
			tv, ok := ft.Tag.Lookup("scale")
			if ok && tv == "-" {
				continue
			}
			err := pe.Encode(rv.Field(i).Interface())
			if err != nil {
				return fmt.Errorf("type %s does not support Encodeable interface and could not be "+
					"encoded field by field, error: %v", t, err)
			}
		}

	// Currently unsupported types
	case reflect.Complex64:
		fallthrough
	case reflect.Complex128:
		fallthrough
	case reflect.Chan:
		fallthrough
	case reflect.Func:
		fallthrough
	case reflect.Interface:
		fallthrough
	case reflect.Map:
		fallthrough
	case reflect.UnsafePointer:
		fallthrough
	case reflect.Invalid:
		return fmt.Errorf("Type %s cannot be encoded", t.Kind())
	default:
		log.Println("not captured")
	}
	return nil
}

// EncodeOption stores optionally present value to the stream.
func (pe Encoder) EncodeOption(hasValue bool, value interface{}) error {
	if !hasValue {
		err := pe.PushByte(0)
		if err != nil {
			return err
		}
	} else {
		err := pe.PushByte(1)
		if err != nil {
			return err
		}
		err = pe.Encode(value)
		if err != nil {
			return err
		}
	}
	return nil
}

// Decoder is a wraper around a Reader that allows decoding data items from a stream.
type Decoder struct {
	Reader io.Reader
}

func NewDecoder(reader io.Reader) *Decoder {
	return &Decoder{Reader: reader}
}

// Read reads bytes from a stream into a buffer
func (pd Decoder) Read(bytes []byte) error {
	c, err := pd.Reader.Read(bytes)
	if err != nil {
		return err
	}
	if c < len(bytes) {
		return fmt.Errorf("Cannot read the required number of bytes %d, only %d available", len(bytes), c)
	}
	return nil
}

// ReadOneByte reads a next byte from the stream.
// Named so to avoid a linter warning about a clash with io.ByteReader.ReadByte
func (pd Decoder) ReadOneByte() (byte, error) {
	buf := []byte{0}
	err := pd.Read(buf)
	if err != nil {
		return buf[0], err
	}
	return buf[0], nil
}

// Decode takes a pointer to a decodable value and populates it from the stream.
func (pd Decoder) Decode(target interface{}) error {
	t0 := reflect.TypeOf(target)
	if t0.Kind() != reflect.Ptr {
		return errors.New("Target must be a pointer, but was " + fmt.Sprint(t0))
	}
	val := reflect.ValueOf(target)
	if val.IsNil() {
		return errors.New("Target is a nil pointer")
	}
	return pd.DecodeIntoReflectValue(val.Elem())
}

// DecodeIntoReflectValue populates a writable reflect.Value from the stream
func (pd Decoder) DecodeIntoReflectValue(target reflect.Value) error {
	t := target.Type()
	if !target.CanSet() {
		return fmt.Errorf("Unsettable value %v", t)
	}

	// If the type implements decodeable, use that implementation
	decodeable := reflect.TypeOf((*Decodeable)(nil)).Elem()
	ptrType := reflect.PtrTo(t)
	if ptrType.Implements(decodeable) {
		var holder reflect.Value
		if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
			slice := reflect.MakeSlice(t, target.Len(), target.Len())
			holder = reflect.New(t)
			holder.Elem().Set(slice)
		} else {
			holder = reflect.New(t)
		}

		err := holder.Interface().(Decodeable).Decode(pd)
		if err != nil {
			return err
		}
		target.Set(holder.Elem())
		return nil
	}

	switch t.Kind() {

	// Boolean and numbers are trivially decoded via binary.Read
	// It will use reflection again and take a performance hit
	// TODO: consider handling every case directly
	case reflect.Bool:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Int:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uintptr:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		intHolder := reflect.New(t)
		intPointer := intHolder.Interface()
		err := binary.Read(pd.Reader, binary.LittleEndian, intPointer)
		if err == io.EOF {
			return errors.New("expected more bytes, but could not decode any more")
		}
		if err != nil {
			return err
		}
		target.Set(intHolder.Elem())

	// If you want to replicate Option<T> behavior in Rust, see OptionBool and an
	// example type OptionInt8 in tests.
	case reflect.Ptr:
		isNil := target.IsNil()
		if isNil {
			// target.set
			// return nil
		}
		ptr := target.Elem()
		err := pd.DecodeIntoReflectValue(ptr)
		if err != nil {
			return err
		}

	// Arrays: derive the length from the array length
	case reflect.Array:
		targetLen := target.Len()
		for i := 0; i < targetLen; i++ {
			err := pd.DecodeIntoReflectValue(target.Index(i))
			if err != nil {
				return err
			}
		}

	// Slices: first compact-encode length, then each item individually
	case reflect.Slice:
		codedLen64, _ := pd.DecodeUintCompact()
		if codedLen64.Uint64() > math.MaxUint32 {
			return errors.New("Encoded array length is higher than allowed by the protocol (32-bit unsigned integer)")
		}
		if codedLen64.Uint64() > uint64(maxInt) {
			return errors.New("Encoded array length is higher than allowed by the platform")
		}
		codedLen := int(codedLen64.Uint64())
		targetLen := target.Len()
		if codedLen != targetLen {
			if int(codedLen) > target.Cap() {
				newSlice := reflect.MakeSlice(t, int(codedLen), int(codedLen))
				target.Set(newSlice)
			} else {
				target.SetLen(int(codedLen))
			}
		}
		for i := 0; i < codedLen; i++ {
			err := pd.DecodeIntoReflectValue(target.Index(i))
			if err != nil {
				return err
			}
		}

	// Strings are encoded as UTF-8 byte slices, just as in Rust
	case reflect.String:
		var b []byte
		err := pd.Decode(&b)
		if err != nil {
			return err
		}
		target.SetString(string(b))

	case reflect.Struct:
		for i := 0; i < target.NumField(); i++ {
			ft := target.Type().Field(i)
			tv, ok := ft.Tag.Lookup("scale")
			if ok && tv == "-" {
				continue
			}
			err := pd.DecodeIntoReflectValue(target.Field(i))
			if err != nil {
				return fmt.Errorf("type %s does not support Decodeable interface and could not be "+
					"decoded field by field, error: %v", ptrType, err)
			}
		}

	// Currently unsupported types
	case reflect.Complex64:
		fallthrough
	case reflect.Complex128:
		fallthrough
	case reflect.Chan:
		fallthrough
	case reflect.Func:
		fallthrough
	case reflect.Interface:
		fallthrough
	case reflect.Map:
		fallthrough
	case reflect.UnsafePointer:
		fallthrough
	case reflect.Invalid:
		return fmt.Errorf("Type %s cannot be decoded", t.Kind())
	}
	return nil
}

// DecodeUintCompact decodes a compact-encoded integer. See EncodeUintCompact method.
func (pd Decoder) DecodeUintCompact() (*big.Int, error) {
	b, _ := pd.ReadOneByte()
	mode := b & 3
	switch mode {
	case 0:
		// right shift to remove mode bits
		return big.NewInt(0).SetUint64(uint64(b >> 2)), nil
	case 1:
		bb, err := pd.ReadOneByte()
		if err != nil {
			return nil, err
		}
		r := uint64(bb)
		// * 2^6
		r <<= 6
		// right shift to remove mode bits and add to prev
		r += uint64(b >> 2)
		return big.NewInt(0).SetUint64(r), nil
	case 2:
		// value = 32 bits + mode
		buf := make([]byte, 4)
		buf[0] = b
		err := pd.Read(buf[1:4])
		if err != nil {
			return nil, err
		}
		// set the buffer in little endian order
		r := binary.LittleEndian.Uint32(buf)
		// remove the last 2 mode bits
		r >>= 2
		return big.NewInt(0).SetUint64(uint64(r)), nil
	case 3:
		// remove mode bits
		l := b >> 2

		if l > 63 { // Max upper bound of 536 is (67 - 4)
			return nil, errors.New("Not supported: l>63 encountered when decoding a compact-encoded uint")
		}
		buf := make([]byte, l+4)
		err := pd.Read(buf)
		if err != nil {
			return nil, err
		}
		Reverse(buf)
		return new(big.Int).SetBytes(buf), nil
	default:
		return nil, errors.New("Code should be unreachable")
	}
}

// Reverse reverses bytes in place (manipulates the underlying array)
func Reverse(b []byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}

// DecodeOption decodes a optionally available value into a boolean presence field and a value.
func (pd Decoder) DecodeOption(hasValue *bool, valuePointer interface{}) error {
	b, _ := pd.ReadOneByte()
	switch b {
	case 0:
		*hasValue = false
	case 1:
		*hasValue = true
		err := pd.Decode(valuePointer)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unknown byte prefix for encoded OptionBool: %d", b)
	}
	return nil
}

// Encodeable is an interface that defines a custom encoding rules for a data type.
// Should be defined for structs (not pointers to them).
// See OptionBool for an example implementation.
type Encodeable interface {
	// ParityEncode encodes and write this structure into a stream
	Encode(encoder Encoder) error
}

// Decodeable is an interface that defines a custom encoding rules for a data type.
// Should be defined for pointers to structs.
// See OptionBool for an example implementation.
type Decodeable interface {
	// ParityDecode populates this structure from a stream (overwriting the current contents), return false on failure
	Decode(decoder Decoder) error
}

// OptionBool is a structure that can store a boolean or a missing value.
// Note that encoding rules are slightly different from other "Option" fields.
type OptionBool struct {
	hasValue bool
	value    bool
}

// NewOptionBoolEmpty creates an OptionBool without a value.
func NewOptionBoolEmpty() OptionBool {
	return OptionBool{false, false}
}

// NewOptionBool creates an OptionBool with a value.
func NewOptionBool(value bool) OptionBool {
	return OptionBool{true, value}
}

// ParityEncode implements encoding for OptionBool as per Rust implementation.
func (o OptionBool) Encode(encoder Encoder) error {
	var err error
	if !o.hasValue {
		err = encoder.PushByte(0)
	} else {
		if o.value {
			err = encoder.PushByte(1)
		} else {
			err = encoder.PushByte(2)
		}
	}
	if err != nil {
		return err
	}
	return nil
}

// ParityDecode implements decoding for OptionBool as per Rust implementation.
func (o *OptionBool) Decode(decoder Decoder) error {
	b, _ := decoder.ReadOneByte()
	switch b {
	case 0:
		o.hasValue = false
		o.value = false
	case 1:
		o.hasValue = true
		o.value = true
	case 2:
		o.hasValue = true
		o.value = false
	default:
		return fmt.Errorf("Unknown byte prefix for encoded OptionBool: %d", b)
	}
	return nil
}

// ToKeyedVec replicates the behaviour of Rust's to_keyed_vec helper.
func ToKeyedVec(value interface{}, prependKey []byte) ([]byte, error) {
	var buffer = bytes.NewBuffer(prependKey)
	err := Encoder{buffer}.Encode(value)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// EncodeToHex encodes `value` with the scale codec, returning a hex string (prefixed by 0x)
func EncodeToHex(value interface{}) (string, error) {
	bz, err := Encode(value)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%#x", bz), nil
}

// DecodeFromHex decodes `str` with the scale codec into `target`. `target` should be a pointer.
func DecodeFromHex(str string, target interface{}) error {
	bz, err := HexDecodeString(str)
	if err != nil {
		return err
	}
	return Decode(bz, target)
}

// Encode encodes `value` with the scale codec with passed EncoderOptions, returning []byte
func Encode(value interface{}) ([]byte, error) {
	var buffer = bytes.Buffer{}
	err := NewEncoder(&buffer).Encode(value)
	if err != nil {
		return buffer.Bytes(), err
	}
	return buffer.Bytes(), nil
}

// Decode decodes `bz` with the scale codec into `target`. `target` should be a pointer.
func Decode(bz []byte, target interface{}) error {
	return NewDecoder(bytes.NewReader(bz)).Decode(target)
}

// HexDecodeString decodes bytes from a hex string. Contrary to hex.DecodeString, this function does not error if "0x"
// is prefixed, and adds an extra 0 if the hex string has an odd length.
func HexDecodeString(s string) ([]byte, error) {
	s = strings.TrimPrefix(s, "0x")

	if len(s)%2 != 0 {
		s = "0" + s
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Hexer interface is implemented by any type that has a Hex() function returning a string
type Hexer interface {
	Hex() string
}

// EncodedLength returns the length of the value when encoded as a byte array
func EncodedLength(value interface{}) (int, error) {
	var buffer = bytes.Buffer{}
	err := NewEncoder(&buffer).Encode(value)
	if err != nil {
		return 0, err
	}
	return buffer.Len(), nil
}

// GetHash returns a hash of the value
func GetHash(value interface{}) (Hash, error) {
	enc, err := Encode(value)
	if err != nil {
		return Hash{}, err
	}
	return blake2b.Sum256(enc), err
}

// Eq compares the value of the input to see if there is a match
func Eq(one, other interface{}) bool {
	return reflect.DeepEqual(one, other)
}

// MustHexDecodeString panics if str cannot be decoded
func MustHexDecodeString(str string) []byte {
	bz, err := HexDecodeString(str)
	if err != nil {
		panic(err)
	}
	return bz
}

// HexEncodeToString encodes bytes to a hex string. Contrary to hex.EncodeToString,
// this function prefixes the hex string with "0x"
func HexEncodeToString(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}

// Hex returns a hex string representation of the value (not of the encoded value)
func Hex(value interface{}) (string, error) {
	switch v := value.(type) {
	case Hexer:
		return v.Hex(), nil
	case []byte:
		return fmt.Sprintf("%#x", v), nil
	default:
		return "", fmt.Errorf("does not support %T", v)
	}
}
