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
	}{
		{"../test/torrent/torrent0.torrent"},
		{"../test/torrent/torrent1.torrent"},
	}

	for idx, test := range tests {

		data, err := ioutil.ReadFile(test.filename)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			binary.Write(w, binary.LittleEndian, data)
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
