package ruby_marshal

import (
	"testing"
	"fmt"
	"os"
)

func TestNewDecoder(t *testing.T) {
	testSet := map[string] func(interface{}) {
		"null": func (v interface{}) {
			if v != nil {
				t.Error("null is should nil")
			}
		},
		"int_1": func (v interface{}) {
			i := v.(int)
			if i != 1 {
				t.Error("int_1 should int 1")
			}
		},
		"int_777": func (v interface{}) {
			i := v.(int)
			if i != 777 {
				t.Error("int_777 should int 777")
			}
		},
		"int_-777": func (v interface{}) {
			i := v.(int)
			if i != -777 {
				t.Error("int_-777 should int -777")
			}
		},
		"int_65537": func (v interface{}) {
			i := v.(int)
			if i != 65537 {
				t.Error("int_65537 should int 65537")
			}
		},
		"int_-65537": func (v interface{}) {
			i := v.(int)
			if i != -65537 {
				t.Error("int_-65537 should int -65537")
			}
		},
		"int_0": func (v interface{}) {
			i := v.(int)
			if i != 0 {
				t.Error("int_0 should int 0")
			}
		},
		"int_-5": func (v interface{}) {
			i := v.(int)
			if i != -5 {
				t.Error("int_-5 should int -5")
			}
		},
		"sym_name": func (v interface{}) {
			i := v.(string)
			if i != "name" {
				t.Error("sym_name should string name")
			}
		},
		"string_hoge": func (v interface{}) {
			i := v.(string)
			if i != "hoge" {
				t.Error("string_hoge should string hoge")
			}
		},
		"hash_1": func (v interface{}) {
			i := v.(string)
			if i != "hoge" {
				t.Error("string_hoge should string hoge")
			}
		},

	}

	for file, eval := range testSet {
		_ = eval
		fmt.Println("\n" + file)
		fp, e := os.Open("test_set/" + file + ".dat")
		defer func() {
			fp.Close()
		}()
		if e != nil {
			panic(e.Error())
		}

		v, e := NewDecoder(fp).Decode()
		if e != nil {
			t.Error(e.Error())
		}
		t.Logf("file: %#v\ttype: %#T\tvalue: %#v\n", file, v, v)
	}
}

func TestDecodeHash(t *testing.T) {
	file := "hash_3"
	fp, e := os.Open("test_set/" + file + ".dat")
	defer func() {
		fp.Close()
	}()
	if e != nil {
		panic(e.Error())
	}

	//t.Logf("Initial: %#v\n", v)
	v, e := NewDecoder(fp).Decode()
	if e != nil {
		t.Error(e.Error())
	}



	t.Logf("%#T\n%#v\n", v, v)

}