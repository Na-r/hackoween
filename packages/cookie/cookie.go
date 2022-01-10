package cookie

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("586E3272357538782F413F4428472B4B6250655368566B597033733676397924")
	store = sessions.NewCookieStore(key)
)

const SESSION_NAME = "hack-o-ween"

func GetCookie(name string, r *http.Request) interface{} {
	session, _ := store.Get(r, SESSION_NAME)

	if val, ok := session.Values[name]; !ok {
		return nil
	} else {
		return val
	}
}

func StoreCookie(name string, val interface{}, w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SESSION_NAME)
	session.Values[name] = val
	session.Save(r, w)
}
