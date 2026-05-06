package debug

import (
	"fmt"
	"net/http"
)

func PrintForm(r *http.Request) {
	r.ParseForm()
	fmt.Println("=== FORM ===")
	for key, values := range r.Form {
		fmt.Printf("  %s: %v\n", key, values)
	}
	fmt.Println("============")
}
