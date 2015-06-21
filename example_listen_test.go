package golisten

import (
	"fmt"
	"log"
	"net/http"
	"os/user"
)

func handler(w http.ResponseWriter, req *http.Request) {
	u, err := user.Current()
	if err != nil {
		log.Printf("Error getting user: %s", err)
		return
	}
	fmt.Fprintf(w, "%s\n", u.Uid)
}

func ExampleListenAndServe() {
	ln, err := Listen("guillaume", "tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", handler)
	log.Fatal(http.Serve(ln, nil))
}
