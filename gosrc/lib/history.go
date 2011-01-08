package lightwave

import (
  "os"
  "json"
  "strings"
  "fmt"
  "bufio"
  "log"
  bin "encoding/binary"
)

type DocumentHistory struct {
  firstInMemoryVersion int64
  deltaFilePosition int64
  history []DocumentMutation
  commitVersion int64
  fdelta *os.File
  findex *os.File
  fsnapshot *os.File
  node *DocumentNode
  broken bool
}

func isDocumentPersisted(uri string) bool {
  lst := strings.Split(uri[1:], "/", -1)
  pathlst := make([]string, len(lst) - 1)
  // Create the path
  if len(lst) > 0 {
    for i, s := range lst[0:len(lst)-1] {
      pathlst[i] = s + ".dir"
    }
  }
  path := strings.Join(pathlst, "/") + "/" + lst[len(lst)-1] + ".deltas"
  _, err := os.Stat(path)
  if err != nil {
    return false
  }
  return true
}

func NewDocumentHistory(node *DocumentNode) *DocumentHistory {
  uri := node.URI()
  lst := strings.Split(uri[1:], "/", -1)
  pathlst := make([]string, len(lst) - 1)
  // Create the path
  if len(lst) > 0 {
    for i, s := range lst[0:len(lst)-1] {
      pathlst[i] = s + ".dir"
    }
    path := strings.Join(pathlst, "/")
    err := os.MkdirAll(path, 0700)
    if err != nil {
      log.Println("Cannot create path " + path)
    }
  }
  // Create the files
  path := strings.Join(pathlst, "/") + "/" + lst[len(lst)-1]
  fdelta, err1 := os.Open(path + ".deltas", os.O_RDWR | os.O_CREATE, 0700)
  findex, err2 := os.Open(path + ".index", os.O_RDWR | os.O_CREATE, 0700)
  fsnapshot, err3 := os.Open(path + ".snapshot", os.O_RDWR | os.O_CREATE, 0700)
  if err1 != nil || err2 != nil || err3 != nil {
    log.Println("Cannot write to file " + uri + ".deltas or .index or .snapshot")
    if fdelta != nil {
      fdelta.Close()
    }
    if findex != nil {
      findex.Close()
    }
  }
  h := &DocumentHistory{node:node, history:make([]DocumentMutation,0), fdelta:fdelta, findex:findex, fsnapshot:fsnapshot}
  h.initialRead()
  return h
}

func (self *DocumentHistory) Close() {
  if self.fdelta != nil {
    self.fdelta.Close()
    self.fdelta = nil
  }
  if self.findex != nil {
    self.findex.Close()
    self.findex = nil
  }
}

func (self *DocumentHistory) Commit() {
  // Write delta and index
  self.fdelta.Seek(0, 2)
  self.findex.Seek(0, 2)
  bufdelta := bufio.NewWriter(self.fdelta)
  bufindex := bufio.NewWriter(self.findex)
  for i := self.commitVersion - self.firstInMemoryVersion; i < int64(len(self.history)); i++ {
    bin.Write(bufindex, bin.LittleEndian, self.deltaFilePosition)
    jsonmsg, _ := json.Marshal(self.history[i])
    bufdelta.Write(jsonmsg)
    bufdelta.WriteByte(0)
    self.deltaFilePosition += int64(len(jsonmsg)) + 1;
  }
  bufdelta.Flush()
  bufindex.Flush()
  self.commitVersion += int64(len(self.history)) - (self.commitVersion - self.firstInMemoryVersion)

  // Write snapshot
  self.fsnapshot.Seek(0, 0)
  jsonmsg, _ := json.Marshal(self.node.doc)
  bin.Write(self.fsnapshot, bin.LittleEndian, self.commitVersion)
  bin.Write(self.fsnapshot, bin.LittleEndian, int(len(jsonmsg)))
  self.fsnapshot.Write(jsonmsg)
}

func (self *DocumentHistory) initialRead() {
  self.deltaFilePosition = 0
  self.firstInMemoryVersion = 0
  self.commitVersion = 0
  self.broken = false
  
  if self.fdelta == nil || self.findex == nil || self.fsnapshot == nil {
    log.Println("Could not open one of the files")
    self.broken = true
    return
  }
  
  // Try to read snapshot
  var version int64
  var snaplen int
  var doc map[string]interface{}
  err := bin.Read(self.fsnapshot, bin.LittleEndian, &version)
  if err == nil {
    err := bin.Read(self.fsnapshot, bin.LittleEndian, &snaplen)
    if err == nil {
      bytes := make([]byte, snaplen)
      n, err := self.fsnapshot.Read(bytes)
      if n == snaplen {
        doc = make(map[string]interface{})
        err = json.Unmarshal(bytes, &doc)
        if err == nil {
          // Successfully loaded a snapshot
          self.firstInMemoryVersion = version
          self.commitVersion = version
        } else {
	  doc = nil
	}
      }
    }
  }
  // No snapshot?
  if doc == nil {
    doc = make(map[string]interface{})
    doc["_meta"] = make(map[string]interface{})
    doc["_data"] = make(map[string]interface{})
    doc["_rev"] = float64(0)
    // TODO
    doc["_hash"] = "TODOHASH"
  }
    
  /*
  truncate := (stat.Size % 8) * 8
  for truncate > 0 {
    self.findex.Seek(truncate - 8, 0)
    var pos int64
    err := bin.Read(self.findex, bin.LittleEndian, pos)
    if err != nil {
      truncate -= 8
      continue
    }
    // Seek the entry in the delta file
    n, err := self.fdelta.Seek(pos, 0)
    if err != nil || n != 8 {
      truncate -= 8
      continue
    }
    // TODO: Try to read the delta at this position    
    break
  }
  */
  
  // Read the deltas that follow the snapshot
  // if stat.Size != truncate {
  //  self.findex.Truncate(truncate)
  //}
  self.history, err = self.rangeFromDisk(self.firstInMemoryVersion, -1)
  if err != nil {
    log.Println(err)
    self.broken = true
    return
  }
  self.firstInMemoryVersion += int64(len(self.history))
  self.commitVersion = self.firstInMemoryVersion
  
  // Apply deltas to the snapshot
  for _, delta := range self.history {
    if !delta.Apply(doc, NoFlags) {
      log.Println("Failed applying delta")
      self.broken = true
      return
    }
  }
  self.node.doc = doc
  
  /*
  bytes := make([]byte, 4)
  var indexFilePosition int64 = 0
  for {
    n, err := self.bufindex.Read(bytes)
    if err == os.EOF {
      break
    }
    if err != nil {
      panic("Failed reading index file")
    }
    if n != 4 {
      findex.Seek(indexFilePosition, 0)
      findex.Truncate(indexFilePosition)
      self.bufindex = bufio.NewReadWriter(bufio.NewReader(self.findex), bufio.NewWriter(self.findex))
    } else {
      indexFilePosition += 4
    }
  }
  */
}

func (self *DocumentHistory) rangeFromDisk(startVersion int64, endVersion int64) ([]DocumentMutation, os.Error) {
  if endVersion < 0 {
    // How many entries are in the index file?
    stat, err := self.findex.Stat()
    if err != nil {
      return nil, os.NewError("Could not stat index file")
    }
    endVersion = stat.Size / 8
  }
  
  if startVersion == endVersion {
    return make([]DocumentMutation, 0), nil
  }

  n, err := self.findex.Seek(startVersion * 8, 0)
  if n != startVersion * 8 || err != nil {
    return nil, os.NewError("Failed to seek in index file")
  }
  var pos int64
  err = bin.Read(self.findex, bin.LittleEndian, &pos)
  if err != nil {
    return nil, os.NewError("Failed to read from index file")
  }
  n, err = self.fdelta.Seek(pos, 0)
  if n != pos || err != nil {
    return nil, os.NewError("Failed to seek in delta file")
  }
   
  result := make([]DocumentMutation, 0, endVersion - startVersion)
  buf := bufio.NewReader(self.fdelta)
  for i := startVersion; i < endVersion; i++ {
    str, err := buf.ReadSlice(0)
    if err != nil {
      return nil, os.NewError("Failed to read delta")
    }
    jsonmsg := make(map[string]interface{})
    err = json.Unmarshal(str[:len(str)-1], &jsonmsg)
    if err != nil {
      return nil, os.NewError("Failed parsing persisted delta")
    }
    result = append(result, DocumentMutation(jsonmsg))
  }  
  return result, nil
}

func (self *DocumentHistory) internalRange(startVersion int64, endVersion int64) (result []DocumentMutation, err os.Error) {
  // Error checking
  if startVersion < 0 || startVersion == self.firstInMemoryVersion + int64(len(self.history)) {
    return nil, os.NewError("Invalid version number")
  }
  if endVersion < 1 || endVersion > self.firstInMemoryVersion + int64(len(self.history)) {
    return nil, os.NewError("Invalid version number")
  }
  if endVersion <= startVersion {
    return nil, os.NewError("Invalid version number")
  }
  if startVersion < self.firstInMemoryVersion {
    result, err = self.rangeFromDisk(startVersion, self.firstInMemoryVersion)
    if err != nil {
      return nil, err
    }
  }
  result = append(result, self.history[:endVersion - self.firstInMemoryVersion]...)
  return result, nil
}
  
func (self *DocumentHistory) Append(mutation DocumentMutation) {
  self.history = append(self.history, mutation)
  // HACK
  self.Commit()
}

func (self *DocumentHistory) Tail(startVersion int64) []DocumentMutation {
  return self.history[startVersion:]
}

func (self *DocumentHistory) Range(startVersion int64, startHash string, endVersion int64, endHash string, limit int64) (result string, err os.Error) {
  // Error checking
  if startVersion < 0 || startVersion == self.firstInMemoryVersion + int64(len(self.history)) {
    return "", os.NewError("Invalid version number")
  }
  if endVersion < 1 || endVersion > self.firstInMemoryVersion + int64(len(self.history)) {
    return "", os.NewError("Invalid version number")
  }
  if endVersion <= startVersion {
    return "", os.NewError("Invalid version number")
  }
    
  // Get the history
  // mutations := self.history[startVersion:endVersion]
  mutations, err := self.internalRange(startVersion, endVersion)
  if err != nil {
    return "", err
  }
  if mutations[0].AppliedAtHash() != startHash {
    return "", os.NewError("Invalid hash")
  }
  if mutations[len(mutations)-1].ResultingHash() != endHash {
    return "", os.NewError("Invalid hash")
  }
  
  // Start encoding it as a JSON Array. Obeye the limit
  list := make([]string, 0, len(mutations) + 2)
  var count int64 = 0
  var bytes int64 = 0
  for _, m := range mutations {
    j, err := json.Marshal(m)
    if err != nil {
      panic("Could not encode my own data")
    }
    bytes += int64(len(j))
    if bytes > limit {
      break
    }
    list = append(list, string(j))
    count++
  }

  // Create response
  result = fmt.Sprintf(`{"appliedDeltas":[%v], "truncated":%v}`, strings.Join(list, ","), startVersion + count)
  return result, nil
}
