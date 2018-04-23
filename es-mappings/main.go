package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bringhub/ebay_api/cmd/es-mapper/bootstrap"
	"github.com/bringhub/ebay_api/cmd/es-mapper/parser"
)

var allStructs = flag.Bool("all", false, "generate marshaler/unmarshalers for all structs in a file")
var specifiedName = flag.String("output_filename", "", "specify the filename of the output")
var processPkg = flag.Bool("pkg", false, "process the whole package instead of just the given file")
var skipFmt = flag.Bool("nofmt", false, "skip final formatting of the json file")

func generate(fname string) (err error) {
	fInfo, err := os.Stat(fname)
	if err != nil {
		return err
	}

	p := parser.Parser{AllStructs: *allStructs}
	if err := p.Parse(fname, fInfo.IsDir()); err != nil {
		return fmt.Errorf("Error parsing %v: %v", fname, err)
	}

	var outName string
	if fInfo.IsDir() {
		outName = filepath.Join(fname, p.PkgName+"_es_mapper.json")
	} else {
		if s := strings.TrimSuffix(fname, ".go"); s == fname {
			return errors.New("Filename must end in '.go'")
		} else {
			outName = s + "_es_mapper.json"
		}
	}

	if *specifiedName != "" {
		outName = *specifiedName
	}
	// fmt.Fprintf(os.Stderr, "struct names: %v", p.StructNames)
	g := bootstrap.Generator{
		PkgPath:    p.PkgPath,
		PkgName:    p.PkgName,
		Types:      p.StructNames,
		OutName:    outName,
		GoPath:     os.Getenv("GOPATH"),
		SkipFormat: skipFmt,
	}

	if err := g.Run(); err != nil {
		return fmt.Errorf("Bootstrap failed: %v", err)
	}
	return nil
}

func main() {
	flag.Parse()

	files := flag.Args()

	gofile := os.Getenv("GOFILE")
	if *processPkg {
		gofile = filepath.Dir(gofile)
	}

	if len(files) == 0 && gofile != "" {
		files = []string{gofile}
	} else if len(files) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	for _, fname := range files {
		if err := generate(fname); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
