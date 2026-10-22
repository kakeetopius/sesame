package db

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/kakeetopius/sesame/internal/ip"
)

// fileType is the main struct for file database
type fileType struct {
	path string
	lock sync.Mutex
}

type ipStore map[string]*ip.IPAddress

func GetFileTypeInstance(path string) (Instance, error) {
	err := os.MkdirAll(filepath.Dir(path), 0o660)
	if err != nil {
		return nil, fmt.Errorf("error opening server database: %w", err)
	}

	_, err = os.OpenFile(path, os.O_RDWR, 0o666)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("error opening server database: %w", err)
		}
	}

	return &fileType{
		path: path,
	}, nil
}

// Type of the db
func (fileType *fileType) Type() string {
	return "fileType"
}

// GetIP from the db
func (fileType *fileType) GetIP(ipAddr string) (*ip.IPAddress, error) {
	ipStore, err := fileType.loadIPStore()
	if err != nil {
		return nil, err
	}
	ip, ok := ipStore[ipAddr]
	if !ok {
		return nil, fmt.Errorf("ip not found")
	}
	return ip, nil
}

// GetAllIPs from the db
func (fileType *fileType) GetAllIPs() ([]*ip.IPAddress, error) {
	ips := []*ip.IPAddress{}
	ipStore, err := fileType.loadIPStore()
	if err != nil {
		return nil, err
	}
	for _, ip := range ipStore {
		ips = append(ips, ip)
	}
	return ips, nil
}

// CheckIPExists checks if IP is in db
func (fileType *fileType) CheckIPExists(ipAddr string) (bool, error) {
	ipStore, err := fileType.loadIPStore()
	if err != nil {
		return false, err
	}
	_, ok := ipStore[ipAddr]
	if ok {
		return true, nil
	}
	return false, nil
}

// AddIP to the db
func (fileType *fileType) AddIP(ip *ip.IPAddress) error {
	ipStore, err := fileType.loadIPStore()
	if err != nil {
		return err
	}
	_, ok := ipStore[ip.IP]
	if ok {
		return fmt.Errorf("ip already exists")
	}
	ipStore[ip.IP] = ip
	if err := fileType.saveIPStore(ipStore); err != nil {
		return fmt.Errorf("error while saving to file : %v", err)
	}
	return nil
}

// DeleteIP from the db
func (fileType *fileType) DeleteIP(ipAddr string) error {
	ipStore, err := fileType.loadIPStore()
	if err != nil {
		return err
	}
	_, ok := ipStore[ipAddr]
	if !ok {
		return fmt.Errorf("ip not found")
	}
	delete(ipStore, ipAddr)
	if err := fileType.saveIPStore(ipStore); err != nil {
		return fmt.Errorf("error while saving to file : %v", err)
	}
	return nil
}

func (fileType *fileType) loadIPStore() (ipStore, error) {
	ipStore := make(ipStore)
	if !fileExists(fileType.path) {
		log.Printf("Db file not found, will be created here: %v\n", fileType.path)
		return ipStore, nil
	}

	fileType.lock.Lock()
	defer fileType.lock.Unlock()
	f, err := os.Open(fileType.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	err = unmarshal(f, ipStore)
	if err != nil {
		return nil, err
	}
	return ipStore, nil
}

func (fileType *fileType) saveIPStore(ipStore ipStore) error {
	fileType.lock.Lock()
	defer fileType.lock.Unlock()

	err := os.MkdirAll(filepath.Dir(fileType.path), 0o750)
	if err != nil {
		return err
	}

	f, err := os.Create(fileType.path)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := marshal(ipStore)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, r)
	return err
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// marshal is a function that marshals the object into an
// io.Reader.
var marshal = func(v ipStore) (io.Reader, error) {
	var buf bytes.Buffer
	e := gob.NewEncoder(&buf)
	err := e.Encode(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf.Bytes()), nil
}

// unmarshal is a function that unmarshals the data from the
// reader into the specified value.
var unmarshal = func(r io.Reader, v ipStore) error {
	d := gob.NewDecoder(r)
	err := d.Decode(&v)
	if err != nil {
		return err
	}
	return nil
}
