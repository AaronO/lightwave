package lightwave

import (
  "testing"
  "log"
  "time"
)

type MyEntry struct {
  Key string
  Value string
  Tags []string
  Mappings map[DocumentMappingId]string
}

func (self *MyEntry) HasTag(tag string) bool {
  for _, t := range self.Tags {
    if t == tag {
      return true
    }
  }
  return false
}

type MyTester struct {
  Indexer *MemoryIndexer
  Entries map[string]*MyEntry
}

func NewMyTester() *MyTester {
  return &MyTester{Entries:make(map[string]*MyEntry)}
}

func (self *MyTester) Map(key string, mapping DocumentMappingId) {
  log.Println("Map key=", key, " mapping=", mapping)
  entry, ok := self.Entries[key]
  if !ok {
    return
  }
  entry.Mappings[mapping] = "yes"
  self.put(entry)
}

func (self *MyTester) Put(key string, value string, tags []string) {
  entry, ok := self.Entries[key]
  if !ok {
    entry = &MyEntry{Key:key, Value:value, Tags:tags, Mappings:make(map[DocumentMappingId]string)}
    self.Entries[key] = entry
  } else {
    entry.Value = value
    entry.Tags = tags
  }
  self.put(entry) 
}

func (self *MyTester) Delete(key string) {
  self.Entries[key] = nil, false
  self.Indexer.Delete(key)
}

func (self *MyTester) put(entry *MyEntry) {
  values := make(map[DocumentMappingId]interface{})
  for mapping, _ := range entry.Mappings {
    entry.Mappings[mapping] = string(mapping) + ".map(" + entry.Value + ")"
    values[mapping] = string(mapping) + ".map(" + entry.Value + ")"
  }
  self.Indexer.Put(entry.Key, values, entry.Tags)
}

type Listener struct {
  Name string
  values map[string]string
}

func NewListener(name string) *Listener {
  return &Listener{Name:name, values:make(map[string]string)}
}

func (self *Listener) AddResult(key string, value interface{}) {
  if _, ok := self.values[key]; ok {
    test.Errorf("Added a key that is already in the set: query %v, key: %v, value: %v", self.Name, key, value)
  }
  self.values[key] = value.(string)
  log.Println(self.Name, " Add Key:", key, " Value:", value.(string))
}

func (self *Listener) UpdateResult(key string, value interface{}) {
  if _, ok := self.values[key]; !ok {
    test.Errorf("Updated a key that is not in the set: query %v, key: %v, value: %v", self.Name, key, value)
  }
  self.values[key] = value.(string)
  log.Println(self.Name, " Update Key:", key, " Value:", value.(string))
}

func (self *Listener) DeleteResult(key string) {
  if _, ok := self.values[key]; !ok {
    test.Errorf("Deleted a key that is not in the set: query %v, key: %v", self.Name, key)
  }
  self.values[key] = "", false
  log.Println(self.Name, " Delete Key:", key)
}

func (self *Listener) Verify(tester *MyTester, mapping DocumentMappingId, hasTags []string, hasNotTags []string) {
  result := make(map[string]string)
  for key, entry := range tester.Entries {
    ok := true
    for _, tag := range hasTags {
      if !entry.HasTag(tag) {
	ok = false
	break
      }
    }
    for _, tag := range hasNotTags {
      if entry.HasTag(tag) {
	ok = false
	break
      }
    }
    if ok {
      result[key] = entry.Mappings[mapping]
    }
  }
  
  if len(result) != len(self.values) {
    test.Errorf("Result set has not the expected size: query: %v, is:%v, should-be:%v", self.Name, self.values, result)
    return
  }
  for key, value := range result {
    v, ok := self.values[key]
    if !ok {
      test.Errorf("The result set is missing the key %v", key)
      return
    }
    if v != value {
      test.Errorf("The result set the wrong value for key: %v, is: %v, should-be:%v", key, v, value)
      return      
    }
  }
}

var test *testing.T

func TestIndexer(t *testing.T) {
  test = t
  tester := NewMyTester()
  indexer := NewMemoryIndexer(tester)
  tester.Indexer = indexer
  go indexer.Run()
  
  tester.Put("/foo", "Hallo Foo", []string{"dumb","dumm"})
  tester.Put("/bar", "Hallo Bar", []string{"stupid","dumm", "dumb"})
  listener := NewListener("q1")
  indexer.Query("q1", listener, "mapone", []string{"dumm"}, []string{})
  listener2 := NewListener("q2")
  indexer.Query("q2", listener2, "maptwo", []string{"stupid"},[]string{})
  tester.Put("/later", "This one comes later", []string{"stupid","dumm", "dumb"})
  tester.Put("/another", "This one does not match", []string{"stupid","dumb"})
  tester.Put("/later", "New value", []string{"stupid","dumm", "dumb"})
  tester.Put("/later", "Another value", []string{"stupid", "dumb"})
  tester.Delete("/bar")
  //tester.Delete("/foo")
  //tester.Delete("/later")
  //tester.Delete("/another")

  time.Sleep(2000000000)

  listener.Verify(tester, "mapone", []string{"dumm"}, []string{})
  listener2.Verify(tester, "maptwo", []string{"stupid"}, []string{})   
}