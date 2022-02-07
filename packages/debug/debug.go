package debug

import (
	"hack-o-ween-site/packages/storage"
	"log"
	"net/http"
)

func DebugButton(w http.ResponseWriter, r *http.Request) {
	log.Println("Runnning Debug Button")
	log.Println(storage.GetFromTable_SessionKey(storage.CURR_EVENT_TABLE, "lalala", "id"))
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
