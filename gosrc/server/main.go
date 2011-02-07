package main

import (
  "lightwave"
  "http"
  "log"
  "fmt"
  "strings"
  "os"
//  "wave"
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

func fileHandler(w http.ResponseWriter, r *http.Request) {
  // Determine the virtual host
  server, err := findServer(r.Host)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  // log.Println("URI is ", "/_static" + r.URL.Path)
  // Parse the URI, prepend "/_static"
  uri, ok := lightwave.NewURI("/_static" + r.URL.Path)
  if !ok {
    errorHandler(w, r, "Error parsing URI")
    return
  }
  // GET handler
  if r.Method != "GET" {
    errorHandler(w, r, "Unsupported HTTP method")
  }
  
  // Requests for the application page are redirected to the login page
  if r.URL.Path == server.Config.MainPage { 
    // Find the session
    _, err := server.SessionDatabase.FindSession(r)
    if err != nil {
      http.Redirect(w, r, server.Config.LoginPage, 303)
      return
    }
  }

  ch := make(chan bool)
  req := &lightwave.GetRequest{lightwave.Request{w, r.URL.RawQuery, ch, lightwave.ClientOrigin, uri, nil}}
  server.Get( req )
  <- ch
}

func clientHandler(w http.ResponseWriter, r *http.Request) {
  // Determine the virtual host
  server, err := findServer(r.Host)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  // Find the session
  session, err := server.SessionDatabase.FindSession(r)
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
    req := &lightwave.GetRequest{lightwave.Request{w, r.URL.RawQuery, ch, lightwave.ClientOrigin, uri, session}}
    server.Get( req )
    <- ch
    log.Println("GET finished ", r.URL)
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
    req := &lightwave.PostRequest{lightwave.Request{w, r.URL.RawQuery, ch, lightwave.ClientOrigin, uri, session}, buffer, ""}
    if req.MimeType, ok = r.Header["Content-Type"]; !ok {
      errorHandler(w, r, "Content-Type is missing in POST")
      return
    }
    server.Post( req )
    <- ch
    log.Println("POST finished ", r.URL)
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
    req := &lightwave.GetRequest{lightwave.Request{w, r.URL.RawQuery, ch, lightwave.FederationOrigin, uri, nil}}
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
    req := &lightwave.PostRequest{lightwave.Request{w, r.URL.RawQuery, ch, lightwave.FederationOrigin, uri, nil}, buffer, ""}
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

func sessionInfoHandler(w http.ResponseWriter, r *http.Request) {
  // Determine the virtual host
  server, err := findServer(r.Host)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  // GET handler
  if r.Method != "GET" {
    errorHandler(w, r, "Unsupported HTTP method")
    return
  }
  // Find the session
  session, err := server.SessionDatabase.FindSession(r)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  session.InfoHandler(w, r)
}

func sessionGetHandler(w http.ResponseWriter, r *http.Request) {
  // Determine the virtual host
  server, err := findServer(r.Host)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  // GET handler
  if r.Method != "GET" {
    errorHandler(w, r, "Unsupported HTTP method")
    return
  }
  // Find the session
  session, err := server.SessionDatabase.FindSession(r)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  wait := make(chan bool)
  session.Enqueue( &lightwave.SessionGetRequest{lightwave.SessionRequest{Response:w, FinishSignal:wait}} )
  <- wait
}

func sessionPollHandler(w http.ResponseWriter, r *http.Request) {
  // Determine the virtual host
  server, err := findServer(r.Host)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  // GET handler
  if r.Method != "GET" {
    errorHandler(w, r, "Unsupported HTTP method")
    return
  }
  // Find the session
  session, err := server.SessionDatabase.FindSession(r)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  wait := make(chan bool)
  session.Enqueue( &lightwave.SessionPollRequest{lightwave.SessionRequest{Response:w, FinishSignal:wait}} )
  <- wait
}

func openHandler(w http.ResponseWriter, r *http.Request) {
  // Determine the virtual host
  server, err := findServer(r.Host)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  // GET handler
  if r.Method != "GET" {
    errorHandler(w, r, "Unsupported HTTP method")
    return
  }
  // Find the session
  session, err := server.SessionDatabase.FindSession(r)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  log.Println("OPEN", r.FormValue("uri"))
  wait := make(chan bool)
  session.Enqueue( &lightwave.SessionOpenDocRequest{lightwave.SessionRequest{Response:w, FinishSignal:wait}, r.FormValue("uri")} )
  <- wait
}

func closeHandler(w http.ResponseWriter, r *http.Request) {
  // Determine the virtual host
  server, err := findServer(r.Host)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  // GET handler
  if r.Method != "GET" {
    errorHandler(w, r, "Unsupported HTTP method")
    return
  }
  // Find the session
  session, err := server.SessionDatabase.FindSession(r)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  wait := make(chan bool)
  session.Enqueue( &lightwave.SessionCloseDocRequest{lightwave.SessionRequest{Response:w, FinishSignal:wait}, r.FormValue("uri")} )
  <- wait
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
  // Determine the virtual host
  server, err := findServer(r.Host)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  
  username := r.FormValue("username")
  password := r.FormValue("password")
  email := r.FormValue("email")
  nickname := r.FormValue("nickname")

  log.Println("Register:",username)
  user, err := server.UserAccountDatabase.SignUpUser(email, nickname, username, password)
  if err != nil {
    log.Println(err)
    http.Redirect(w, r, server.Config.SignupPage, 303)
    return
  }
  s, err := server.SessionDatabase.CreateSession(user.Username)
  if err != nil {
    log.Println(err)
    http.Redirect(w, r, server.Config.LoginPage, 303)
    return
  }  
  s.SetCookie(w)
  http.Redirect(w, r, server.Config.MainPage, 303)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
  // Determine the virtual host
  server, err := findServer(r.Host)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  
  username := r.FormValue("username")
  password := r.FormValue("password")
  log.Println("LOGIN:",username)
  if err = server.UserAccountDatabase.CheckCredentials(username, password); err != nil {
    log.Println(err)
    http.Redirect(w, r, server.Config.LoginPage, 303)
    return
  }
  s, err := server.SessionDatabase.CreateSession(username)
  if err != nil {
    log.Println(err)
    http.Redirect(w, r, server.Config.LoginPage, 303)
    return
  }
  s.SetCookie(w)
  http.Redirect(w, r, server.Config.MainPage, 303)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
  // Determine the virtual host
  server, err := findServer(r.Host)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  // GET handler
  if r.Method != "GET" {
    errorHandler(w, r, "Unsupported HTTP method")
    return
  }
  // Find the session
  session, err := server.SessionDatabase.FindSession(r)
  if err != nil {
    errorHandler(w, r, err.String())
    return
  }
  server.SessionDatabase.DeleteSession(session)
  http.Redirect(w, r, server.Config.LoginPage, 303)
}

func configure() (*lightwave.Config, os.Error) {
  config, err := lightwave.ReadConfig()
  if err != nil {
    return nil, err
  }
  // Start all servers listed in the configuration file
  for _, s := range config.Servers {
    serverconfig, err := lightwave.ReadServerConfig(config, s)
    if err != nil {
      return nil, err
    }
    // Start the server
    server := lightwave.NewServer(serverconfig)
    servers[serverconfig.Hostname] = server
    // Make static files available
    f := lightwave.NewStaticNode(server, serverconfig.StaticRoot)
    go f.Run()
    server.AddChild(f)
    // TODO: This is a hack
    server.LocalHost().Users().CreateUser("weis")
    server.LocalHost().Users().CreateUser("tux")
    server.LocalHost().Users().CreateUser("konqi")
    // TODO: This is a hack
    if _, err := server.UserAccountDatabase.FindUser("weis"); err != nil {
      server.UserAccountDatabase.SignUpUser("weis@mail.com", "Torben", "weis", "pass")
    }
    if _, err := server.UserAccountDatabase.FindUser("tux"); err != nil {
      server.UserAccountDatabase.SignUpUser("tux123@mail.com", "Tux", "tux", "pass2")
    }
    if _, err := server.UserAccountDatabase.FindUser("konqi"); err != nil {  
      server.UserAccountDatabase.SignUpUser("kon@mail.com", "Konqi", "konqi", "pass3")
    }
    // End hack
    // Start the server
    go server.Run()
  }
  return config, nil
}

func main() {  
  log.SetFlags( log.Lshortfile)
  // Configure and start all servers
  config, err := configure()
  if err != nil {
    log.Exitln( err )
    return
  }
  
//  lightwave.RegisterNodeFactory("application/x-protobuf-wave", wave.NewWaveletNode)
//  lightwave.RegisterNodeFactory("application/x-json-wave", wave.NewWaveletNode)
  
  // Allow for JSON documents
  lightwave.RegisterNodeFactory("application/json", lightwave.DocumentNodeFactory)
  // Login pages post here to login and to be redirected
  http.HandleFunc("/_login", loginHandler)
  // Logout gets thus page to logout and to be redirected
  http.HandleFunc("/_logout", logoutHandler)
  // RPC to retrieve information about the session bound to the session cookie
  http.HandleFunc("/_sessioninfo", sessionInfoHandler)
  // The SignUp page posts here to register a new user and to be redirected
  http.HandleFunc("/_signup", signupHandler)
  // Opens a document and sends a snapshot to the client. Further updates to the document are sent as well
  http.HandleFunc("/_open", openHandler)
  // Stops sending updates of this document to the client
  http.HandleFunc("/_close", closeHandler)
  // Retrieves the latest document updates belonging to the current session. Return immediately if there are none.
  http.HandleFunc("/_sessionget", sessionGetHandler)
  // Retrieves the latest document updates belonging to the current session. Wait if there are none.
  http.HandleFunc("/_sessionpoll", sessionPollHandler)
  
  // Behave like a wave server with HTTP transport
//  http.HandleFunc("/wave/fed/", waveFederationHandler)
  // Run the generalized federation protocol via HTTP. It is more powerful than wave but non-standard
  http.HandleFunc("/fed/", federationHandler)
  // Run the client/server protocol
  http.HandleFunc("/client/", clientHandler)
  // Serve static files (HTML, images)
  http.HandleFunc("/", fileHandler)
  err = http.ListenAndServe(fmt.Sprintf(":%v", config.Port), nil)
  if err != nil {
    log.Exitln(err)
  }
}
