package wave

import (
  "os"
  proto "goprotobuf.googlecode.com/hg/proto"  
)

type WaveletHistory struct {
  history []*ProtocolWaveletDelta  
}

func NewWaveletHistory() *WaveletHistory {
  return &WaveletHistory{}
}

func (self *WaveletHistory) Append(mutation *ProtocolWaveletDelta) {
  self.history = append(self.history, mutation)
  for i := 1; i < len(mutation.Operation); i++ {
	self.history = append(self.history, nil)
  }
}

func (self *WaveletHistory) Range(startVersion int64, startHash []byte, endVersion int64, endHash []byte, limit int64) (result *ProtocolWaveletHistory, err os.Error) {
  // Error checking
  if startVersion < 0 || startVersion == int64(len(self.history)) {
	return nil, os.NewError("Invalid version number")
  }
  if endVersion < 1 || endVersion > int64(len(self.history)) {
	return nil, os.NewError("Invalid version number")
  }
  if endVersion <= startVersion {
	return nil, os.NewError("Invalid version number")
  }
	
  s := &ProtocolHashedVersion{Version:proto.Int64(startVersion), HistoryHash:startHash}
  e := &ProtocolHashedVersion{Version:proto.Int64(endVersion), HistoryHash:endHash}
  
  // Get the history
  mutations := self.history[startVersion:endVersion]
  if mutations[0] == nil || !mutations[0].HashedVersion.Equals(s) {
	return nil, os.NewError("Invalid hash or version")
  }
  if mutations[len(mutations)-1] == nil || mutations[len(mutations)-1].HistoryHash.Equals(e) {
	return nil, os.NewError("Invalid hash or version")
  }

  // Remove all items which are nil
  list := make([]*ProtocolWaveletDelta, 0, endVersion - startVersion)
  pos := 0
  for _, p := range mutations {
	if p == nil {
	  continue
	}
	list[pos] = p
	pos++
  }
  
  // TODO: Implement the limit
  result = &ProtocolWaveletHistory{}
  result.Deltas := list[:]
  return result, nil
}
