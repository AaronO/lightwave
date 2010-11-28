package main

import (
  "lightwave"
  "http"
  "log"
  "fmt"
  "strings"
  "os"
  "wave"
)

var servers map[string]*lightwave.Server = make(map[string]*lightwave.Server)

func errorHandler(writer http.ResponseWriter, req *http.Request, errorText string) {
  log.Println(errorText)
  fmt.Fprintf(writer, "<h1>Error in handling %s</h1>", req.URL.Path)
  fmt.Fprintf(writer, "<p>%s</p>", errorText)
}

// Determine the virtual host
func findServer(host string) (server *lightwave.Server, err os.Error) {
  // Strip off the port number
  if index := strings.Index(host, ":"); index != -1 {
	host = host[:index]
  }
  server, ok := servers[host]
  if !ok {
	return nil, os.NewError("Unknown virtual server: " + host)
  }
  return server, nil
}

// Maps wave federation requests to requests of the generalized federation protocol.
// Therefore, this handler is performing some URL rewriting
func waveFederationHandler(w http.ResponseWriter, r *http.Request) {
  // Determine the virtual host
  server, err := findServer(r.Host)
  if err != nil {
	errorHandler(w, r, err.String())
	return
  }
  // The URL is of the form http://host/wave/fed/wave-host/wave-id/wavelet-host/wavelet-id    
  waveurl, err := wave.NewWaveUrlFromId(r.URL.Path[5:])
  if err != nil {
	errorHandler(w, r, "Error parsing wave ID")
	return
  }
  uri, _ := lightwave.NewURI("/" + waveurl.WaveletDomain + "/" + waveurl.WaveDomain + "$" + waveurl.WaveId + "$" + waveurl.WaveletId)
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
	var ok bool
	if req.MimeType, ok = r.Header["Content-Type"]; !ok {
	  errorHandler(w, r, "Content-Type is missing in POST")
	  return
	}
	if req.MimeType != "application/x-protobuf-wave" {
	  errorHandler(w, r, "Content-Type " + req.MimeType + " is not understood")
	  return
	}
	server.Post( req )
	<- ch
	log.Println("POST finished")
  } else {
	errorHandler(w, r, "Unsupported HTTP method")
  }  
}

func clientHandler(w http.ResponseWriter, r *http.Request) {
  // Determine the virtual host
  server, err := findServer(r.Host)
  if err != nil {
	errorHandler(w, r, err.String())
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
  // Determine the virtual host
  server, err := findServer(r.Host)
  if err != nil {
	errorHandler(w, r, err.String())
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

  lightwave.RegisterNodeFactory("application/x-protobuf-wave", wave.NewWaveletNode)
  lightwave.RegisterNodeFactory("application/x-json-wave", wave.NewWaveletNode)
  lightwave.RegisterNodeFactory("application/json", lightwave.NewDocumentNode)
  
  // Behave like a wave server with HTTP transport
  http.HandleFunc("/wave/fed/", waveFederationHandler)
  // Run the generalized federation protocol via HTTP. It is more powerful than wave but non-standard
  http.HandleFunc("/fed/", federationHandler)
  http.HandleFunc("/client/", clientHandler)
  http.ListenAndServe(":8080", nil)
}
