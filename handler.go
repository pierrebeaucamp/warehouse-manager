package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

// Add adds a new file to a the cloud storage provider listed in the cookie
func Add(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	err = gd.Add(cookie.Value, r.Body, ps.ByName("filepath"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// AuthURLJSON is the struct which will be encoded into JSON once it's been
// initialized by AuthURL().
type AuthURLJSON struct {
	URL string `json:"url"`
}

// AuthURL gets an oauth2 URL from one of the supported libraries (depending
// on httprouter.Params) and returns the link encoded in JSON.
// If httprouter.Params specify an unsupported library, http.StatusNotFound
// is returned.
func AuthURL(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var a = AuthURLJSON{}

	switch ps.ByName("provider") {
	default:
		http.Error(w, "Provider not found", http.StatusNotFound)
		return

	case "google":
		a = AuthURLJSON{URL: gd.AuthURL()}
	}

	j, _ := json.Marshal(a)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(j))
}

// BrowseJSON is the struct which will be encoded into JSON once it's been
// initialized by Browse()
type BrowseJSON struct {
	FileList []string `json:"file_list"`
}

// Browse returns the content of a directory as a json list
func Browse(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "authenticated", http.StatusUnauthorized)
		return
	}

	list, err := gd.Browse(cookie.Value, ps.ByName("filepath"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	j, _ := json.Marshal(BrowseJSON{FileList: list})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(j))
}

// Delete deletes the given file
func Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	err = gd.Delete(cookie.Value, ps.ByName("filepath"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// PublishJSON is the struct which will be encoded into JSON once it's been
// initialized by Publish
type PublishJSON struct {
	URL string `json:"url"`
}

// Publish sets the given file to public
func Publish(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	link, err := gd.Publish(cookie.Value, ps.ByName("filepath"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	j, _ := json.Marshal(PublishJSON{URL: link})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(j))
}

// Read gets the given file and returns its content
func Read(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	response, err := gd.Read(cookie.Value, ps.ByName("filepath"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if response == nil {
		return
	}

	io.Copy(w, response.Body)
	response.Body.Close()
}

// ValidateJSON is the struct which will be encoded into JSON once it's been
// initialized by Validate().
type ValidateJSON struct {
	Token  string    `json:"access_token"`
	Expiry time.Time `json:"expiry,omitempty"`
}

// Validate reads the Form Values of a request and validates the oauth2.
// After the code is validated, it returns the user token.
func Validate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	state := r.FormValue("state")

	var err error
	var token string
	var expiry time.Time

	// TODO: validate state token and map to a specific provider
	switch state {
	default:
		http.Error(w, "Invalid state token", http.StatusBadRequest)
		return

	case "google":
		token, expiry, err = gd.Validate(r.FormValue("code"))
	}

	if err != nil {
		http.Error(w, "Auth Code invalid", http.StatusBadRequest)
		return
	}

	j, _ := json.Marshal(ValidateJSON{Token: token, Expiry: expiry})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(j))
}
