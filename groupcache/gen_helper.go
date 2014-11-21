// Copyright 2014 Gopkg. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ingore

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	flagRevert = flag.Bool("revert", false, "revert all changes")
)

var (
	targetName    = `groupcache`
	targetMaster  = `master`
	targetVersion = `d2a1805917c4ba5f227d5b1f9a871dd274a07542`
	targeFilename = `zz-groupcache-` + targetVersion[:12] + `.tar.gz`
	targetURL     = `http://github.com/golang/groupcache/archive/` + targetVersion + `.tar.gz`
)

var convertMap = [][2]string{
	[2]string{
		`"github.com/golang/groupcache`,
		`"github.com/gopkg/cache/groupcache`,
	},
	[2]string{
		`"code.google.com/p/goprotobuf/proto`,
		`"github.com/gopkg/cache/groupcache/proto`,
	},
}

func main() {
	flag.Parse()

	// try download target
	if !isValidTarGzFile(targeFilename) {
		os.Remove(targeFilename)
		if err := downloadTarGzFile(targetURL, targeFilename); err != nil {
			log.Fatalf("download %s failed, err = %v", targetName, err)
		}
	}

	unpackSourceCode(targeFilename)
	fixAllImportPath(".")
	saveVersion()

	// Done
	fmt.Println("Done")
}

func isValidTarGzFile(filename string) bool {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return false
	}
	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return false
	}
	defer gzReader.Close()

	trReader := tar.NewReader(gzReader)
	for {
		if _, err := trReader.Next(); err != nil {
			if err != io.EOF {
				return false
			}
			break
		}
	}

	return true
}
func downloadTarGzFile(url, filename string) (err error) {
	defer func() {
		if err != nil {
			os.Remove(filename)
		}
	}()

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(f, resp.Body)
	return
}

func unpackSourceCode(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("unpackSourceCode: ioutil.ReadFile filed, err = %v", err)
	}

	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		log.Fatalf("unpackSourceCode: gzip.NewReader filed, err = %v", err)
	}
	defer gzReader.Close()

	trReader := tar.NewReader(gzReader)
	for {
		header, err := trReader.Next()
		if err != nil {
			if err != io.EOF {
				log.Fatalf("unpackSourceCode: trReader.Next filed, err = %v", err)
			}
			break
		}

		// skip other files
		if header.FileInfo().IsDir() {
			continue
		}

		// unpack files
		name := header.Name[strings.Index(header.Name, "/")+1:]
		if name == "pax_global_header" {
			continue
		}
		os.MkdirAll(path.Dir(name), 0666)
		fw, err := os.Create(name)
		if err != nil {
			log.Fatalf("unpackSourceCode: os.Create filed, err = %v", err)
		}
		defer fw.Close()

		_, err = io.Copy(fw, trReader)
		if err != nil {
			log.Fatalf("unpackSourceCode: io.Copy filed, err = %v", err)
		}
	}
}

func fixAllImportPath(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal("filepath.Walk: ", err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, "gen.go") {
			return nil
		}
		if strings.HasSuffix(path, "gen_helper.go") {
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			fixImportPath(path)
		}
		return nil
	})
}

func fixImportPath(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("ioutil.ReadFile: ", err)
	}

	for _, v := range convertMap {
		oldPath, newPath := v[0], v[1]
		if !*flagRevert {
			data = bytes.Replace(data, []byte(oldPath), []byte(newPath), -1)
		} else {
			data = bytes.Replace(data, []byte(newPath), []byte(oldPath), -1)
		}
	}

	if err = ioutil.WriteFile(filename, data, 0666); err != nil {
		log.Fatal("ioutil.WriteFile: ", err)
	}

	if !*flagRevert {
		fmt.Printf("convert %s ok\n", filename)
	} else {
		fmt.Printf("revert %s ok\n", filename)
	}
}

func saveVersion() {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s: %s\n", targetName, targetURL)
	if err := ioutil.WriteFile(targetName+"-version.txt", buf.Bytes(), 0666); err != nil {
		log.Fatal("ioutil.WriteFile: ", err)
	}
}
