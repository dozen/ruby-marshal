package ruby_marshal

import (
	"bufio"
	"errors"
	"io"
	"reflect"
)

const (
	SUPPORTED_MAJOR_VERSION = 4
	SUPPORTED_MINOR_VERSION = 8

	NIL_SIGN         = '0'
	TRUE_SIGN        = 'T'
	FALSE_SIGN       = 'F'
	FIXNUM_SIGN      = 'i'
	RAWSTRING_SIGN   = '"'
	SYMBOL_SIGN      = ':'
	SYMBOL_LINK_SIGN = ';'
	OBJECT_SIGN      = 'o'
	OBJECT_LINK_SIGN = '@'
	ARRAY_SIGN       = '['
	IVAR_SIGN        = 'I'
	HASH_SIGN        = '{'
	BIGNUM_SIGN      = 'l'
	REGEXP_SIGN      = '/'
	CLASS_SIGN       = 'c'
	MODULE_SIGN      = 'm'
)

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: bufio.NewReader(r)}
}

type Decoder struct {
	r       *bufio.Reader
	objects []*reflect.Value
	symbols []string
}

func (d *Decoder) unmarshal() *reflect.Value {
	typ, _ := d.r.ReadByte()

	var v reflect.Value

	switch typ {
	case NIL_SIGN: // 0 - nil
		v = reflect.ValueOf(nil)
	case TRUE_SIGN: // T - true
		v = reflect.ValueOf(true)
	case FALSE_SIGN: // F - false
		v = reflect.ValueOf(false)
	case FIXNUM_SIGN: // i - integer
		d.parseInt(&v)
	case RAWSTRING_SIGN: // " - string
		d.parseString(&v)
	case SYMBOL_SIGN: // : - symbol
		d.parseSymbol(&v)
	case SYMBOL_LINK_SIGN: // ; - symbol symlink
		d.parseSymLink(&v)
	case OBJECT_LINK_SIGN: // @ - object link
		panic("not supported.")
	case IVAR_SIGN: // I - IVAR (encoded string or regexp)
		d.parseIvar(&v)
	case ARRAY_SIGN: // [ - array
		panic("not supported.")
	case OBJECT_SIGN: // o - object
		panic("not supported.")
	case HASH_SIGN: // { - hash
		d.parseHash(&v)
	case BIGNUM_SIGN: // l - bignum
		panic("not supported.")
	case REGEXP_SIGN: // / - regexp
		panic("not supported.")
	case CLASS_SIGN: // c - class
		panic("not supported.")
	case MODULE_SIGN: // m -module
		panic("not supported.")
	default:
		return nil
	}
	return &v
}

func (d *Decoder) parseInt(v *reflect.Value) {
	var result int
	b, _ := d.r.ReadByte()
	c := int(int8(b))
	if c == 0 {
		result = 0
	} else if 5 < c && c < 128 {
		result = c - 5
	} else if -129 < c && c < -5 {
		result = c + 5
	} else {
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
	}

	*v = reflect.ValueOf(&result)
}

func (d *Decoder) parseSymbol(v *reflect.Value) {
	d.parseString(v)
	d.symbols = append(d.symbols, v.Elem().String())
}

func (d *Decoder) parseSymLink(v *reflect.Value) {
	index := reflect.ValueOf(new(int))
	d.parseInt(&index)
	v.Elem().SetString(d.symbols[index.Elem().Int()])
}

func (d *Decoder) parseObjectLink(v *reflect.Value) {
	index := reflect.ValueOf(new(int))
	d.parseInt(&index)
	v = d.objects[index.Elem().Int()]
}

func (d *Decoder) parseString(v *reflect.Value) {
	var len reflect.Value
	d.parseInt(&len)
	b := make([]byte, len.Elem().Int())
	d.r.Read(b)
	str := string(b)
	*v = reflect.ValueOf(&str)
}

type iVar struct {
	str string
}

func (d *Decoder) parseIvar(v *reflect.Value) {
	*v = *d.unmarshal()

	var varLen reflect.Value
	d.parseInt(&varLen)

	if varLen.Elem().Int() == 1 {
		symbol := d.unmarshal() // :E
		_ = d.unmarshal()       // T
		d.symbols = append(d.symbols, symbol.Elem().String())
	}

	ivar := reflect.ValueOf(&iVar{v.Elem().String()})
	d.objects = append(d.objects, &ivar)
}

func (d *Decoder) parseHash(v *reflect.Value) {
	var sizeV reflect.Value
	d.parseInt(&sizeV)
	size := int(sizeV.Elem().Int())
	hash := make(map[string]interface{}, size)

	for i := 0; i < size; i++ {
		key := d.unmarshal()
		value := d.unmarshal()
		hash[key.String()] = value
	}

	hashV := reflect.ValueOf(hash)
	v = &hashV
}

func (d *Decoder) Decode(v interface{}) error {
	major, err := d.r.ReadByte()
	minor, err := d.r.ReadByte()

	if err != nil {
		return errors.New("cant decode MAJOR, MINOR version")
	}
	if major != SUPPORTED_MAJOR_VERSION || minor > SUPPORTED_MINOR_VERSION {
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
	} else if *r == reflect.ValueOf(nil) {
		v = nil
	} else {
		val.Elem().Set(r.Elem())
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
	if _, err := e.w.Write([]byte{SUPPORTED_MAJOR_VERSION, SUPPORTED_MINOR_VERSION}); err != nil {
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
		e.w.WriteByte(FIXNUM_SIGN)
		return e.encInt(int(val.Int()))
	case reflect.String:
		e.w.WriteByte(IVAR_SIGN)
		return e.encString(val.String())
	}
	return nil
}

func (e *Encoder) encBool(val bool) error {
	if val {
		return e.w.WriteByte(TRUE_SIGN)
	}
	return e.w.WriteByte(FALSE_SIGN)
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

func (e *Encoder) _encRawString(str string) error {
	// | len (Fixnum) | stirng |
	if err := e.encInt(len(str)); err != nil {
		return err
	}

	_, err := e.w.WriteString(str)
	return err
}

func (e *Encoder) encString(str string) error {
	// | I | " | RawString( string ) | FixNum( 1 ) | Symbol( E ) | True |
	if _, err := e.w.Write([]byte{IVAR_SIGN, RAWSTRING_SIGN}); err != nil {
		return err
	}
	if err := e._encRawString(str); err != nil {
		return err
	}
	if err := e.encInt(1); err != nil {
		return err
	}
	if err := e._encSymbol("E"); err != nil {
		return err
	}
	return e.encBool(true)
}

func (e *Encoder) _encSymbol(str string) error {
	if index, ok := e.symbols[str]; ok {
		if err := e.w.WriteByte(SYMBOL_LINK_SIGN); err != nil {
			return err
		}
		return e.encInt(index)
	}

	e.symbols[str] = e.symbolsIndex
	e.symbolsIndex++

	if err := e.w.WriteByte(SYMBOL_SIGN); err != nil {
		return err
	}
	if err := e.encInt(len(str)); err != nil {
		return err
	}
	_, err := e.w.WriteString(str)
	return err
}

func (e *Encoder) _encObject() error {
	return nil
}
