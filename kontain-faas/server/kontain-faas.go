/*
Copyright (c) 2016-2017 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Design Notes 2/23/2020 (function data passing):
- Input and output for a function is passed in regular files.
- The names of these files is passed as arguments to KRUN and in turn to KM.
- The function kontainer 'main' is snapshot aware and uses new KM hypercalls to get
  parameters and return results.
*/

package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	proxyUtils "github.com/kubeless/kubeless/pkg/function-proxy/utils"
	"github.com/kubeless/kubeless/pkg/functions"
)

var faasMutex = &sync.Mutex{}

var (
	funcContext functions.Context
)

func init() {
	timeout := os.Getenv("FUNC_TIMEOUT")
	if timeout == "" {
		timeout = "180"
	}
	funcContext = functions.Context{
		FunctionName: os.Getenv("FUNC_HANDLER"),
		Timeout:      timeout,
		Runtime:      os.Getenv("FUNC_RUNTIME"),
		MemoryLimit:  os.Getenv("FUNC_MEMORY_LIMIT"),
	}
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func writeRequest(faasName string, e functions.Event) error {
	url := e.Extensions.Request.URL.String()
	return ApiHandlerWriteRequest(faasName, e.Extensions.Request.Method, url, e.Extensions.Request.Header, e.Data)
}

func readResponse(faasName string, e functions.Event) (int, []byte, error) {
	code, res, err := ApiHandlerReadResponse(faasName)
	return code, res, err
}

func processRequest(event functions.Event) (int, []byte, error) {
	faasMutex.Lock()
	defer faasMutex.Unlock()

	urlString := event.Extensions.Request.URL.String()
	faasName, err := GetCallFunction(urlString)
	if err != nil {
		return http.StatusNotFound, []byte(""), err
	}
	err = writeRequest(faasName, event)
	if err != nil {
		return http.StatusInternalServerError, []byte(""), err
	}
	err = ApiHandlerExecCallFunction(faasName)
	if err != nil {
		return http.StatusInternalServerError, []byte(""), err
	}
	return readResponse(faasName, event)
}

func handle(ctx context.Context, w http.ResponseWriter, r *http.Request) ([]byte, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return []byte{}, err
	}
	event := functions.Event{
		Data:           string(data),
		EventID:        r.Header.Get("event-id"),
		EventType:      r.Header.Get("event-type"),
		EventTime:      r.Header.Get("event-time"),
		EventNamespace: r.Header.Get("event-namespace"),
		Extensions: functions.Extension{
			Request:  r,
			Response: w,
			Context:  ctx,
		},
	}

	code, res, err := processRequest(event)

	w.WriteHeader(code)
	return []byte(res), err
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", r.Header.Get("access-control-request-method"))
		w.Header().Set("Access-Control-Allow-Headers", r.Header.Get("access-control-request-headers"))
		return
	}
	proxyUtils.Handler(w, r, handle)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	mux.HandleFunc("/healthz", health)
	mux.Handle("/metrics", proxyUtils.PromHTTPHandler())
	server := proxyUtils.NewServer(mux)

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()

	proxyUtils.GracefulShutdown(server)
}
