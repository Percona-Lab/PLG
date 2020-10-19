package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

type route struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Duration int    `json:"duration"`
	Name     string `json:"name"`
}

//Config file
type Config struct {
	Exporters []route `json:"exporters"`
	Time      int     `json:"time"`
	Bind      string  `json:"bind"`
}

func parseConfig(configName string) (*Config, error) {
	fp, err := os.Open(configName)
	if err != nil {
		return nil, err
	}

	cb, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, err
	}

	var c Config
	err = json.Unmarshal(cb, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

//Record is a facade for doRecord function
func Record(c *Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Time)*time.Second)

	var wg sync.WaitGroup

	for _, v := range c.Exporters {
		wg.Add(1)
		go doRecord(ctx, &wg, v)
	}

	wg.Wait()
	cancel()

	return nil

}

func doRecord(ctx context.Context, wg *sync.WaitGroup, r route) {
	defer wg.Done()
	ticker := time.NewTicker(time.Duration(r.Duration) * time.Second)

	tmpFile, err := ioutil.TempFile(".", "tmp*")
	if err != nil {
		panic(err)
	}

	var positions []int64

	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	var tmpBuffer bytes.Buffer
	loop := true
	for loop {
		select {
		case <-ctx.Done():
			loop = false
			break
		case <-ticker.C:
			client := &http.Client{
				Timeout: time.Duration(r.Duration)*time.Second - 100*time.Millisecond,
			}

			req, err := http.NewRequest("GET", r.URL, nil)
			if err != nil {
				panic(err)
			}
			req.SetBasicAuth(r.Username, r.Password)

			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			gz := gzip.NewWriter(&tmpBuffer)
			if _, err := gz.Write(body); err != nil {
				panic(err)
			}
			if err := gz.Close(); err != nil {
				panic(err)
			}

			positions = append(positions, int64(tmpBuffer.Len()))
			tmpFile.Write(tmpBuffer.Bytes())
			tmpBuffer.Reset()
		}
	}

	fp, err := os.Create(r.Name)
	defer fp.Close()
	if err != nil {
		panic(err)
	}

	binary.Write(fp, binary.BigEndian, int64(len(positions)))
	for _, position := range positions {
		binary.Write(fp, binary.BigEndian, position)
	}

	tmpFile.Seek(0, io.SeekStart)
	io.Copy(fp, tmpFile)
}
