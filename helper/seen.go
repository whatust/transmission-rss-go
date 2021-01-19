package helper

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

// SeenTorrent ...
type SeenTorrent interface {
	LoadSeen(string) error
	SaveSeen(string) error
	Contain(string) bool
	AddSeen(string)
}

// SeenSet ...
type SeenSet struct {
	mu sync.RWMutex
	New map[string]struct{}
	Old map[string]struct{}
}

// LoadSeen ...
func (set *SeenSet) LoadSeen(fileName string) error {

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var aux struct{}
	for scanner.Scan() {
		set.Old[scanner.Text()] = aux
	}

	return nil
}

// SaveSeen ...
func (set *SeenSet) SaveSeen(fileName string) error {

	fmt.Println("Saving seen torrents...")

	file, err := os.OpenFile(
		fileName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)

	if err != nil {
		return err
	}
	defer file.Close()

	for k := range set.New {
		_, err := fmt.Fprintln(file, k)
		if err != nil {
			return err
		}
	}

	return nil
}

// Contain ...
func (set *SeenSet) Contain (uID string) bool {

	_, in := set.Old[uID]
	if in {
		return true
	}

	set.mu.RLock()
	_, in = set.New[uID]
	set.mu.RUnlock()

	return in
}

// AddSeen ...
func (set *SeenSet) AddSeen (uID string) {

	var aux struct{}

	if !set.Contain(uID) {
		set.mu.Lock()
		set.New[uID] = aux
		set.mu.Unlock()
	}
}
