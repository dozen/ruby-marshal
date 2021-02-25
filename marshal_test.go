package rbmarshal

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
)

const (
	// null
	Null = "040830"
	// "hoge"
	String = "0408492209686f6765063a064554"
	// :name
	SymName = "04083a096e616d65"
	// 0
	Int0 = "04086900"
	// 1
	Int1 = "04086906"
	// 2
	Int2 = "04086907"
	// -5
	IntM5 = "040869f6"
	// 777
	Int777 = "040869020903"
	// -777
	IntM777 = "040869fef7fc"
	// 65537
	Int65537 = "04086903010001"
	// -65537
	IntM65537 = "040869fdfffffe"
	// { host: "localhost", db: 1 }
	Hash1 = "04087b073a09686f737449220e6c6f63616c686f7374063a0645543a0764626906"
	// { "name" => "taro", "age" => 21 }
	Hash2 = "04087b074922096e616d65063a0645544922097461726f063b0054492208616765063b0054691a"
	// { user: { name: "matsumoto-yasunori", age: 57 }, job: "voice-actor" }
	Hash3 = "04087b073a09757365727b073a096e616d654922176d617473756d6f746f2d796173756e6f7269063a0645543a08616765693e3a086a6f62492210766f6963652d6163746f72063b0754"
)

var (
	// 1612874507
	PosBignum = []byte{0x04, 0x08, 0x6c, 0x2b, 0x07, 0x0b, 0x83, 0x22, 0x60}
	// -15241578750190521
	NegBignum = []byte{0x04, 0x08, 0x6c, 0x2d, 0x09, 0xb9, 0xa3, 0x38, 0x97, 0x22, 0x26, 0x36, 0x00}
)

func TestDecodeNull(t *testing.T) {
	b, err := hex.DecodeString(Null)
	if err != nil {
		t.Skip(err.Error())
	}
	var v interface{}
	NewDecoder(bytes.NewReader(b)).Decode(&v)
	if v != nil {
		t.Errorf("not nil. Type: %T\tValue: %#v", v, v)
	}
}

func TestDecodeString(t *testing.T) {
	b, err := hex.DecodeString(String)
	if err != nil {
		t.Skip(err.Error())
	}
	var v string
	NewDecoder(bytes.NewReader(b)).Decode(&v)
	if v != "hoge" {
		t.Errorf("not \"hoge\". Type: %T\tValue: %#v", v, v)
	}
}

func TestDecodeSymName(t *testing.T) {
	b, err := hex.DecodeString(SymName)
	if err != nil {
		t.Skip(err.Error())
	}
	var v string
	NewDecoder(bytes.NewReader(b)).Decode(&v)
	if v != "name" {
		t.Errorf("not \"name\". Type: %T\tValue: %#v", v, v)
	}
}

func TestDecodeInt0(t *testing.T) {
	b, err := hex.DecodeString(Int0)
	if err != nil {
		t.Skip(err.Error())
	}
	var v int
	NewDecoder(bytes.NewReader(b)).Decode(&v)
	if v != 0 {
		t.Errorf("not 0. Type: %T\tValue: %#v", v, v)
	}
}

func TestDecodeInt1(t *testing.T) {
	b, err := hex.DecodeString(Int1)
	if err != nil {
		t.Skip(err.Error())
	}
	var v int
	NewDecoder(bytes.NewReader(b)).Decode(&v)
	if v != 1 {
		t.Errorf("not 1. Type: %T\tValue: %#v", v, v)
	}
}

func TestDecodeInt2(t *testing.T) {
	b, err := hex.DecodeString(Int2)
	if err != nil {
		t.Skip(err.Error())
	}
	var v int
	NewDecoder(bytes.NewReader(b)).Decode(&v)
	if v != 2 {
		t.Errorf("not 2. Type: %T\tValue: %#v", v, v)
	}
}

func TestDecodeIntM5(t *testing.T) {
	b, err := hex.DecodeString(IntM5)
	if err != nil {
		t.Skip(err.Error())
	}
	var v int
	NewDecoder(bytes.NewReader(b)).Decode(&v)
	if v != -5 {
		t.Errorf("not 5. Type: %T\tValue: %#v", v, v)
	}
}

func TestDecodeInt777(t *testing.T) {
	b, err := hex.DecodeString(Int777)
	if err != nil {
		t.Skip(err.Error())
	}
	var v int
	NewDecoder(bytes.NewReader(b)).Decode(&v)
	if v != 777 {
		t.Errorf("not 777. Type: %T\tValue: %#v", v, v)
	}
}

func TestDecodeIntM777(t *testing.T) {
	b, err := hex.DecodeString(IntM777)
	if err != nil {
		t.Skip(err.Error())
	}
	var v int
	NewDecoder(bytes.NewReader(b)).Decode(&v)
	if v != -777 {
		t.Errorf("not -777. Type: %T\tValue: %#v", v, v)
	}
}

func TestDecodeInt65537(t *testing.T) {
	b, err := hex.DecodeString(Int65537)
	if err != nil {
		t.Skip(err.Error())
	}
	var v int
	NewDecoder(bytes.NewReader(b)).Decode(&v)
	if v != 65537 {
		t.Errorf("not 0. Type: %T\tValue: %#v", v, v)
	}
}

func TestDecodeIntM65537(t *testing.T) {
	b, err := hex.DecodeString(IntM65537)
	if err != nil {
		t.Skip(err.Error())
	}
	var v int
	NewDecoder(bytes.NewReader(b)).Decode(&v)
	if v != -65537 {
		t.Errorf("not -65537. Type: %T\tValue: %#v", v, v)
	}
}

func TestDecodeHash1(t *testing.T) {
	b, err := hex.DecodeString(Hash1)
	if err != nil {
		t.Skip(err.Error())
	}
	var v RedisConf
	NewDecoder(bytes.NewReader(b)).Decode(&v)
	if !(v.Host == "localhost" && v.DB == 1) {
		t.Errorf("not matched. Value: %#v", v)
	}
}

type RedisConf struct {
	Host string `ruby:"host"`
	DB   int    `ruby:"db"`
}

func TestDecodeHash2(t *testing.T) {
	b, err := hex.DecodeString(Hash2)
	if err != nil {
		t.Skip(err.Error())
	}
	var v User
	NewDecoder(bytes.NewReader(b)).Decode(&v)
	if !(v.Name == "taro" && v.Age == 21) {
		t.Errorf("not matched. Value: %#v", v)
	}
}

func TestDecodeHash3(t *testing.T) {
	b, err := hex.DecodeString(Hash3)
	if err != nil {
		t.Skip(err.Error())
	}
	var v Profile
	NewDecoder(bytes.NewReader(b)).Decode(&v)

	if !(v.Job == "voice-actor" && v.User.Name == "matsumoto-yasunori" && v.User.Age == 57) {
		t.Errorf("not matched. Value: %#v", v)
	}
}

func TestDecodeBignum(t *testing.T) {
	var v *big.Int
	NewDecoder(bytes.NewReader(PosBignum)).Decode(&v)

	want := big.NewInt(1612874507)
	if v.Cmp(want) != 0 {
		t.Errorf("not matched. Value: %#v", v)
	}
}

func TestDecodeNegativeBignum(t *testing.T) {
	var v *big.Int
	NewDecoder(bytes.NewReader(NegBignum)).Decode(&v)

	want := big.NewInt(-15241578750190521)
	if v.Cmp(want) != 0 {
		t.Errorf("not matched. Value: %#v", v)
	}
}

type Profile struct {
	User User   `ruby:"user"`
	Job  string `ruby:"job"`
}

type User struct {
	Name string `ruby:"name"`
	Age  int    `ruby:"age"`
}

func TestNewEncoder(t *testing.T) {
	w := bytes.NewBuffer([]byte{})
	e := NewEncoder(w)
	v := "めっちゃ日本語"
	if err := e.Encode(&v); err != nil {
		t.Error(err.Error())
	}

	encoded := w.Bytes()
	fmt.Printf("encoded:\t%#v\n", encoded)
	fmt.Printf("encoded:\t%x\n", encoded)

	var str string
	NewDecoder(bytes.NewReader(encoded)).Decode(&str)
	fmt.Printf("%#x\n%#v\n", str, str)
}
