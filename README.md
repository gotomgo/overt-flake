# overt-flake
Flake ID Generation server developed for use with Overtone - by Overtone Studios

Overt-flake is a Flake ID generator and server (written in GO) along the lines of Twitter Snowflake, Boundary Flake and others. It deviates from other implementations in small but important ways:

1. Identifiers are 128-bits
2. External configuration information such as worker id and data-center id are not needed. Machine identifiers, both stable and unstable are used instead
3. The Overtone Epoch (Jan 1, 2017) is the default epoch used. The code allows for any epoch, including Twitter Epoch or Unix Epoch. The primary reason is that 1/1/2017 is a Sunday (not important for ID generation) and nostalgically, it is also the day technical development of Overtone began.

## Simple Client Example

```golang
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gotomgo/overt-flake/ofsclient"
)

func main() {
	// create a new client with our auth token ("test") and at least 1 server address
	client, err := ofsclient.NewClient("test", "0.0.0.0:4444")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating OFS client: %s\n", err)
		os.Exit(-1)
	}

	// generate 10 ids in the form of big.Int's
	bigInts, err := client.GenerateIDsAsBigInt(10)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating overt-flake ids: %s\n", err)
		os.Exit(-1)
	}

	// dump
	for i := 0; i < len(bigInts); i++ {
		fmt.Println(bigInts[i].String())
	}

	// wait (just because)
	select {
	case <-time.After(1 * time.Second):
	}

	// for completeness sake
	client.Close()
}
```
