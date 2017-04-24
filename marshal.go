package ruby_marshal

import (
	"bufio"
	"io"
	"errors"
	"reflect"
	"fmt"
)

const (
	SUPPORTED_MAJOR_VERSION = 4
	SUPPORTED_MINOR_VERAION = 8
)

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: bufio.NewReader(r)}
}

type Decoder struct {
	r *bufio.Reader
	objects []interface{}
	symbols []string
}

func (d *Decoder) Read(p []byte) (int, error) {
	return 0, nil //dummy
}

func (d *Decoder) unmarshal() reflect.Value {
	typ, _ := d.r.ReadByte()
	fmt.Printf("ruby type: %#v\n", typ)

	switch typ {
	case 0x30: // 0 - nil
		return reflect.ValueOf((interface{})(nil))
	case 0x54: // T - true
		return reflect.ValueOf(true)
	case 0x46: // F - false
		return reflect.ValueOf(false)
	case 0x69: // i - integer
		return reflect.ValueOf(d.parseInt())
	case 0x22: // " - string
		return reflect.ValueOf(d.parseString())
	case 0x3A: // : - symbol
		return reflect.ValueOf(d.parseSymbol())
	case 0x3B: // ; - symbol symlink
		return reflect.ValueOf(d.unmarshal())
	case 0x40: // @ - object link
	case 0x49: // I - IVAR (encoded string or regexp)
		return reflect.ValueOf(d.parseIvar())
	case 0x5B: // [ - array
	case 0x6F: // o - object
	case 0x7B: // { - hash
		return reflect.ValueOf(d.parseHash())
	case 0x6C: // l - bignum
	case 0x2F: // / - regexp
	case 0x63: // c - class
	case 0x6D: // m -module
	default:
		return reflect.ValueOf(nil)
	}
	panic("unsupported typecode: " + fmt.Sprintf("%#v", typ))
}

func (d *Decoder) parseInt() int {
	var result int
	b, _ := d.r.ReadByte()
	c := int(b)
	if c == 0 {
		return 0
	} else if 5 < c && c < 128 {
		return c - 5
	} else if -129 < c && c < -5 {
		return c + 5
	}

	if c > 0 {
		result = 0
		for i := 0; i < c; i++ {
			n, _ := d.r.ReadByte()
			result |= int(uint(n) << (8 * uint(i)))
		}
	} else {
		c = -c
		result = -1
		for i := 0; i < c; i++ {
			n, _ := d.r.ReadByte()
			result &= ^(0xff << (8 * uint(i)))
			result |= int(uint(n) << (8 * uint(i)))
		}
	}
	return result
}

func (d *Decoder) parseSymbol() string {
	symbol := d.parseString()
	d.symbols = append(d.symbols, symbol)
	return symbol
}

func (d *Decoder) parseObjectLink() interface{} {
	index := d.parseInt()
	return d.objects[index]
}

func (d *Decoder) parseString() string {
	len := d.parseInt()
	str := make([]byte, len)
	d.r.Read(str)
	fmt.Printf("str: %#v\n", str)
	return string(str)
}

type Ivar struct {
	str string
	encoding string
}

func (d *Decoder) parseIvar() string {
	str := d.unmarshal()

	var encoding string
	var symbol string
	lengthOfSymbolChar := d.parseInt()

	if lengthOfSymbolChar == 1 {
		symbol = d.unmarshal().String() // symbol
		//value := d.unmarshal().Bool() // value
		fmt.Println("value unmarshal")
		value := d.unmarshal()
		fmt.Printf("value: %#v\n", value)

		d.objects = append(d.objects, value)

		if string(symbol) == "E" {
			/*if value == true {
				encoding = "utf8"
			} else {
				encoding = "ascii"
			}*/
		}
	}

	strString := str.String()
	ivar := Ivar{strString, encoding}
	d.objects = append(d.objects, ivar)
	return strString
}

func (d *Decoder) parseHash() interface{} {
	size := d.parseInt()
	hash := make(map[string]reflect.Value, size)

	for i := 0; i < int(size); i++ {
		key := d.unmarshal()
		value := d.unmarshal()
		hash[key.String()] = value
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
	fmt.Printf("r: %#v\n", r)
	if !r.IsValid() {
		v = nil
		return nil
	}
	val.Elem().Set(r)
	return nil
}
