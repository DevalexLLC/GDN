package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrAlreadyExists = errors.New("file already exists")
)

// The DB interface defines methods to manipulate the files.
type DB interface {
	Get(id int) *File
	GetAll() []*File
	Find(filename, hash, acl string) []*File
	Add(a *File) (int, error)
	Update(a *File) error
	Delete(id int)
}

// Thread-safe in-memory map of files.
type filesDB struct {
	sync.RWMutex
	m   map[int]*File
	seq int
}

// The one and only database instance.
var db DB

func init() {
	db = &filesDB{
		m: make(map[int]*File),
	}
	// Fill the database
	db.Add(&File{Id: 1, FileName: "hello.go", Hash: "F810B74143BE5F06D1CE1A22D9FEE7D6", ACL: "private"})
	db.Add(&File{Id: 2, FileName: "3420 Boelter Hall.txt", Hash: "F966AA92D412BB814BA98426264CE375", ACL: "public-read-write"})
}

// GetAll returns all files from the database.
func (db *filesDB) GetAll() []*File {
	db.RLock()
	defer db.RUnlock()
	if len(db.m) == 0 {
		return nil
	}
	ar := make([]*File, len(db.m))
	i := 0
	for _, v := range db.m {
		ar[i] = v
		i++
	}
	return ar
}

// Find returns files that match the search criteria.
func (db *filesDB) Find(filename, hash, acl string) []*File {
	db.RLock()
	defer db.RUnlock()
	var res []*File
	for _, v := range db.m {
		if v.FileName == filename || filename == "" {
			if v.Hash == hash || hash == "" {
				if v.ACL == acl || acl == "" {
					res = append(res, v)
				}
			}
		}
	}
	return res
}

// Get returns the file identified by the id, or nil.
func (db *filesDB) Get(id int) *File {
	db.RLock()
	defer db.RUnlock()
	return db.m[id]
}

// Add creates a new file and returns its id, or an error.
func (db *filesDB) Add(a *File) (int, error) {
	db.Lock()
	defer db.Unlock()
	// Return an error if filename-hash already exists
	if !db.isUnique(a) {
		return 0, ErrAlreadyExists
	}
	// Get the unique ID
	db.seq++
	a.Id = db.seq
	// Store
	db.m[a.Id] = a
	return a.Id, nil
}

// Update changes the file identified by the id. It returns an error if the
// updated file is a duplicate.
func (db *filesDB) Update(a *File) error {
	db.Lock()
	defer db.Unlock()
	if !db.isUnique(a) {
		return ErrAlreadyExists
	}
	db.m[a.Id] = a
	return nil
}

// Delete removes the file identified by the id from the database. It is a no-op
// if the id does not exist.
func (db *filesDB) Delete(id int) {
	db.Lock()
	defer db.Unlock()
	delete(db.m, id)
}

// Checks if the file already exists in the database, based on the FileName and Hash
// fields.
func (db *filesDB) isUnique(a *File) bool {
	for _, v := range db.m {
		if v.FileName == a.FileName && v.Hash == a.Hash && v.Id != a.Id {
			return false
		}
	}
	return true
}

// The File data structure, serializable in JSON, XML and text using the Stringer interface.
type File struct {
	XMLName  xml.Name `json:"-" xml:"file"`
	Id       int      `json:"id" xml:"id,attr"`
	FileName string   `json:"filename" xml:"filename"`
	Hash     string   `json:"hash" xml:"hash"`
	ACL      string   `json:"acl" xml:"acl"`
}

func (a *File) String() string {
	return fmt.Sprintf("%s - %s (%s)", a.FileName, a.Hash, a.ACL)
}
