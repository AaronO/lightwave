package lightwave

import (
  "os"
  "json"
  "strings"
  "fmt"
)

type DocumentHistory struct {
  history []DocumentMutation  
}

func NewDocumentHistory() *DocumentHistory {
  return &DocumentHistory{}
}

func (self *DocumentHistory) Append(mutation DocumentMutation) {
  self.history = append(self.history, mutation)
}

func (self *DocumentHistory) Tail(startVersion int64) []DocumentMutation {
  return self.history[startVersion:]
}

func (self *DocumentHistory) Range(startVersion int64, startHash string, endVersion int64, endHash string, limit int64) (result string, err os.Error) {
  // Error checking
  if startVersion < 0 || startVersion == int64(len(self.history)) {
	return "", os.NewError("Invalid version number")
  }
  if endVersion < 1 || endVersion > int64(len(self.history)) {
	return "", os.NewError("Invalid version number")
  }
  if endVersion <= startVersion {
	return "", os.NewError("Invalid version number")
  }
	
  // Get the history
  mutations := self.history[startVersion:endVersion]
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
