package main

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

func getPositions(file *os.File) ([]int64, error) {
	b := make([]byte, 8)

	_, err := file.Read(b)
	if err != nil {
		return nil, err
	}

	lenp := binary.BigEndian.Uint64(b)

	var positions []int64
	var i uint64

	for i = 0; i < lenp; i++ {
		_, err = file.Read(b)
		if err != nil {
			return nil, err
		}

		positions = append(positions, int64(binary.BigEndian.Uint64(b)))
	}

	return positions, nil
}

//Serve is a facade for starting HTTP server and reading previously recording files
func Serve(c *Config) error {

	it := make(map[string]int, len(c.Exporters))
	positions := make(map[string][]int64, len(c.Exporters))
	var lock = sync.RWMutex{}

	for _, exporter := range c.Exporters {
		fp, err := os.Open(exporter.Name)
		defer fp.Close()
		if err != nil {
			panic(err)
		}

		pos, err := getPositions(fp)
		if err != nil {
			panic(err)
		}

		positions[exporter.Name] = pos
		it[exporter.Name] = 0

	}

	metrics := func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()

		name := ""

		for _, exporter := range c.Exporters {

			if r.URL.String()[1:] == strings.SplitN(exporter.URL, "/", 4)[3] {
				name = exporter.Name
			}
		}

		if name == "" {
			lock.Unlock()
			w.WriteHeader(404)
			fmt.Fprintf(w, "Unable to find metric: %s", r.URL)
			return
		}

		it[name]++
		if it[name] > len(positions[name])-1 {
			it[name] = 0
		}

		response, err := readRecorderE(name, it[name], positions[name])
		if err != nil {
			lock.Unlock()
			w.WriteHeader(500)
			fmt.Fprintf(w, "Error: %s", err)
			return
		}

		fmt.Fprintf(w, "%s", response)
		lock.Unlock()
	}

	http.HandleFunc("/metrics", metrics)
	err := http.ListenAndServe(c.Bind, nil)
	if err != nil {
		return err
	}
	return nil
}

func readRecorderE(name string, entry int, positions []int64) (string, error) {
	fp, err := os.Open(name)
	defer fp.Close()

	if err != nil {
		return "", err
	}

	sum := int64(8 + len(positions)*8)
	for i := 0; i < entry; i++ {
		sum += positions[i]
	}

	fp.Seek(sum, io.SeekStart)
	entryB := make([]byte, positions[entry])
	fp.Read(entryB)

	g, err := gzip.NewReader(bytes.NewBuffer(entryB))
	defer g.Close()
	if err != nil {
		return "", err
	}

	d, err := ioutil.ReadAll(g)
	if err != nil {
		return "", err
	}

	return string(d), nil
}
