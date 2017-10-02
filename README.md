# overt-flake
Flake ID Generation server developed for use with Overtone - by Overtone Studios

Overt-flake is a Flake ID generator and server (written in GO) along the lines of Twitter Snowflake, Boundary Flake and others. It deviates from other implementations in small but important ways:

1. Identifiers are 128-bits
2. External configuration information such as worker id and data-center id are not needed. Machine identifiers, both stable and unstable are used instead
3. The Overtone Epoch (Jan 1, 2017) is the default epoch used. The code allows for any epoch, including Twitter Epoch or Unix Epoch. The primary reason is that 1/1/2017 is a Sunday (not important for ID generation) and nostalgically, it is also the day technical development of Overtone began.
4. The server is not specific to overt-flake identifiers. Write your own implementation of flake.Generator and plug it in to the server. A
Twitter Snowflake compatible ID generator is included

## Over-Flake ID Format

```
//  ---------------------------------------------------------------------------
//  Layout - Big Endian
//  ---------------------------------------------------------------------------
//
//  [0:6]   48 bits | Upper 48 bits of timestamp (milliseconds since the epoch)
//  [6:8]   16 bits | a per-interval sequence # (interval == 1 millisecond)
//  [8:14]  48 bits | a hardware id
//  [14:16] 16 bits | process ID
//
//  ---------------------------------------------------------------------------
//  | 0 | 1 | 2 | 3 | 4 | 5 |  6  |  7  |  8  | 9 | A | B | C | D |  E  |  F  |
//  ---------------------------------------------------------------------------
//  |           48 bits     |  16 bits  |         48 bits         |  16 bits  |
//  ---------------------------------------------------------------------------
//  |          timestamp    |  sequence |        HardwareID       | ProcessID |
//  ---------------------------------------------------------------------------
//  Notes
//  ---------------------------------------------------------------------------
//  The time bits are the most significant bits because they have the primary
//  impact on the sort order of ids. The sequence # is next most significant
//  as it is the tie-breaker when the time portions are equivalent.
//
//  Note that the lower 64 bits are basically random and not specifically
//  useful for ordering, although they play their part when the upper 64-bits
//  are equivalent between two ids. Again, the ordering outcome in this
//  situation is somewhat random, but generally somewhat repeatable (hardware
//  id should be consistent and stable a vast majority of the time).
//  ---------------------------------------------------------------------------
```

## Server Console Usage

executing ofsrvr -help shows the following usage:

```
ofsrvr: 0.3.1, Runtime: gc, Compiler: go1.8.3, Copyright © 2017 Overtone Studios, Inc.
Usage: ofsrvr [options]
Options:
    -ip              specify the network interface/address to listen on                 default=0.0.0.0:4444
    -hidtype         specify the type of the hardware ID provider                       default=mac
    -gentype         specify the type of generator used to generate IDs                 default=default
    -epoch           specify the epoch in milliseconds elapsed since Unix Epoch         default=1483228800000
    -waitfor         specify a time at which id generation may start, but not before    default=0
    -auth            specify the sequence of characters that make up the auth token     default=""
    -config          specify a path to a configuration file                             default=""
    -hid             specify a hardware id to use when -hidype == "fixed"               default=""

Notes:
* arguments specified on the command-line override values specified in -config file
* waitfor *must* be specified on the command line

Hid Types:
    simple           simple MAC hardware ID provider
    mac              standard MAC hardware ID provider (default)
    fixed            specifies that a fixed hardware id is used (see -hid)

Generator Types:
    default          the standard overt-flake ID generator

Common Options:
    -help, --help    Show this message
    -v, --version    Show version
```

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

## Acknowledgements
There is almost nothing new under the sun here. The packaging and level of functionality may exceed some available packages, but the fundamental concepts I borrowed are embodied in many, many, many existing packages. I reviewed most of them and was inspired by a small number of them:

* [NOEQD](https://github.com/noeq/noeqd)
* [goflake](https://github.com/nmjmdr/goflake)

I really loved [NOEQD's](https://github.com/noeq/noeqd) approach to the client/server interactions (SIMPLE) so
I literally stole the code and extended it a bit. I borrowed other things from NOEQD as well. If you need 64 bit identifiers
ala Twitter Snowflake, [NOEQD](https://github.com/noeq/noeqd) is a great choice. An even better choice would be to use the
OvertFlake server with a NOEQD/Twitter 64-bit id generator!

[goflake](https://github.com/nmjmdr/goflake) uses Hardware MAC IDs and produces a SHA1 of all available MAC addresses. I originally
felt like a literal MAC address should not be part of the flake, and would need to be obsfucated (to prevent any possible system
re-trace or fingerprint by nefarious agents) so the [goflake](https://github.com/nmjmdr/goflake) implementation struck me as the right approach.

I will say I am not a big fan of the 160 bit identifier (from Factual\Skuld'd) because I feel like the extra 32 bits are wasted on
12 bits of time that will never be used, and an additional 16 bits for sequence # seems like a waste of 16 bits. As a matter of fact, 42 bits for time would be sufficient for most systems (assuming a recent epoch), and we can use the extra 6 bits for the sequence # (22 bits total) which would allow for a single node to produce ~4.2 million ids per millisecond without stalling.
