package cookie

import (
	"fmt"

	"github.com/gorilla/sessions"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("586E3272357538782F413F4428472B4B6250655368566B597033733676397924")
	store = sessions.NewCookieStore(key)
)

func init() {
	fmt.Printf("%s\n", key)
}
