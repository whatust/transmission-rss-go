package helper

import (
	"bufio"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestContainAdd(t *testing.T) {

	var seen SeenSet = SeenSet{
		Old: make(map[string]struct{}),
		New: make(map[string]struct{}),
	}

	var tests = []struct {
		input    string
		expected bool
	}{
		{"string1", false},
		{"string2", false},
		{"string1", true},
	}

	for idx, test := range tests {
		if output := seen.Contain(test.input); output != test.expected {
			t.Errorf("Test %v Failed: %v inputted, %v expected, received %v", idx, test.input, test.expected, output)
		}
		seen.AddSeen(test.input)
	}

}

func TestLoadSave(t *testing.T) {

	var seen SeenSet = SeenSet{
		Old: make(map[string]struct{}),
		New: make(map[string]struct{}),
	}

	filename := "../test/seen/load"

	err := seen.LoadSeen(filename)
	if err != nil {
		t.Errorf("Test Failed: %v", err)
	}

	var tests = []struct {
		input    string
		expected bool
	}{
		{"string1", false},
		{"string2", false},
		{"string3", true},
		{"string4", true},
		{"string5", true},
	}

	for idx, test := range tests {
		if output := seen.Contain(test.input); output != test.expected {
			t.Errorf("Test %v Failed: %v inputted, %v expected, reveived %v", idx, test.input, test.expected, output)
		}
		seen.AddSeen(test.input)
	}

	outputname := "../test/seen/save"
	copy(filename, outputname)

	seen.SaveSeen(outputname)

	file, err := os.Open(outputname)

	if err != nil {
		t.Errorf("test Failed: %v", err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if _, ok := seen.Old[scanner.Text()]; !ok {
			t.Errorf("Test Failed: Unknown entry %v on saved file", scanner.Text())
		}
	}
	file.Close()

	seen.SaveSeen(outputname)
	seen.SaveSeen(outputname)
	seen.SaveSeen(outputname)

	var seenl SeenSet = SeenSet{
		Old: make(map[string]struct{}),
		New: make(map[string]struct{}),
	}

	err = seenl.LoadSeen(outputname)
	if err != nil {
		t.Errorf("Test Failed: %v", err)
	}

	if !reflect.DeepEqual(seen.Old, seenl.Old) {
		t.Errorf("Test Failed: Loaded set and saved set are not equal")
	}
}

func copy(src, dst string) error {

	_, err := os.Stat(src)
	if err != nil {
		return err
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)

	return err
}
