package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetTorrent(t *testing.T) {

	var tests = []struct {
		filename string
		statusCode int
	}{
		{"../test/torrent/torrent0.torrent", 200},
		{"../test/torrent/torrent1.torrent", 200},
		{"../test/torrent/torrent1.torrent", 429},
		{"../test/torrent/torrent1.torrent", 404},
	}

	for idx, test := range tests {

		data, err := ioutil.ReadFile(test.filename)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if test.statusCode != 200 {
				w.WriteHeader(test.statusCode)
			} else {
				binary.Write(w, binary.LittleEndian, data)
			}
		}))

		client := server.Client()

		outfilename := strings.Replace(test.filename, ".torrent", "_save.torrent", -1)
		getTorrent(server.URL, client, outfilename)

		saveData, err := ioutil.ReadFile(outfilename)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		if !bytes.Equal(saveData, data) {
			t.Errorf("Test %v Failed:\ngot      %s\nexpected %s", idx, saveData, data)
		}
	}
}
