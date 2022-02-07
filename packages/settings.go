package packages

import (
	"hack-o-ween-site/packages/cookie"
	"hack-o-ween-site/packages/storage"
	"net/http"
	"strconv"
)

func SaveSettings(w http.ResponseWriter, r *http.Request) {
	// Gets the user's session_key, and uses it to get all the user's
	// current settings
	var session_key string
	if temp := cookie.GetCookie("session_key", r); temp != nil {
		session_key = temp.(string)
	} else {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

    row := storage.GetAllFromTable_SessionKey("Settings", session_key)
	var id, name_type_old, theme_type_old int
	row.Scan(&id, &name_type_old, &theme_type_old)

	// Pull the user's selected new settings options
	// If any error with the value, use the old one instead
	name_type, err := strconv.Atoi(r.FormValue("name_type"))
	if err != nil || name_type > 2 || name_type < 0 {
		name_type = int(name_type_old)
	}

	theme_type, err := strconv.Atoi(r.FormValue("theme_type"))
	if err != nil || theme_type > 1 || theme_type < 0 {
		theme_type = int(theme_type_old)
	}

	storage.UpdateUserSettings(session_key, name_type, theme_type)

	http.Redirect(w, r, "/settings", http.StatusTemporaryRedirect)
}
