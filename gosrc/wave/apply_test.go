package wave

import (
  "testing"
)

func TestURL(t *testing.T) {
  u, err := NewWaveUrl("wave://foo.com/w+abc/conv+root")
  if err != nil {
	t.Errorf("Parsing of wave url failed")
  }
  if u.WaveId != "w+abc" {
	t.Errorf("Wrong wave id")
  }
  if u.WaveDomain != "foo.com" {
	t.Errorf("Wrong wave domain")
  }
  if u.WaveletId != "conv+root" {
	t.Errorf("Wrong wavelet id")
  }
  if u.WaveletDomain != "foo.com" {
	t.Errorf("Wrong wavelet domain")
  }
}

func opCharacters(str string) *ProtocolDocumentOperation_Component {
  c := &ProtocolDocumentOperation_Component{Characters:&str}
  return c
}

func opElementStart(typ string) *ProtocolDocumentOperation_Component {
  attr := [...]*ProtocolDocumentOperation_Component_KeyValuePair{}[:]
  c := &ProtocolDocumentOperation_Component{ElementStart:&ProtocolDocumentOperation_Component_ElementStart{Type:&typ,Attribute:attr}}
  return c
}

func opElementEnd() *ProtocolDocumentOperation_Component {
  b := true
  c := &ProtocolDocumentOperation_Component{ElementEnd:&b}
  return c
}

func opRetain(count int32) *ProtocolDocumentOperation_Component {
  c := &ProtocolDocumentOperation_Component{RetainItemCount:&count}
  return c
}

func TestApply(t *testing.T) {
  u, err := NewWaveUrl("wave://foo.com/w+abc/conv+root")
  if err != nil {
	t.Errorf("Parsing of wave url failed")
  }
  // Create a wavelet
  w := NewWavelet(u)
  // Create an empty document
  doc := w.Document("b+1")
  
  // Create a mutation
  docop := &ProtocolDocumentOperation{}
  docop.Component = [...]*ProtocolDocumentOperation_Component{ opElementStart("p"), opCharacters("Hallo Welt"), opElementEnd() }[:]
  err = docop.ApplyTo(doc)
  if err != nil {
	t.Errorf("Applying delta failed: %v", err)
  }
  if doc.Text() != "Hallo Welt" {
	t.Errorf("Applying delta resulted in wrong document content")
  }

  // Create a mutation
  docop = &ProtocolDocumentOperation{}
  docop.Component = [...]*ProtocolDocumentOperation_Component{ opRetain(1), opCharacters("Wow! "), opRetain(11) }[:]
  err = docop.ApplyTo(doc)
  if err != nil {
	t.Errorf("Applying delta failed: %v", err)
  }
  if doc.Text() != "Wow! Hallo Welt" {
	t.Errorf("Applying delta resulted in wrong document content")
  }
  
  // Create a mutation
  docop = &ProtocolDocumentOperation{}
  docop.Component = [...]*ProtocolDocumentOperation_Component{ opRetain(6), opElementStart("line"), opElementEnd(), opRetain(11) }[:]
  err = docop.ApplyTo(doc)
  if err != nil {
	t.Errorf("Applying delta failed: %v", err)
  }
  if doc.Text() != "Wow! Hallo Welt" {
	t.Errorf("Applying delta resulted in wrong document content")
  }  
}
