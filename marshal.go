package ruby_marshal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"reflect"
)

const (
	SUPPORTED_MAJOR_VERSION = 4
	SUPPORTED_MINOR_VERAION = 8
)

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: bufio.NewReader(r)}
}

type Decoder struct {
	r       *bufio.Reader
	objects []interface{}
	symbols []string
}

func (d *Decoder) unmarshal() interface{} {
	typ, _ := d.r.ReadByte()

	switch typ {
	case 0x30: // 0 - nil
		return nil
	case 0x54: // T - true
		return true
	case 0x46: // F - false
		return false
	case 0x69: // i - integer
		return d.parseInt()
	case 0x22: // " - string
		return d.parseString()
	case 0x3A: // : - symbol
		return d.parseSymbol()
	case 0x3B: // ; - symbol symlink
		return d.parseSymLink()
	case 0x40: // @ - object link
	case 0x49: // I - IVAR (encoded string or regexp)
		return d.parseIvar()
	case 0x5B: // [ - array
	case 0x6F: // o - object
	case 0x7B: // { - hash
		return d.parseHash()
	case 0x6C: // l - bignum
	case 0x2F: // / - regexp
	case 0x63: // c - class
	case 0x6D: // m -module
	default:
		return nil
	}
	panic("unsupported typecode: " + fmt.Sprintf("%#v", typ))
}

func (d *Decoder) parseInt() int {
	var result int
	b, _ := d.r.ReadByte()
	c := int(int8(b))
	if c == 0 {
		return 0
	} else if 5 < c && c < 128 {
		return c - 5
	} else if -129 < c && c < -5 {
		return c + 5
	}

	cInt8 := int8(b)
	if cInt8 > 0 {
		result = 0
		for i := int8(0); i < cInt8; i++ {
			n, _ := d.r.ReadByte()
			result |= int(uint(n) << (8 * uint(i)))
		}
	} else {
		result = -1
		c = -c
		for i := 0; i < c; i++ {
			n, _ := d.r.ReadByte()
			result &= ^(0xff << uint(8*i))
			result |= int(n) << uint(8*i)
		}
	}
	return result
}

func (d *Decoder) parseSymbol() string {
	symbol := d.parseString()
	d.symbols = append(d.symbols, symbol)
	return symbol
}

func (d *Decoder) parseSymLink() string {
	index := d.parseInt()
	return d.symbols[index]
}

func (d *Decoder) parseObjectLink() interface{} {
	index := d.parseInt()
	return d.objects[index]
}

func (d *Decoder) parseString() string {
	len := d.parseInt()
	str := make([]byte, len)
	d.r.Read(str)
	return string(str)
}

type iVar struct {
	str      string
	encoding string
}

func (d *Decoder) parseIvar() string {
	str := d.unmarshal()

	var encoding string
	var symbol interface{}
	lengthOfSymbolChar := d.parseInt()

	if lengthOfSymbolChar == 1 {
		symbol = d.unmarshal()
		value := d.unmarshal()

		d.objects = append(d.objects, value)

		if symbol.(string) == "E" {
			/*if value == true {
				encoding = "utf8"
			} else {
				encoding = "ascii"
			}*/
		}
	}

	strString := str.(string)
	ivar := iVar{strString, encoding}
	d.objects = append(d.objects, ivar)
	return strString
}

func (d *Decoder) parseHash() interface{} {
	size := d.parseInt()
	hash := make(map[string]interface{}, size)

	for i := 0; i < int(size); i++ {
		key := d.unmarshal()
		value := d.unmarshal()
		hash[key.(string)] = value
	}

	return hash
}

func (d *Decoder) Decode(v interface{}) error {
	major, err := d.r.ReadByte()
	minor, err := d.r.ReadByte()

	if err != nil {
		return errors.New("cant decode MAJOR, MINOR version")
	}

	if major != SUPPORTED_MAJOR_VERSION || minor > SUPPORTED_MINOR_VERAION {
		return errors.New("unsupported marshal version")
	}

	val := reflect.ValueOf(v)

	if val.Kind() != reflect.Ptr {
		return errors.New("pointer need.")
	}

	r := d.unmarshal()
	if r == nil {
		v = nil
		return nil
	}

	if val.Elem().Kind() == reflect.Struct {
		MapToStruct(r, v)
	} else {
		val.Elem().Set(reflect.ValueOf(r))
	}

	return nil
}

func MapToStruct(mi interface{}, o interface{}) {
	oValue := reflect.ValueOf(o).Elem()
	oType := reflect.TypeOf(o).Elem()
	m := mi.(map[string]interface{})

	for i := 0; i < oValue.NumField(); i++ {
		field := oType.Field(i)
		val := m[field.Tag.Get("ruby")]
		if val == nil {
			continue
		}

		if mm, ok := val.(map[string]interface{}); ok {
			if fieldObj := oValue.Field(i); fieldObj.Kind() == reflect.Ptr {
				if fieldObj.IsNil() {
					newObj := reflect.New(fieldObj.Type().Elem())
					fieldObj.Set(newObj)
				}
				MapToStruct(mm, fieldObj.Interface())
			} else {
				MapToStruct(mm, fieldObj.Addr().Interface())
			}
		} else {
			oValue.Field(i).Set(reflect.ValueOf(val))
		}
	}

}

type Encoder struct {
	w            *bufio.Writer
	symbols      map[string]int
	symbolsIndex int
	objects      map[*interface{}]int
	objectsIndex int
	stringObj    iVar
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:            bufio.NewWriter(w),
		symbols:      map[string]int{},
		symbolsIndex: 0,
		objects:      map[*interface{}]int{},
		objectsIndex: 0,
	}
}

func (e *Encoder) Encode(v interface{}) error {
	if _, err := e.w.Write([]byte{SUPPORTED_MAJOR_VERSION, SUPPORTED_MINOR_VERAION}); err != nil {
		return err
	}

	e.marshal(v)

	e.w.Flush()
	return nil
}

func (e *Encoder) marshal(v interface{}) error {
	vKind := reflect.TypeOf(v).Kind()
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	if vKind == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	switch typ.Kind() {
	case reflect.Bool:
		return e.encBool(val.Bool())
	case reflect.Int:
		e.w.WriteByte('i')
		return e.encInt(int(val.Int()))
	case reflect.String:
		e.w.WriteByte('I')
		return e.encString(val.String())
	}
	return nil
}

func (e *Encoder) encBool(val bool) error {
	if val {
		return e.w.WriteByte('T')
	}
	return e.w.WriteByte('F')
}

func (e *Encoder) encInt(i int) error {
	var len int

	if i == 0 {
		return e.w.WriteByte(0)
	} else if 0 < i && i < 123 {
		return e.w.WriteByte(byte(i + 5))
	} else if -124 < i && i <= -1 {
		return e.w.WriteByte(byte(i - 5))
	} else if 122 < i && i <= 0xff {
		len = 1
	} else if 0xff < i && i <= 0xffff {
		len = 2
	} else if 0xffff < i && i <= 0xffffff {
		len = 3
	} else if 0xffffff < i && i <= 0x3fffffff {
		//for compatibility with 32bit Ruby, Fixnum should be less than 1073741824
		len = 4
	} else if -0x100 <= i && i < -123 {
		len = -1
	} else if -0x10000 <= i && i < -0x100 {
		len = -2
	} else if -0x1000000 <= i && i < -0x100000 {
		len = -3
	} else if -0x40000000 <= i && i < -0x1000000 {
		//for compatibility with 32bit Ruby, Fixnum should be greater than -1073741825
		len = -4
	}

	if err := e.w.WriteByte(byte(len)); err != nil {
		return err
	}
	if len < 0 {
		len = -len
	}

	for c := 0; c < len; c++ {
		if err := e.w.WriteByte(byte(i >> uint(8*c) & 0xff)); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encRawString(str string) error {
	if err := e.encInt(len(str)); err != nil {
		return err
	}

	_, err := e.w.WriteString(str)
	return err
}

func (e *Encoder) encString(str string) error {
	if err := e.w.WriteByte('"'); err != nil {
		return err
	}
	if err := e.encRawString(str); err != nil {
		return err
	}

	//symbol :E 1個 なので、 Fixnum(1)を書き出す
	if err := e.encInt(1); err != nil {
		return err
	}

	if err := e.encSymbol("E"); err != nil {
		return err
	}
	return e.encBool(true)
}

func (e *Encoder) encSymbol(str string) error {
	if index, ok := e.symbols[str]; ok {
		if err := e.w.WriteByte(';'); err != nil {
			return err
		}
		return e.encInt(index)
	}

	e.symbols[str] = e.symbolsIndex
	e.symbolsIndex++

	if err := e.w.WriteByte(':'); err != nil {
		return err
	}
	if err := e.encInt(len(str)); err != nil {
		return err
	}
	_, err := e.w.WriteString(str)
	return err
}

func (e *Encoder) encObject() error {
	return nil
}
