// +build nosystray

package systray

import "log"

func Run() {
	log.Println("Systray support not compiled in. Use -tags=\"!nosystray\" to enable.")
}