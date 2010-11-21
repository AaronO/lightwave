package main

import (
  "lightwave"
  "http"
  "log"
  "fmt"
  "strings"
)

var servers map[string]*lightwave.Server = make(map[string]*lightwave.Server)

func errorHandler(writer http.ResponseWriter, req *http.Request, errorText string) {
  log.Println(errorText)
  fmt.Fprintf(writer, "<h1>Error in handling %s</h1>", req.URL.Path)
  fmt.Fprintf(writer, "<p>%s</p>", errorText)
}

// Maps wave federation requests to requests of the generalized federation protocol.
// Therefore, this handler is just performing some URL rewriting
func waveFederationHandler(w http.ResponseWriter, r *http.Request) {
  // TODO
}

func clientHandler(w http.ResponseWriter, r *http.Request) {
  // Determine the virtual host
  host := r.Host
  if index := strings.Index(r.Host, ":"); index != -1 {
	host = host[:index]
  }
  server, ok := servers[host]
  if !ok {
	errorHandler(w, r, "Unknown virtual server: " + host)
	return
  }	
  // Parse the URI, strip off the "/client" in front
  uri, ok := lightwave.NewURI(r.URL.Path[7:])
  if !ok {
	errorHandler(w, r, "Error parsing URI")
	return
  }
  // GET handler
  if r.Method == "GET" {
	ch := make(chan bool)
	req := &lightwave.GetRequest{lightwave.Request{w, r.URL.RawQuery, ch, lightwave.ClientOrigin, uri}}
	server.Get( req )
	<- ch
	log.Println("GET finished")
  // POST handler
  } else if r.Method == "POST" {
	// TODO: Meaningful limit on the size of data
	if r.ContentLength < 0 || r.ContentLength > 100000 {
	  errorHandler(w, r, "Negative or oversize content length in HTTP POST body")
	  return
	}
	buffer := make([]byte, r.ContentLength)
	_, err := r.Body.Read(buffer)
	if err != nil {
	  errorHandler(w, r, "Error reading from HTTP POST body")
	  return
	}
	ch := make(chan bool)
	req := &lightwave.PostRequest{lightwave.Request{w, r.URL.RawQuery, ch, lightwave.ClientOrigin, uri}, buffer, ""}
	if req.MimeType, ok = r.Header["Content-Type"]; !ok {
	  errorHandler(w, r, "Content-Type is missing in POST")
	  return
	}
	server.Post( req )
	<- ch
	log.Println("POST finished")
  } else {
	errorHandler(w, r, "Unsupported HTTP method")
  }
}

func federationHandler(w http.ResponseWriter, r *http.Request) {
  // Determine the virtual server
  host := r.Host
  if index := strings.Index(r.Host, ":"); index != -1 {
	host = host[:index]
  }
  server, ok := servers[host]
  if !ok {
	errorHandler(w, r, "Unknown virtual server: " + host)
	return
  }	
  // Parse the URI, strip off the "/fed" in front
  uri, ok := lightwave.NewURI(r.URL.Path[4:])
  if !ok {
	errorHandler(w, r, "Error parsing URI")
	return
  }
  // GET handler
  if r.Method == "GET" {
	ch := make(chan bool)
	req := &lightwave.GetRequest{lightwave.Request{w, r.URL.RawQuery, ch, lightwave.FederationOrigin, uri}}
	server.Get( req )
	<- ch
	log.Println("GET finished")
  // POST handler
  } else if r.Method == "POST" || r.Method == "PUT" {
	// TODO: Meaningful limit on the size of data
	if r.ContentLength < 0 || r.ContentLength > 100000 {
	  errorHandler(w, r, "Negative or oversize content length in HTTP POST body")
	  return
	}
	buffer := make([]byte, r.ContentLength)
	_, err := r.Body.Read(buffer)
	if err != nil {
	  errorHandler(w, r, "Error reading from HTTP POST body")
	  return
	}
	ch := make(chan bool)
	req := &lightwave.PostRequest{lightwave.Request{w, r.URL.RawQuery, ch, lightwave.FederationOrigin, uri}, buffer, ""}
	if req.MimeType, ok = r.Header["Content-Type"]; !ok {
	  errorHandler(w, r, "Content-Type is missing in POST")
	  return
	}
	server.Post( req )
	<- ch
	log.Println("POST finished")
  } else {
	errorHandler(w, r, "Unsupported HTTP method")
  }
}

func main() {  
  log.SetFlags( log.Lshortfile)

  server := lightwave.NewServer(&lightwave.ServerManifest{Domain:"localhost", HostName:"localhost", Port:8080})
  servers["localhost"] = server
  go server.Run()

  server2 := lightwave.NewServer(&lightwave.ServerManifest{Domain:"sony", HostName:"sony", Port:8080})
  servers["sony"] = server2
  go server2.Run()

  // Behave like a wave server with HTTP transport
  http.HandleFunc("/wave/fed/", waveFederationHandler)
  // Run the generalized federation protocol via HTTP. It is more powerful than wave but non-standard
  http.HandleFunc("/fed/", federationHandler)
  http.HandleFunc("/client/", clientHandler)
  http.ListenAndServe(":8080", nil)
}
