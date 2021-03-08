//  Copyright © 2021 Kontain Inc. All rights reserved.

//  Kontain Inc CONFIDENTIAL

//   This file includes unpublished proprietary source code of Kontain Inc. The
//   copyright notice above does not evidence any actual or intended publication of
//   such source code. Disclosure of this source code or any related proprietary
//   information is strictly prohibited without the express written permission of
//   Kontain Inc.

//   Sequential testing of a server with statistics

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"../test/httpstatm"
)

// Flags struct for locating all flags
type Flags struct {
	port     string
	entry    string
	end      string
	reqN     int
	reqDelay int
	verbose  bool
}

func (fg *Flags) serialRequests() httpstatm.Result {
	var result httpstatm.Result
	for i := 0; i < fg.reqN; i++ {
		req, err := http.NewRequest("GET", "http://"+fg.entry+":"+fg.port+"/"+fg.end, nil)
		if err != nil {
			log.Fatal(err)
		}
		ctx := httpstatm.WithClientTrace(req.Context(), &result)
		req = req.WithContext(ctx)
		client := http.DefaultClient
		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := io.Copy(ioutil.Discard, res.Body); err != nil {
			log.Fatal(err)
		}
		res.Body.Close()
		result.End(time.Now())
		time.Sleep(time.Duration(fg.reqDelay) * time.Millisecond)
		client.CloseIdleConnections()
	}
	return result
}

func main() {
	var flags Flags
	flag.StringVar(&flags.port, "port", "8090", "port number")
	flag.StringVar(&flags.entry, "entry", "localhost", "entrypoint")
	flag.StringVar(&flags.end, "end", "healthz", "endpoint")
	flag.IntVar(&flags.reqN, "req", 10, "number of requests")
	flag.IntVar(&flags.reqDelay, "rd", 10, "delays between requests(ms)")
	flag.Parse()
	result := flags.serialRequests()
	fmt.Printf("%+v\n", result)
}
