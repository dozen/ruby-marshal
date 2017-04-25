# ruby-marshal

Parse Ruby Marshal Binary Data.

```
go get github.com/dozen/ruby-marshal
```

## Example

```
irb(main):001:0> Marshal.dump({ user: { name: "matsumoto-yasunori", age: 57 }, job: "voice-actor" }).unpack("H*")
=> ["04087b073a09757365727b073a096e616d654922176d617473756d6f746f2d796173756e6f7269063a0645543a08616765693e3a086a6f62492210766f6963652d6163746f72063b0754"]
```

```
b, _ := hex.DecodeString("04087b073a..(omit)..f72063b0754")
var v Profile
NewDecoder(bytes.NewReader(b)).Decode(&v)
#=> Profile{User:User{ Name:"matsumoto-yasunori", Age:57 }, Job:"voice-actor" }
```

## Usage

Follow [marshal_test.go](marshal_test.go).

## Reference

* [rainerborene/rum](https://github.com/rainerborene/rum)
* [instore/node-marshal](https://github.com/instore/node-marshal)
