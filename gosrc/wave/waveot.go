package wave

import (
  "http"
  "os"
  "strings"
  vec "container/vector"
  proto "goprotobuf.googlecode.com/hg/proto"  
)

//----------------------------------------------------------------------
// WaveUrl

type WaveUrl struct {
  WaveId string
  WaveDomain string
  WaveletId string
  WaveletDomain string
}

func NewWaveUrl(url string) (result *WaveUrl, err os.Error) {
  u, err := http.ParseURL(url)
  if err != nil {
	return nil, err
  }
  if u.Scheme != "wave" {
	return nil, os.NewError("Not a wave url")
  }
  result = &WaveUrl{WaveletDomain: u.Host}
  
  wave := u.Path[1:]
  i := strings.Index(wave, "/")
  if i == -1 || i == len(wave) - 1 {
	return nil, os.NewError("Malformed wave url")
  }
  result.WaveletId = wave[i+1:]
  wave = wave[:i]
  i = strings.Index(wave, "$")
  if i == -1 {
      result.WaveDomain = result.WaveletDomain;
      result.WaveId = wave;
  } else {
      result.WaveDomain = wave[:i]
      result.WaveId = wave[i+1:]
  }    
  // TODO: Check that only allowed characters are used
  return result, nil
}

func NewWaveUrlFromId(id string) (result *WaveUrl, err os.Error) {
  split := strings.Split(id, "/", -1)
  if len(split) != 4 {
	return nil, os.NewError("Malformed wavelet-id")
  }
  result = &WaveUrl{}
  result.WaveDomain = split[0]
  result.WaveId = split[1]
  result.WaveletDomain = split[2]
  result.WaveletId = split[3]
  // TODO: Check that only allowed characters are used
  return result, nil
}

/**
 * @return a URL of the form wave://waveletDomain/waveDomain$waveId/waveletId or null.
 */
func (self *WaveUrl) String() string {
  var str = "wave://" + self.WaveletDomain + "/";
  if self.WaveDomain == self.WaveletDomain {
    str += self.WaveId
  } else {
    str += self.WaveDomain + "$" + self.WaveId
  }
  str += "/" + self.WaveletId;
  return str
}

//----------------------------------------------------------------------
// Wavelet

type Wavelet struct {
  Url *WaveUrl
  Documents map[string]*WaveletDocument
  /**
   * The participants of the wave.
   */
  Participants vec.StringVector
  /**
   * The current version of the wavelet
   */
  HashedVersion ProtocolHashedVersion
}

func NewWavelet(url *WaveUrl) (result *Wavelet) {
  result = &Wavelet{Url:url}
  result.HashedVersion.HistoryHash = []byte( url.String() )
  result.HashedVersion.Version = proto.Int64(0)
  result.Documents = make(map[string]*WaveletDocument)
  return result
}

/**
 * Gets or creates a wavelet document.
 *
 * @param docname is a document name such as "b+124".
 * @param return an instance of WaveletDocument
 */
func (self *Wavelet) Document(docname string) *WaveletDocument {
  if doc, ok := self.Documents[docname]; ok {
	return doc
  }
  doc := NewWaveletDocument(docname, self)
  self.Documents[docname] = doc
  return doc
}

func (self *Wavelet) HasParticipant(jid string) bool {
  for _, p := range self.Participants {
	if p == jid {
	  return true
	}
 }
  return false
}

func (self *Wavelet) AddParticipant(jid string) bool {
  if self.HasParticipant(jid) {
	return true
  }
  self.Participants.Push(jid)
  return true
}

func (self *Wavelet) RemoveParticipant(jid string) bool {
  for i, p := range self.Participants {
	if p == jid {
	  self.Participants.Delete(i)
	  return true
	}
  }
  return false
}

/**
 * Applies a delta to this wavelet.
 *
 * @param delta is of type protocol.ProtocolWaveletDelta.
 */
func (self *Wavelet) ApplyDelta( delta *ProtocolWaveletDelta ) os.Error {
  for _, op := range delta.Operation {  
    if op.AddParticipant != nil {
      if !self.AddParticipant( *op.AddParticipant ) {
		return os.NewError("Could not add participant")
	  }
    }
    if op.RemoveParticipant != nil {
      if !self.RemoveParticipant( *op.RemoveParticipant ) {
		return os.NewError("Could not remove participant")
	  }
    }
    if op.MutateDocument != nil {
	  doc := self.Document(*op.MutateDocument.DocumentId)
      docop := op.MutateDocument.DocumentOperation
      if err := docop.ApplyTo( doc ); err != nil {
		return err
	  }
    }    
  }
  return nil
}

//----------------------------------------------------------
// WaveletDocument

type DocumentFormat map[string]string

type WaveletDocument struct {
  Wavelet *Wavelet
  DocId string
  Content vec.Vector
}

func NewWaveletDocument(docid string, wavelet* Wavelet) (result *WaveletDocument) {
  result = &WaveletDocument{Wavelet:wavelet, DocId:docid}
  return result
}

func (self *WaveletDocument) Text() string {  
  t := ""
  for _, x := range self.Content {
	if IsTextNode(x) {
	  t += x.(*TextNode).Text()
	}
  }
  return t
}

//----------------------------------------------------------
// TextNode

type TextNode struct {
  Document *WaveletDocument
  Format DocumentFormat
  text string
}

func (self *TextNode) Text() string {
  return self.text
}

func (self *TextNode) InsertText( pos int, str string ) {
  self.text = self.text[:pos] + str + self.text[pos:]
}

func (self *TextNode) AppendText( str string ) {
  self.text = self.text + str
}

func (self *TextNode) RemoveText( pos int, count int ) {
  self.text = self.text[:pos] + self.text[pos + count:]
}

func (self *TextNode) Split( pos int ) (n1, n2 *TextNode) {
  n1 = &TextNode{Document:self.Document, Format:self.Format, text:self.text[:pos]}
  n2 = &TextNode{Document:self.Document, Format:self.Format, text:self.text[pos:]}
  return n1, n2
}

func IsTextNode(n interface{}) bool {
  if n == nil {
	return false
  }
  if _, ok := n.(*TextNode); ok {
	return true
  }
  return false
}

//----------------------------------------------------------
// ElementStart

type ElementStart struct {
  Document *WaveletDocument
  Format DocumentFormat
  Attributes map[string]string
  Type string
}

//----------------------------------------------------------
// ElementEnd

type ElementEnd struct {
  Document *WaveletDocument
  Format DocumentFormat
}

// ---------------------------------------------------------
// ProtocolHashedVersion

func (self *ProtocolHashedVersion) Equals(v *ProtocolHashedVersion) bool {
  if *self.Version != *v.Version {
	return false
  }
  if len(self.HistoryHash) != len(v.HistoryHash) {
	return false
  }
  for i, b := range self.HistoryHash {
	if b != v.HistoryHash[i] {
	  return false
	}
  }
  return true
}

//----------------------------------------------------------
// ProtocolDocumentOperation

/**
 * Applies the document operation to a wavelet document.
 */
func (self *ProtocolDocumentOperation) ApplyTo(doc *WaveletDocument) (err os.Error) {
  // Position in doc.content
  contentIndex := 0
  textIndex := 0
  // Increased by one whenever a delete_element_start is encountered and decreased by 1 upon delete_element_end
  deleteDepth := 0
  insertDepth := 0
  
  // The annotation that applies to the left of the current document position, type is a dict
  // var docAnnotation DocumentFormat
  // var updatedAnnotation DocumentFormat
  // The annotation update that applies to the right of the current document position  
  // var annotationUpdate DocumentFormat        
  var c interface{}
  
  // Loop until all ops are processed
  for _, op := range self.Component {
	
	if contentIndex < len(doc.Content) {
	  c = doc.Content[contentIndex]
	} else {
	  c = nil
	}
	
    if op.ElementStart != nil {
	  //
	  // Insert element start
	  //
      if deleteDepth > 0 {
        return os.NewError("Cannot insert inside a delete sequence")
	  }

	  // Inserting in the middle of a text node?
	  if c != nil && textIndex > 0 {
		t := c.(*TextNode)
		t1, t2 := t.Split(textIndex)
		doc.Content.Set( contentIndex, t1 )
		doc.Content.Insert( contentIndex + 1, t2 )
		contentIndex++
		textIndex = 0
	  }
	  
	  // Create a map of attributes
      attribs := make(map[string]string)
      for _, a := range op.ElementStart.Attribute {
        attribs[*a.Key] = *a.Value;
      }

	  // Insert the element
	  element := &ElementStart{Document:doc, Attributes:attribs}
	  doc.Content.Insert( contentIndex, element )
	  contentIndex++
      insertDepth++;
	  
	  // Compute annotation
      // TODO updatedAnnotation = this.computeAnnotation(docAnnotation, annotationUpdate, annotationUpdateCount);
    } else if op.ElementEnd != nil {
	  //
	  // Insert element end
	  //
      if deleteDepth > 0 {
        return os.NewError("Cannot insert inside a delete sequence")
	  }
      if insertDepth == 0 {
        return os.NewError("Cannot delete element end without deleting element start")
	  }
	  if textIndex != 0 {
		panic("textIndex should be 0 when inserting element end")
	  }

	  // Insert the element end
	  element := &ElementEnd{Document:doc}
	  doc.Content.Insert( contentIndex, element )
	  contentIndex++
	  insertDepth--;

	  // Compute annotation
      // TODO updatedAnnotation = this.computeAnnotation(docAnnotation, annotationUpdate, annotationUpdateCount);
    } else if op.Characters != nil {
	  //
	  // Insert characters
	  //
      if deleteDepth > 0 {
        return os.NewError( "Cannot insert inside a delete sequence" )
	  }
      if len(*op.Characters) == 0 {
        continue
	  }
      
	  // Compute annotation
      // TODO updatedAnnotation = this.computeAnnotation(docAnnotation, annotationUpdate, annotationUpdateCount);
	  
	  // Insert to the right of a text node?
	  if contentIndex > 0 && textIndex == 0 && IsTextNode(doc.Content[contentIndex-1]) {
		t := doc.Content[contentIndex-1].(*TextNode)
		t.AppendText( *op.Characters )
	  } else if IsTextNode(c) {
		c.(*TextNode).InsertText( textIndex, *op.Characters )
		textIndex += len(*op.Characters)
	  } else {
		t := &TextNode{Document:doc, text:*op.Characters}
		doc.Content.Insert( contentIndex, t )
		contentIndex++
		textIndex = 0
	  }
    } else if op.DeleteCharacters != nil {      
	  //
	  // Delete characters
	  //
      if insertDepth > 0 {
        return os.NewError("Cannot delete inside an insertion sequence")
	  }
      if !IsTextNode(c) {
		return os.NewError("Operation expected characters in the document")
	  }
  	  
  	  t := c.(*TextNode)
	  if t.text[textIndex:textIndex + len(*op.DeleteCharacters)] != *op.DeleteCharacters {
		return os.NewError("Cannot delete characters here, because at this position there are no characters")
	  }
	  t.RemoveText( textIndex, len(*op.DeleteCharacters))
	  // Reached end of text node?
	  if textIndex == len(t.text) {
		// Deleted the entire text node ?
		if textIndex == 0 {
		  doc.Content.Delete(contentIndex)
		} else {
		  textIndex = 0
		  contentIndex++
		}
	  }
    } else if op.RetainItemCount != nil {
	  //
	  // Retain
	  //
      if insertDepth > 0 {
        return os.NewError( "Cannot retain inside an insertion sequence" )
	  }
      if c == nil {
        return os.NewError( "document op is larger than doc" )
	  }
      if deleteDepth > 0 {
        return os.NewError( "Cannot retain inside a delete sequence" )
	  }
            
      for count := int32(0); count < *op.RetainItemCount; count++ {
		if c == nil {
          return os.NewError( "document op is larger than doc" )
		}

		/* TODO
		docAnnotation = doc.format[contentIndex];
		updatedAnnotation = this.computeAnnotation(docAnnotation, annotationUpdate, annotationUpdateCount);
		// Update the annotation
		doc.format[contentIndex] = updatedAnnotation;
		*/

        if _, ok := c.(*ElementStart); ok {
		  // Retaining an element start
		  contentIndex++
		  if contentIndex < len(doc.Content) {
			c = doc.Content[contentIndex]
		  } else {
			c = nil
		  }
        } else if _, ok = c.(*ElementEnd); ok {        
		  // Retaining an element end?
		  contentIndex++
		  if contentIndex < len(doc.Content) {
			c = doc.Content[contentIndex]
		  } else {
			c = nil
		  }
        } else {
		  t := c.(*TextNode)
		  textIndex++
		  if textIndex == len(t.text) {
			textIndex = 0
			contentIndex++
			if contentIndex < len(doc.Content) {
			  c = doc.Content[contentIndex]
			} else {
			  c = nil
			}
		  }
		}
      }
    } else if op.DeleteElementStart != nil {
	  //
	  // Delete element start
	  //
      if insertDepth > 0 {
        return os.NewError( "Cannot delete inside an insertion sequence" )
	  }
      if c == nil {
		return os.NewError( "Document is shorter than op")
	  }
	  s, ok := c.(*ElementStart)
	  if !ok {
        return os.NewError( "Cannot delete element start at this position, because in the document there is none" )
	  }	  
      if s.Type != *op.DeleteElementStart.Type {
        return os.NewError( "Cannot delete element start because Op and Document have different element type" )
	  }
	  if len(s.Attributes) != len(op.DeleteElementStart.Attribute) {
		return os.NewError("Number of attributes does not match")
	  }

	  // Create a dictionary of attributes
      attribs := make(map[string]string)
      for _, a := range op.DeleteElementStart.Attribute {
        attribs[*a.Key] = *a.Value;
      }
      // Compare attributes from the doc and docop. They should be equal
      for k, v := range s.Attributes {      
		cmp, ok := attribs[k]
		if !ok {
		  return os.NewError("Attribute values do not match")
		}
        if v != cmp {
          return os.NewError( "Cannot delete element start because attribute values differ" )
		}
      }
      
      // Delete it
      doc.Content.Delete( contentIndex )	  
	  // Count how many opening elements have been deleted. The corresponding closing elements must be deleted, too.
      deleteDepth++;
    } else if op.DeleteElementEnd != nil {
	  //
	  // Delete element end
	  //
      if insertDepth > 0 {
        return os.NewError( "Cannot delete inside an insertion sequence" )
	  }
      if c == nil {
		return os.NewError( "Document is shorter than op")
	  }
	  _, ok := c.(*ElementEnd)
	  if !ok {
        return os.NewError( "Cannot delete element end at this position, because in the document there is none" )
	  }	  
      // Is there a matching openeing element?
      if deleteDepth == 0 {
        return os.NewError( "Cannot delete element end, because matching delete element start is missing" )
	  }
	  /* TODO
	  // If there is an annotation boundary change in the deleted characters, this change must be applied
      var anno = doc.format[contentIndex];
      if ( anno != docAnnotation )
      {
        docAnnotation = anno;
        updatedAnnotation = this.computeAnnotation(docAnnotation, annotationUpdate, annotationUpdateCount);
      }
	  */

      // Delete it
      doc.Content.Delete( contentIndex )	  
      deleteDepth--;
    } else if op.UpdateAttributes != nil {
	  //
	  // Update Attributes
	  //
      if insertDepth > 0 || deleteDepth > 0 {
        return os.NewError( "Cannot update attributes inside an insertion sequence" )
	  }
      if c == nil {
		return os.NewError( "Document is shorter than op")
	  }
	  s, ok := c.(*ElementStart)
	  if !ok {
        return os.NewError( "Cannot update at this position, because in the document there is no element start" )
	  }	  
      
      /* TODO
	  // Compute the annotation for this element start
      var anno = doc.format[contentIndex];
      if ( anno != docAnnotation )
      {
        docAnnotation = anno;
        updatedAnnotation = this.computeAnnotation(docAnnotation, annotationUpdate, annotationUpdateCount);
      }
      doc.format[contentIndex] = updatedAnnotation;
      */
	  
      // Update the attributes
      for _, update := range op.UpdateAttributes.AttributeUpdate {
        // Add a new attribute?
        if update.OldValue == nil {
          if _, ok := s.Attributes[*update.Key]; ok {
            return os.NewError( "Cannot update attributes because old attribute value is not mentioned in Op" )
		  }
        } else {
		  // Delete or change an attribute
		  v, ok := s.Attributes[*update.Key]
		  if !ok || v != *update.OldValue {
			return os.NewError("Cannot update attributes because old attribute value does not match with Op")
		  }
		}
		if update.NewValue == nil {
		  s.Attributes[*update.Key] = "", false
		} else {
		  s.Attributes[*update.Key] = *update.NewValue
        }
      }
      // Go to the next item
      contentIndex++
    } else if op.ReplaceAttributes != nil {
	  //
	  // Replace Attributes
	  //
      if insertDepth > 0 || deleteDepth > 0 {
        return os.NewError( "Cannot replace attributes inside an insertion sequence" )
	  }
      if c == nil {
		return os.NewError( "Document is shorter than op")
	  }
	  s, ok := c.(*ElementStart)
	  if !ok {
        return os.NewError( "Cannot replace attributes at this position, because in the document there is no element start" )
	  }

	  // Create a dictionary of attributes
      attribs := make(map[string]string)
      for _, v := range op.ReplaceAttributes.OldAttribute {
        attribs[*v.Key] = *v.Value
      }
      // Compare attributes from the doc and docop. They should be equal
      for k, v := range s.Attributes {      
		cmp, ok := attribs[k]
		if !ok {
		  return os.NewError("Attribute values do not match")
		}
        if v != cmp {
          return os.NewError( "Cannot replace attributes because attribute values differ" )
		}
      }

	  /* TODO
      // Compute the annotation for this element start
      var anno = doc.format[contentIndex];
      if ( anno != docAnnotation )
      {
        docAnnotation = anno;
        updatedAnnotation = this.computeAnnotation(docAnnotation, annotationUpdate, annotationUpdateCount);
      }
      doc.format[contentIndex] = updatedAnnotation;
      */
	  
      // Change the attributes of the element
      s.Attributes = make(map[string]string)
      for _, v := range op.ReplaceAttributes.NewAttribute {
        s.Attributes[*v.Key] = *v.Value
      }
      // Go to the next item
      contentIndex++
    } else if op.AnnotationBoundary != nil {
	  /* TODO
      // Change the 'annotationUpdate' and find out when the annotation update becomes empty.
      for( var a in op.annotation_boundary.end )
      {        
        var key = op.annotation_boundary.end[a];
        if ( !annotationUpdate[key] )
          throw "Cannot end annotation because the doc and op annotation do not match.";
        delete annotationUpdate[key];
        annotationUpdateCount--;
      }
      // Change the 'annotationUpdate' and find out when the annotation update becomes empty.
      for( var a in op.annotation_boundary.change )
      {
        var change = op.annotation_boundary.change[a];
        if ( !annotationUpdate[change.key] )
          annotationUpdateCount++;
        annotationUpdate[change.key] = change;
      }
      // The commented line below is WRONG because the update cannot be applied to the format on the left side of the cursor.
      // updatedAnnotation = this.computeAnnotation(docAnnotation, annotationUpdate, annotationUpdateCount);
      */
    }
  }
    
  if deleteDepth != 0 {
    return os.NewError( "Not all delete element starts have been matched with a delete element end" )
  }
  if insertDepth != 0 {
    return os.NewError( "Not all opened elements have been closed" )
  }
  if contentIndex < len(doc.Content) {
     return os.NewError( "op is too small for document" )
  }
  return nil
}
