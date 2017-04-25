package ruby_marshal

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestNewDecoder(t *testing.T) {
	testSet := map[string]func(interface{}){
		"null": func(v interface{}) {
			if v != nil {
				t.Error("null is should nil")
			}
		},
		"int_1": func(v interface{}) {
			i := v.(int)
			if i != 1 {
				t.Error("int_1 should int 1")
			}
		},
		"int_777": func(v interface{}) {
			i := v.(int)
			if i != 777 {
				t.Error("int_777 should int 777")
			}
		},
		"int_-777": func(v interface{}) {
			i := v.(int)
			if i != -777 {
				t.Error("int_-777 should int -777")
			}
		},
		"int_65537": func(v interface{}) {
			i := v.(int)
			if i != 65537 {
				t.Error("int_65537 should int 65537")
			}
		},
		"int_-65537": func(v interface{}) {
			i := v.(int)
			if i != -65537 {
				t.Error("int_-65537 should int -65537")
			}
		},
		"int_0": func(v interface{}) {
			i := v.(int)
			if i != 0 {
				t.Error("int_0 should int 0")
			}
		},
		"int_-5": func(v interface{}) {
			i := v.(int)
			if i != -5 {
				t.Error("int_-5 should int -5")
			}
		},
		"sym_name": func(v interface{}) {
			i := v.(string)
			if i != "name" {
				t.Error("sym_name should string name")
			}
		},
		"string_hoge": func(v interface{}) {
			i := v.(string)
			if i != "hoge" {
				t.Error("string_hoge should string hoge")
			}
		},
		"hash_1": func(v interface{}) {
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

		var v interface{}
		//t.Logf("Initial: %#v\n", v)
		if e := NewDecoder(fp).Decode(&v); e != nil {
			t.Error(e.Error())
		}
		t.Logf("file: %#v\ttype: %#T\tvalue: %#v\n", file, v, v)
	}
}

func TestDecodeInt1(t *testing.T) {
	file := "int_1"
	fp, e := os.Open("test_set/" + file + ".dat")
	defer func() {
		fp.Close()
	}()
	if e != nil {
		panic(e.Error())
	}
	var i int
	if e := NewDecoder(fp).Decode(&i); e != nil {
		t.Error(e.Error())
	}
	t.Logf("Type: %#T\t Value: %#v\n", i, i)
}

func TestDecodeHash(t *testing.T) {
	file := "session"
	fp, e := os.Open("test_set/" + file + ".dat")
	defer func() {
		fp.Close()
	}()
	if e != nil {
		panic(e.Error())
	}

	var v Session
	//t.Logf("Initial: %#v\n", v)
	if e := NewDecoder(fp).Decode(&v); e != nil {
		t.Error(e.Error())
	}

	t.Logf("%#v\n", v)
}

func TestMapToStruct(t *testing.T) {
	m := interface{}(map[string]interface{}{
		"name": "jack",
		"age":  21,
	})
	u := User{}
	MapToStruct(m, &u)
	t.Logf("Map: %#v\n", m)
	t.Logf("Map: %#v\n", u)

	i := 1
	vi := reflect.ValueOf(&i)
	vvi := reflect.ValueOf(&vi)
	//t.Logf("%#v\n%#v\n%#v\n", i, vi, vvi)

	t.Logf("%#v\n%#v\n%#v\n",
		i,
		vi,
		vvi.Elem(),
	)
}

func TestRedisConfigUnMarshal(t *testing.T) {
	file := "hash_1"
	fp, e := os.Open("test_set/" + file + ".dat")
	defer func() {
		fp.Close()
	}()

	conf := RedisConfig{}
	if e = NewDecoder(fp).Decode(&conf); e != nil {
		t.Error(e.Error())
	}
	t.Logf("%#v\n", conf)
}

func TestSessionMarshal(t *testing.T) {
	file := "session"
	fp, e := os.Open("test_set/" + file + ".dat")
	if e != nil {
		panic(e.Error())
	}
	defer func() {
		fp.Close()
	}()

	//session := new(Session)
	//session := &Session{User: &User{}}
	session := Session{}
	if e = NewDecoder(fp).Decode(&session); e != nil {
		t.Error(e.Error())
	}
	t.Logf("%#v, %#v\n", session, session.User)
}

type RedisConfig struct {
	Host string `ruby:"host"`
	DB   int    `ruby:"db"`
}

type Session struct {
	User  User   `ruby:"user"`
	Flash string `ruby:"flash"`
}

type User struct {
	Name string `ruby:"name"`
	Age  int    `ruby:"age"`
}
