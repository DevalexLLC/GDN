package main

import (
	"fmt"
	"github.com/codegangsta/martini"
	"net/http"
	"path/filepath"
	"strconv"
)

// GetFiles returns the list of files (possibly filtered).
func GetFiles(r *http.Request, enc Encoder, db DB) string {
	// Get the query string arguments, if any
	qs := r.URL.Query()
	filename, hash, acl := qs.Get("filename"), qs.Get("hash"), qs.Get("acl")

	if filename != "" || hash != "" || acl != "" {
		// At least one filter, use Find()
		return Must(enc.Encode(toIface(db.Find(filename, hash, acl))...))
	}
	// Otherwise, return all files
	return Must(enc.Encode(toIface(db.GetAll())...))
}

// GetFile returns the requested file.
func GetFile(enc Encoder, db DB, parms martini.Params) (int, string) {
	id, err := strconv.Atoi(parms["id"])
	al := db.Get(id)
	if err != nil || al == nil {
		// Invalid id, or does not exist
		return http.StatusNotFound, Must(enc.Encode(
			NewError(ErrCodeNotExist, fmt.Sprintf("the file with id %s does not exist", parms["id"]))))
	}
	return http.StatusOK, Must(enc.Encode(al))
}

// AddFile creates the posted file.
func AddFile(w http.ResponseWriter, r *http.Request, enc Encoder, db DB) (int, string) {
	al := getPostFile(r)
	// Copy the file to a storage location
	
	// Do not store the entire path, just the filename
	_, file := filepath.Split(al.FileName)
	al.FileName = file
	
	// Add the file to the database
	id, err := db.Add(al)
	switch err {
	case ErrAlreadyExists:
		// Duplicate
		return http.StatusConflict, Must(enc.Encode(
			NewError(ErrCodeAlreadyExists, fmt.Sprintf("the file '%s' from '%s' already exists", al.Hash, al.FileName))))
	case nil:
		// TODO : Location is expected to be an absolute URI, as per the RFC2616
		w.Header().Set("Location", fmt.Sprintf("/files/%d", id))
		return http.StatusCreated, Must(enc.Encode(al))
	default:
		panic(err)
	}
}

// UpdateFile changes the specified file.
func UpdateFile(r *http.Request, enc Encoder, db DB, parms martini.Params) (int, string) {
	al, err := getPutFile(r, parms)
	if err != nil {
		// Invalid id, 404
		return http.StatusNotFound, Must(enc.Encode(
			NewError(ErrCodeNotExist, fmt.Sprintf("the file with id %s does not exist", parms["id"]))))
	}
	err = db.Update(al)
	switch err {
	case ErrAlreadyExists:
		return http.StatusConflict, Must(enc.Encode(
			NewError(ErrCodeAlreadyExists, fmt.Sprintf("the file '%s' from '%s' already exists", al.Hash, al.FileName))))
	case nil:
		return http.StatusOK, Must(enc.Encode(al))
	default:
		panic(err)
	}
}

// Parse the request body, load into an File structure.
func getPostFile(r *http.Request) *File {
	filename, hash, acl := r.FormValue("filename"), r.FormValue("hash"), r.FormValue("acl")

	return &File{
		FileName: filename,
		Hash:     hash,
		ACL:      acl,
	}
}

// Like getPostFile, but additionnally, parse and store the `id` query string.
func getPutFile(r *http.Request, parms martini.Params) (*File, error) {
	al := getPostFile(r)
	id, err := strconv.Atoi(parms["id"])
	if err != nil {
		return nil, err
	}
	al.Id = id
	return al, nil
}

// Martini requires that 2 parameters are returned to treat the first one as the
// status code. Delete is an idempotent action, but this does not mean it should
// always return 204 - No content, idempotence relates to the state of the server
// after the request, not the returned status code. So I return a 404 - Not found
// if the id does not exist.
func DeleteFile(enc Encoder, db DB, parms martini.Params) (int, string) {
	id, err := strconv.Atoi(parms["id"])
	al := db.Get(id)
	if err != nil || al == nil {
		return http.StatusNotFound, Must(enc.Encode(
			NewError(ErrCodeNotExist, fmt.Sprintf("the file with id %s does not exist", parms["id"]))))
	}
	db.Delete(id)
	return http.StatusNoContent, ""
}

func toIface(v []*File) []interface{} {
	if len(v) == 0 {
		return nil
	}
	ifs := make([]interface{}, len(v))
	for i, v := range v {
		ifs[i] = v
	}
	return ifs
}
