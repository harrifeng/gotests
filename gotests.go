package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cweill/gotests/code"
	"github.com/cweill/gotests/render"
	"golang.org/x/tools/imports"
)

// Generates test cases and returns the number of cases generated.
func generateTestCases(testPath, src string) {
	info := code.Parse(src)
	if len(info.ExportedFuncs()) == 0 {
		return
	}
	f, err := os.Create(testPath)
	if err != nil {
		fmt.Printf("oc.Create: %v\n", err)
		return
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	if err := render.Header(w, info); err != nil {
		fmt.Printf("render.Header: %v\n", err)
		os.Remove(f.Name())
		return
	}
	var count int
	for _, fun := range info.ExportedFuncs() {
		if err := render.TestCases(w, fun); err != nil {
			fmt.Printf("render.TestCases: %v\n", err)
			continue
		}
		fmt.Printf("Generated test for %v.%v\n", info.Package, fun.Name)
		count++
	}
	if err := w.Flush(); err != nil {
		fmt.Printf("bufio.Flush: %v\n", err)
		os.Remove(f.Name())
		return
	}
	if err := processImports(f); err != nil {
		fmt.Printf("processImports: %v\n", err)
	}
	if count == 0 {
		os.Remove(f.Name())
	}
}

func processImports(f *os.File) error {
	v, err := ioutil.ReadFile(f.Name())
	if err != nil {
		return fmt.Errorf("ioutil.ReadFile: %v", err)
	}
	b, err := imports.Process(f.Name(), v, nil)
	if err != nil {
		return fmt.Errorf("imports.Process: %v\n", err)
	}
	n, err := f.WriteAt(b, 0)
	if err != nil {
		return fmt.Errorf("file.Write: %v\n", err)
	}
	if err := f.Truncate(int64(n)); err != nil {
		return fmt.Errorf("file.Truncate: %v\n", err)
	}
	return nil
}

func main() {
	for _, path := range os.Args[1:] {
		for _, src := range sourceFiles(path) {
			testPath := strings.Replace(src, ".go", "_test.go", -1)
			generateTestCases(testPath, src)
		}
	}
}

func sourceFiles(path string) []string {
	var srcs []string
	path, err := filepath.Abs(path)
	if err != nil {
		fmt.Printf("filepath.Abs: %v\n", err)
		return nil
	}
	if filepath.Ext(path) == "" {
		ps, err := filepath.Glob(path + "/*.go")
		if err != nil {
			fmt.Printf("filepath.Glob: %v\n", err)
			return nil
		}
		for _, p := range ps {
			if !isTestFile(p) {
				srcs = append(srcs, p)
			}
		}
	} else if filepath.Ext(path) == ".go" {
		if !isTestFile(path) {
			srcs = append(srcs, path)
		}
	}
	return srcs
}

func isTestFile(path string) bool {
	ok, err := filepath.Match("*_test.go", path)
	if err != nil {
		fmt.Printf("filepath.Match: %v\n", err)
		return false
	}
	if ok {
		return true
	}
	return false
}
