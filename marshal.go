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

func (d *Decoder) Read(p []byte) (int, error) {
	return 0, nil //dummy
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
			result &= ^(0xff << uint(8 * i))
			result |= int(n) << uint(8 * i)
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

type Ivar struct {
	str      string
	encoding string
}

func (d *Decoder) parseIvar() string {
	str := d.unmarshal()

	var encoding string
	var symbol string
	lengthOfSymbolChar := d.parseInt()

	if lengthOfSymbolChar == 1 {
		symbol = d.unmarshal().(string) // symbol
		value := d.unmarshal()

		d.objects = append(d.objects, value)

		if string(symbol) == "E" {
			/*if value == true {
				encoding = "utf8"
			} else {
				encoding = "ascii"
			}*/
		}
	}

	strString := str.(string)
	ivar := Ivar{strString, encoding}
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
