package rbmarshal

import (
	"fmt"
	"io"
)

type schemaDumper struct {
	d      *Decoder
	w      io.Writer
	indent int
}

func DebugDumpSchema(r io.Reader, w io.Writer) error {
	d := NewDecoder(r)
	// read and validate header
	major, err := d.r.ReadByte()
	if err != nil {
		return err
	}
	minor, err := d.r.ReadByte()
	if err != nil {
		return err
	}
	if major != SUPPORTED_MAJOR_VERSION || minor > SUPPORTED_MINOR_VERSION {
		return fmt.Errorf("unsupported marshal version %d.%d", major, minor)
	}

	sd := &schemaDumper{d: d, w: w}
	return sd.dumpNext()
}

func (sd *schemaDumper) writeLine(format string, a ...interface{}) {
	indent := ""
	for i := 0; i < sd.indent; i++ {
		indent += "    "
	}
	fmt.Fprintf(sd.w, "%s%s\n", indent, fmt.Sprintf(format, a...))
}

func (sd *schemaDumper) dumpNext() error {
	typ, err := sd.d.r.ReadByte()
	if err != nil {
		return err
	}

	switch typ {
	case NIL_SIGN:
		sd.writeLine("nil")
	case TRUE_SIGN:
		sd.writeLine("true")
	case FALSE_SIGN:
		sd.writeLine("false")
	case FIXNUM_SIGN:
		n := sd.d.parseInt()
		sd.writeLine("fixnum %d", n)
	case RAWSTRING_SIGN:
		s := sd.d.parseString()
		sd.writeLine("raw_string len=%d \"%s\"", len(s), s)
	case SYMBOL_SIGN:
		s := sd.d.parseString()
		sd.d.symbols = append(sd.d.symbols, s)
		sd.writeLine("symbol %q", s)
	case SYMBOL_LINK_SIGN:
		idx := sd.d.parseInt()
		name := fmt.Sprintf("#%d", idx)
		if idx >= 0 && idx < len(sd.d.symbols) {
			name = sd.d.symbols[idx]
		}
		sd.writeLine("symbol_link #%d -> %q", idx, name)
	case OBJECT_LINK_SIGN:
		// panic("not supported.")
		idx := sd.d.parseInt()
		sd.writeLine("object_link #%d", idx)
	case IVAR_SIGN:
		sd.writeLine("IVAR {")
		sd.indent++
		// inner value (string/regexp/obj)
		if err := sd.dumpNext(); err != nil {
			return err
		}
		// number of ivars
		cnt := sd.d.parseInt()
		sd.writeLine("ivar_count = %d", cnt)
		for i := 0; i < cnt; i++ {
			// ivar name (symbol)
			if err := sd.dumpNext(); err != nil {
				return err
			}
			// ivar value
			if err := sd.dumpNext(); err != nil {
				return err
			}
		}
		sd.indent--
		sd.writeLine("}")
	case ARRAY_SIGN:
		size := sd.d.parseInt()
		sd.writeLine("array size=%d [", size)
		sd.indent++
		for i := 0; i < size; i++ {
			if err := sd.dumpNext(); err != nil {
				return err
			}
		}
		sd.indent--
		sd.writeLine("]")
	case OBJECT_SIGN:
		sd.writeLine("object {")
		sd.indent++
		// class (usually symbol or class name)
		if err := sd.dumpNext(); err != nil {
			return err
		}
		// number of instance variables
		cnt := sd.d.parseInt()
		sd.writeLine("ivar_count = %d", cnt)
		for i := 0; i < cnt; i++ {
			if err := sd.dumpNext(); err != nil {
				return err
			}
			if err := sd.dumpNext(); err != nil {
				return err
			}
		}
		sd.indent--
		sd.writeLine("}")
	case HASH_SIGN:
		size := sd.d.parseInt()
		sd.writeLine("hash size=%d {", size)
		sd.indent++
		for i := 0; i < size; i++ {
			if err := sd.dumpNext(); err != nil {
				return err
			}
			if err := sd.dumpNext(); err != nil {
				return err
			}
		}
		sd.indent--
		sd.writeLine("}")
	case BIGNUM_SIGN:
		// reuse parseBignum to consume bytes
		i := sd.d.parseBignum()
		// print limited representation
		if i.BitLen() > 200 {
			sd.writeLine("bignum (%s...)", i.String()[:200])
		} else {
			sd.writeLine("bignum %s", i.String())
		}
	case REGEXP_SIGN:
		// pattern is a raw string, then an options fixnum
		pat := sd.d.parseString()
		opts := sd.d.parseInt()
		sd.writeLine("regexp pattern=%q options=%d", pat, opts)
	case CLASS_SIGN:
		sd.writeLine("class {")
		sd.indent++
		if err := sd.dumpNext(); err != nil {
			return err
		}
		sd.indent--
		sd.writeLine("}")
	case MODULE_SIGN:
		sd.writeLine("module {")
		sd.indent++
		if err := sd.dumpNext(); err != nil {
			return err
		}
		sd.indent--
		sd.writeLine("}")
	default:
		sd.writeLine("unknown byte: 0x%02x", typ)
	}
	return nil
}
