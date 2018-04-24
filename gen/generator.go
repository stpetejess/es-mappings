package gen

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"reflect"
	"strings"
	"unicode"

	"github.com/davecgh/go-spew/spew"
)

type Generator struct {
	out *bytes.Buffer

	pkgName    string
	pkgPath    string
	hashString string

	skipFmt bool

	// types that encoders were already generated for
	typesSeen map[reflect.Type]bool

	// types that encoders were requested for (e.g. by encoders of other types)
	typesUnseen []reflect.Type

	debug bool
}

// NewGenerator initializes and returns a Generator.
func NewGenerator(filename string) *Generator {
	ret := &Generator{
		typesSeen: make(map[reflect.Type]bool),
	}

	// Use a file-unique prefix on all auxiliary funcs to avoid
	// name clashes.
	hash := fnv.New32()
	hash.Write([]byte(filename))
	ret.hashString = fmt.Sprintf("%x", hash.Sum32())

	return ret
}

func (g *Generator) Debug() {
	g.debug = true
}

func (g *Generator) SkipFmt() {
	g.skipFmt = true
}

// SetPkg sets the name and path of output package.
func (g *Generator) SetPkg(name, path string) {
	g.pkgName = name
	g.pkgPath = path
}

// addTypes requests to generate encoding/decoding funcs for the given type.
func (g *Generator) addType(t reflect.Type) {
	if g.typesSeen[t] {
		return
	}
	for _, t1 := range g.typesUnseen {
		if t1 == t {
			return
		}
	}
	g.typesUnseen = append(g.typesUnseen, t)
}

// Add requests to generate marshaler/unmarshalers and encoding/decoding
// funcs for the type of given object.
func (g *Generator) Add(obj interface{}) {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	g.addType(t)
}

// Run runs the generator and outputs generated code to out.
/*
{
	"mappings": {
		"sometype": {
			"properties": {

			}
		}
	}
}
*/
func (g *Generator) Run(out io.Writer) error {
	g.out = &bytes.Buffer{}

	if g.debug {
		fmt.Fprintf(os.Stderr, "types unseen: %v", spew.Sdump(g.typesUnseen))
	}

	g.out.WriteString(`{"mappings":{`)
	for len(g.typesUnseen) > 0 {
		t := g.typesUnseen[len(g.typesUnseen)-1]
		g.typesUnseen = g.typesUnseen[:len(g.typesUnseen)-1]
		g.typesSeen[t] = true

		name := g.getName(t)
		g.out.WriteString(fmt.Sprintf(`"%s":{`, name))

		if err := g.genMappings(t); err != nil {
			return err
		}

		g.out.WriteString(`}`)
		if len(g.typesUnseen) > 0 {
			g.out.WriteString(`,`)
		}
	}
	g.out.WriteString(`}}`)
	data := &bytes.Buffer{}

	if !g.skipFmt {
		if g.debug {
			fmt.Fprintln(os.Stderr, "formatting output...")
		}
		if err := json.Indent(data, g.out.Bytes(), "", "	"); err != nil {
			return err
		}
	} else {
		data = g.out
	}
	_, err := out.Write(data.Bytes())
	return err
}

func (g *Generator) genMappings(t reflect.Type) error {
	if g.debug {
		fmt.Fprintf(os.Stderr, "generating mappings for: %v", spew.Sdump(t))
	}
	// get the tags
	nf := t.NumField()

	g.out.WriteString(`"properties":{`)

	for i := 0; i < nf; i++ {
		f := t.Field(i)
		tag := f.Tag.Get(`es-mappings`)
		if len(tag) > 0 {
			if i > 0 {
				g.out.WriteString(`,`)
			}
			parts := strings.Split(tag, `,`)
			if len(parts) < 1 {
				return errors.New("es-mappings tags must specify a name and a type. Eg: `es-mappings:\"_id,keyword\"`")
			}
			mname, mtype := parts[0], parts[1]
			if len(mname) < 1 {
				mname = camelToSnake(f.Name)
			}
			g.out.WriteString(fmt.Sprintf(`"%s":{"type":"%s"`, mname, mtype))
			// TODO: handle pointers?
			if f.Type.Kind() == reflect.Struct {
				g.out.WriteString(`,`)
				if err := g.genMappings(f.Type); err != nil {
					return err
				}
			}
			g.out.WriteString(`}`)
		}
	}
	g.out.WriteString(`}`)
	return nil
}

func (g *Generator) getName(t reflect.Type) string {
	return camelToSnake(t.Name())
}

func camelToSnake(name string) string {
	var ret bytes.Buffer

	multipleUpper := false
	var lastUpper rune
	var beforeUpper rune

	for _, c := range name {
		// Non-lowercase character after uppercase is considered to be uppercase too.
		isUpper := (unicode.IsUpper(c) || (lastUpper != 0 && !unicode.IsLower(c)))

		if lastUpper != 0 {
			// Output a delimiter if last character was either the first uppercase character
			// in a row, or the last one in a row (e.g. 'S' in "HTTPServer").
			// Do not output a delimiter at the beginning of the name.

			firstInRow := !multipleUpper
			lastInRow := !isUpper

			if ret.Len() > 0 && (firstInRow || lastInRow) && beforeUpper != '_' {
				ret.WriteByte('_')
			}
			ret.WriteRune(unicode.ToLower(lastUpper))
		}

		// Buffer uppercase char, do not output it yet as a delimiter may be required if the
		// next character is lowercase.
		if isUpper {
			multipleUpper = (lastUpper != 0)
			lastUpper = c
			continue
		}

		ret.WriteRune(c)
		lastUpper = 0
		beforeUpper = c
		multipleUpper = false
	}

	if lastUpper != 0 {
		ret.WriteRune(unicode.ToLower(lastUpper))
	}
	return string(ret.Bytes())
}
