package lightwave

import (
  "log"
  ioutil "io/ioutil"
  "strings"
  "http"
)

// --------------------------------------------------------
// Helper functions

// Creates a HTTP error response
func makeHtmlErrorResponse(res http.ResponseWriter, errorText string) {
  log.Println(errorText)
  res.SetHeader("Content-Type", "text/html")
  _, err := res.Write( []byte(errorText) )
  if err != nil {
	log.Println("Failed writing HTTP response")
  }
}

// -----------------------------------------------------------
// StaticNode

type StaticNode struct {
  NodeBase
  FilePath string
}

func NewStaticNode(parent Node, filePath string) *StaticNode {
  r := &StaticNode{FilePath:filePath, NodeBase:NodeBase{parent:parent, name:"_static", postChannel:make(chan *PostRequest), getChannel:make(chan *GetRequest), stopChannel:make(chan bool), pubSubChannel:make(chan *PubSubRequest)}}
  return r
}

func (self *StaticNode) Run() {
  for {
	select {
	  case req := <-self.postChannel:
		self.post(req)
	  case req := <-self.getChannel:
		self.get(req)
	  case req := <-self.pubSubChannel:
		self.pubSub(req)
	  case <-self.stopChannel:
		return
	}
  }
}

func (self *StaticNode) post(req *PostRequest) {
  makeHtmlErrorResponse(req.Response, "Unsupported HTTP method")
  req.FinishSignal <- false
}

func (self *StaticNode) get(req *GetRequest) {
  uri, ok := req.URI.(StaticURI)
  if !ok {
	makeHtmlErrorResponse(req.Response, "Unsupported URI type for static content")
	req.FinishSignal <- false
	return	  
  }

  if len(uri.NameSeq) == 0 {
	uri.NameSeq = append( uri.NameSeq, "index.html" )
  }
  filename := uri.NameSeq[len(uri.NameSeq) - 1]
  if strings.HasSuffix(filename, ".html" ) {
	req.Response.SetHeader("Content-Type", "text/html")
  } else if strings.HasSuffix(filename, ".js" ) {
	req.Response.SetHeader("Content-Type", "text/javascript")
  } else if strings.HasSuffix(filename, ".jpg" ) {
	req.Response.SetHeader("Content-Type", "image/jpeg")
  } else if strings.HasSuffix(filename, ".png" ) {
	req.Response.SetHeader("Content-Type", "image/png")
  } else if strings.HasSuffix(filename, ".gif" ) {
	req.Response.SetHeader("Content-Type", "image/gif")
  } else {
	req.Response.SetHeader("Content-Type", "application/octet-stream")
  }

  n := self.FilePath + "/" + strings.Join(uri.NameSeq, "/")
  log.Println("GETting file ", n)
  data, err := ioutil.ReadFile(n)
  if err != nil {
	makeHtmlErrorResponse(req.Response, "File not found")
	req.FinishSignal <- false
	return	  
  }
  req.Response.Write(data)
  req.FinishSignal <- true
}

func (self *StaticNode) pubSub(req* PubSubRequest) {
  // Do nothing by design
}

