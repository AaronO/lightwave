package lightwave

import (
  "log"
  "time"
  sqlite "gosqlite.googlecode.com/hg/sqlite"
)

// ----------------------------------------------------

type DocumentMapper interface {
  Map(key string, mapping DocumentMappingId)
}

// ----------------------------------------------------
// DocumentMappingId

type DocumentMappingId string

// ----------------------------------------------------
// QueryListener

type QueryListener interface {
  AddResult(queryId string, key string, value interface{})
  DeleteResult(queryId string, key string)  
}

// ----------------------------------------------------
// Query

type Query struct {
  Id string
  PrimaryTag string
  Mapping DocumentMappingId
  HasTags []string
  HasNotTags []string
  Listener QueryListener
}

func (self *Query) AddResult(key string, value interface{}) {
  self.Listener.AddResult(self.Id, key, value)
}

func (self *Query) DeleteResult(key string) {
  self.Listener.DeleteResult(self.Id, key)
}

// ----------------------------------------------------
// Internal messages

type putMsg struct {
  key string
  value map[DocumentMappingId]interface{}
  tags []string
}

type deleteMsg struct {
  key string
}

type queryMsg struct {
  queryId string
  listener QueryListener
  mapping DocumentMappingId
  hasTags []string
  hasNotTags []string
}

type stopQueryMsg struct {
  queryId string
}

// ----------------------------------------------------
// MemoryIndexer

type MemoryIndexer struct {
  mapper DocumentMapper
  indices map[string]*MemoryIndex
  entries map[string]*MemoryIndexEntry
  /**
   * The keys of entries that have been modified and not yet persisted to disk
   */
  modifiedEntries map[string]bool
  queries map[string]*Query
  channel chan interface{}
  diskIndexer *DiskIndexer
  ticker *time.Ticker
}

func NewMemoryIndexer(mapper DocumentMapper, diskIndexer *DiskIndexer) *MemoryIndexer {
  r := &MemoryIndexer{mapper:mapper, indices:make(map[string]*MemoryIndex), entries:make(map[string]*MemoryIndexEntry), queries:make(map[string]*Query), channel:make(chan interface{}, 1000)}
  r.diskIndexer = diskIndexer
  r.modifiedEntries = make(map[string]bool)
  r.ticker = time.NewTicker(1000000000 * 10)
  return r
}

func (self *MemoryIndexer) Run() {
  for {
    select {
    case msg := <-self.channel:
      switch msg.(type) {
      case putMsg:
	m := msg.(putMsg)
	log.Println("INDEXER put ", m)
	self.put(m.key, m.value, m.tags)
      case deleteMsg:
	m := msg.(deleteMsg)
	self.delete(m.key)
      case queryMsg:
	m := msg.(queryMsg)
	self.query(m.queryId, m.listener, m.mapping, m.hasTags, m.hasNotTags)
      case stopQueryMsg:
	m := msg.(stopQueryMsg)
	self.stopQuery(m.queryId)
      default:
	log.Println("Received message of unsupported type")
      }
    case <-self.ticker.C:
      self.writeToDisk()
    }
  }
}

func (self *MemoryIndexer) writeToDisk() {
  for key, _ := range self.modifiedEntries {
    entry, ok := self.entries[key]
    if !ok {
      self.diskIndexer.Delete(key)
    } else {
      self.diskIndexer.Put(key, entry.Value, entry.Tags)
    }
  } 
  self.modifiedEntries = make(map[string]bool)
  log.Println("INDEX persisted")
}

func (self *MemoryIndexer) Put(key string, value map[DocumentMappingId]interface{}, tags []string) {
  // Create a copy of tags and values.
  // This is required, because otherwise the values might change and we have race conditions
  newtags := make([]string, len(tags))
  copy(newtags, tags)
  newvalues := make(map[DocumentMappingId]interface{})
  for k, v := range value {
    newvalues[k] = v
  }
  self.channel <- putMsg{key:key, value:newvalues, tags:newtags}
}

func (self *MemoryIndexer) Delete(key string) {
  self.channel <- deleteMsg{key:key}
}

func (self *MemoryIndexer) Query(queryId string, listener QueryListener, mapping DocumentMappingId, hasTags []string, hasNotTags []string) {
  self.channel <- queryMsg{queryId:queryId, listener:listener, mapping:mapping, hasTags:hasTags, hasNotTags:hasNotTags}
}

func (self *MemoryIndexer) StopQuery(queryId string) {
  self.channel <- stopQueryMsg{queryId:queryId}
}

func (self *MemoryIndexer) put(key string, value map[DocumentMappingId]interface{}, tags []string) {
  self.modifiedEntries[key] = true
  var oldentry *MemoryIndexEntry = nil
  entry, ok := self.entries[key]
  if !ok {
    // Create a new entry
    entry = &MemoryIndexEntry{Key:key, Value:value, Tags:tags}
    self.entries[key] = entry
  } else {
    oldentry = &MemoryIndexEntry{Key:key, Value:entry.Value, Tags:entry.Tags}
  }

  // Fill in the tags for the updated entry
  entry.Value = value
  entry.Tags = tags
  
  for _, tag := range tags {
    index, indexok := self.indices[tag]
    if !indexok {
      // Create the index
      index = NewMemoryIndex(self, tag)
      self.indices[tag] = index
    }
    index.Put(entry, oldentry)
  }
  // Remove entry from indexes where it no longer belongs
  if oldentry != nil {
    for _, tag := range oldentry.Tags {
      if !entry.HasTag(tag) {
	if index, ok := self.indices[tag]; ok {
	  index.Delete(oldentry)
	}
      }
    }
  }
}

func (self *MemoryIndexer) delete(key string) {
  entry, ok := self.entries[key]
  if !ok {
    return
  }
  for _, tag := range entry.Tags {
    if index, ok := self.indices[tag]; ok {
      index.Delete(entry)
    }
  }
  self.entries[entry.Key] = nil, false
}

func (self *MemoryIndexer) query(queryId string, listener QueryListener, mapping DocumentMappingId, hasTags []string, hasNotTags []string) {
  query := &Query{HasTags:hasTags, HasNotTags:hasNotTags, Listener:listener}
  // Assign an ID
  query.Id = queryId
  query.Mapping = mapping
  self.queries[query.Id] = query
  // Determine the primary index
  if len(query.HasTags) == 0 {
    log.Println("Query has empty HasTags array")
    return
  }
  var min int = 0x7fffffff
  minTag := ""
  var minIndex *MemoryIndex
  for _, tag := range query.HasTags {
    index, ok := self.indices[tag]
    if !ok {
      // Create the index
      index = NewMemoryIndex(self, tag)
      self.indices[tag] = index
    }
    if index.Len() < min {
      min = index.Len()
      minTag = tag
      minIndex = index
    }
  }
  query.PrimaryTag = minTag

  // Query the persistent storage and then overlay it with the in-memory database
  diskResult := self.diskIndexer.Query(query.Mapping, query.HasTags, query.HasNotTags)
  // Remove all entries which are in the in-memory database since these are fresher
  for key, _ := range diskResult {
    if _, ok := self.entries[key]; ok {
      diskResult[key] = "", false
    }
  }
  // Send all query results straight back to the caller
  for key, value := range diskResult {
    query.AddResult(key, value)
  }
  // Register at the primary index. This will query the in-memory database
  minIndex.StartQuery(query, true)
}

func (self *MemoryIndexer) stopQuery(queryId string) {
  query, ok := self.queries[queryId]
  if !ok {
    return
  }
  self.queries[queryId] = nil, false
  index, ok := self.indices[query.PrimaryTag]
  if !ok {
    return
  }
  index.StopQuery(query)
}

func (self *MemoryIndexer) RequestMapping(key string, mapping DocumentMappingId) {
  self.mapper.Map(key, mapping)
}

// ----------------------------------------------------
// MemoryIndex

type MemoryIndex struct {
  indexer *MemoryIndexer
  tag string
  entries map[string]*MemoryIndexEntry
  queries map[string]*Query
}

func NewMemoryIndex(indexer *MemoryIndexer, tag string) *MemoryIndex {
  index := &MemoryIndex{indexer:indexer, tag:tag, queries:make(map[string]*Query), entries:make(map[string]*MemoryIndexEntry)}
  return index
}

func (self *MemoryIndex) Len() int {
  return len(self.entries)
}

func (self *MemoryIndex) StartQuery(query* Query, execute bool) {
  self.queries[query.Id] = query;
  // Execute the query
  if execute {
    for key, entry := range self.entries {
      if entry.Match(query) {
	// Add to the result set
	if entry.HasMapping(query.Mapping) {
	  query.AddResult(key, entry.Value[query.Mapping])
	} else {
	  self.indexer.RequestMapping(key, query.Mapping)
	}
      }
    }
  }
}

func (self *MemoryIndex) StopQuery(query* Query) {
  self.queries[query.Id] = nil, false
}

func (self *MemoryIndex) Put(entry *MemoryIndexEntry, oldentry *MemoryIndexEntry) {
  self.entries[entry.Key] = entry
  // Does the new entry somehow affect the registered queries?
  for _, query := range self.queries {
    if entry.Match(query) {
      if entry.HasMapping(query.Mapping) {
	// Send only if the value has changed
	if oldentry == nil || !oldentry.HasMapping(query.Mapping) || entry.Value[query.Mapping] != oldentry.Value[query.Mapping] {
	  // Send update to the query listener
	  query.AddResult(entry.Key, entry.Value[query.Mapping])
	}
      } else {
	self.indexer.RequestMapping(entry.Key, query.Mapping)
      }
    }
  }
}

func (self *MemoryIndex) Delete(entry *MemoryIndexEntry) {
  self.entries[entry.Key] = nil, false
  // The entry was part of a query? Then notify it
  for _, query := range self.queries {
    if entry.Match(query) {
      // Send update to the query listeneristener
      query.DeleteResult(entry.Key)
    }
  }
}

// ----------------------------------------------------
// MemoryIndexEntry

type MemoryIndexEntry struct {
  Key string
  Value map[DocumentMappingId]interface{}
  Tags []string
}

func (self *MemoryIndexEntry) HasTag(tag string) bool {
  for _, t := range self.Tags {
    if t == tag {
      return true
    }
  }
  return false
}

func (self *MemoryIndexEntry) HasMapping(mapping DocumentMappingId) bool {
  _, ok := self.Value[mapping]
  return ok
}

func (self *MemoryIndexEntry) Match(query* Query) bool {
  for _, tag := range query.HasTags {
    if !self.HasTag(tag) {
      return false
    }
  }
  for _, tag := range query.HasNotTags {
    if self.HasTag(tag) {
      return false
    }
  }
  return true
}

// -----------------------------------------------------
// DiskIndexer

type DiskIndexer struct {
  dbcon *sqlite.Conn
}

func NewDiskIndexer(dbname string) *DiskIndexer {
  r := &DiskIndexer{}
  dbcon, err := sqlite.Open(dbname)
  if err != nil {
    panic("Cannot access user database")
  }
  r.dbcon = dbcon
  stmnt, err := r.dbcon.Prepare("CREATE TABLE IF NOT EXISTS Digest ( key VARCHAR(30), mapping VARCHAR(30), digest VARCHAR(1000) )")
  if err != nil {
    panic("Cannot prepare stmnt for create account table")
  }
  err = stmnt.Exec()
  if err != nil {
    panic("Cannot create account table")
  }
  stmnt.Next()

  stmnt, err = r.dbcon.Prepare("CREATE TABLE IF NOT EXISTS Tags ( key VARCHAR(30), tag VARCHAR(1000) )")
  if err != nil {
    panic("Cannot prepare stmnt for create account table")
  }
  err = stmnt.Exec()
  if err != nil {
    panic("Cannot create account table")
  }
  stmnt.Next()

/*  stmnt, err = r.dbcon.Prepare("CREATE INDEX IF NOT EXISTS TagIndex ON Tags ( tag )")
  if err != nil {
    panic("Cannot prepare stmnt for create account table")
  }
  err = stmnt.Exec()
  if err != nil {
    panic("Cannot create account table")
  }
  stmnt.Next()
*/
  
  return r
}

func (self *DiskIndexer) Delete(key string) {
  stmnt, err := self.dbcon.Prepare("DELETE FROM Digest WHERE key = ?1")
  if err != nil {
    panic(err.String())
  }
  err = stmnt.Exec(key)
  if err != nil {
    panic(err.String())
  }  
  stmnt.Next()
  
  stmnt, err = self.dbcon.Prepare("DELETE FROM Tags WHERE key = ?1")
  if err != nil {
    panic(err.String())
  }
  err = stmnt.Exec(key)
  if err != nil {
    panic(err.String())
  }  
  stmnt.Next()
}

func (self *DiskIndexer) Put(key string, values map[DocumentMappingId]interface{}, tags []string) {
  self.Delete(key)
  log.Println("DB WRITE ", key, values, tags)
  
  stmnt, err := self.dbcon.Prepare("INSERT INTO Digest VALUES ( ?1, ?2, ?3 )")
  if err != nil {
    panic(err.String())
  }
  for mapping, value := range values {
    err = stmnt.Exec(key, string(mapping), value)
    if err != nil {
      panic(err.String())
    }  
    stmnt.Next()
  }

  stmnt, err = self.dbcon.Prepare("INSERT INTO Tags VALUES ( ?1, ?2 )")
  if err != nil {
    panic(err.String())
  }
  for _, tag := range tags {
    err = stmnt.Exec(key, tag)
    if err != nil {
      panic(err.String())
    }  
    stmnt.Next()
  }
}

func (self *DiskIndexer) Query(mapping DocumentMappingId, tags []string, noTags []string) map[string]string{
  sql := "SELECT key, digest FROM Digest WHERE mapping = ?1"
  for _, tag := range tags {
    sql += " AND key IN ( SELECT key FROM Tags WHERE tag = \"" + tag + "\""
  }
  for _, tag := range noTags {
    sql += " AND NOT EXISTS ( SELECT key FROM Tags WHERE tag = \"" + tag + "\" )"
  }
  for i := 0; i < len(tags); i++ {
    sql += ")"
  }
  log.Println(sql, mapping)
  stmnt, err := self.dbcon.Prepare(sql)
  if err != nil {
    panic(err.String())
  }
  err = stmnt.Exec(string(mapping))
  if err != nil {
    panic(err.String())
  }
  result := make(map[string]string)
  for stmnt.Next() {
    var key string
    var digest string
    err = stmnt.Scan(&key, &digest)
    if err != nil {
      panic(err.String())
    }  
    log.Println("DB READ", key, digest)
    result[key] = digest
  }
  return result
}