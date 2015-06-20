# golisten

Privilege de-escalation listen in Go.

## Overview

`golisten` expects to be root. If it is not, it will error.
As root, perform a `net.Listen()`. Once this is done, we can perform the
privilege de-escalation.
Because of the Go thread model, it is not safe to do so with a "simple" `syscall.Setuid()`.
In order to ensure that the whole process (all threads) are de-escalated, `golsiten` will
fork itself as the requested user while inheriting the privileged listened file descriptor.

`golisten.ListenAndServe` works just like `http.ListenAndServe` but expect the target user
to be run as.

## Example

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os/user"

	"github.com/creack/golisten"
)

func handler(w http.ResponseWriter, req *http.Request) {
	u, err := user.Current()
	if err != nil {
		log.Printf("Error getting user: %s", err)
		return
	}
	fmt.Fprintf(w, "%s\n", u.Uid)
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(golisten.ListenAndServe("guillaume", ":80", nil))
}
```
