/*
  The main package is here to stay
*/
package lightwave

import (
  "os"
  "math"
  "fmt"
  "log"
  vec "container/vector"
)

// ------------------------------------------------------------------------------
// Helper

func IntMin(j, i int) int {
  if j < i {
	return j
  }
  return i
}

// ------------------------------------------------------------------------------
// Functions for type-checking JSON-encoded mutations

func IsInsertMutation( obj interface{} ) bool {
  switch t := obj.(type) {
	case map[string]interface{}:
	  if _, ok := obj.(map[string]interface{})["$object"]; ok {
		return false
	  }
	  if _, ok := obj.(map[string]interface{})["$array"]; ok {
		return false
	  }
	  if _, ok := obj.(map[string]interface{})["$text"]; ok {
		return false
	  }
	  if _, ok := obj.(map[string]interface{})["$skip"]; ok {
		return false
	  }
	  if _, ok := obj.(map[string]interface{})["$delete"]; ok {
		return false
	  }
	  if _, ok := obj.(map[string]interface{})["$lift"]; ok {
		return false
	  }
	  if _, ok := obj.(map[string]interface{})["$squeeze"]; ok {
		return false
	  }	
	  return true
	case []interface{}, float64, bool, string, int, int64:
	  return true
  }
  return false
}

func IsObjectMutation( obj interface{} ) bool {
  switch t := obj.(type) {
	case map[string]interface{}:
	  o, ok := obj.(map[string]interface{})["$object"]
	  if !ok {
		return false
	  }
	  b, ok := o.(bool)
	  if !ok || !b {
		return false
	  }
	  return true
  }
  return false
}

func isMutationArray( obj interface{}, kind string ) bool {
  switch t := obj.(type) {
	case map[string]interface{}:
	  o, err := obj.(map[string]interface{})[kind]
	  if !err {
		return false
	  }
	  _, err = o.([]interface{})
	  if !err {
		return false
	  }
	  return true
  }
  return false
}

func IsArrayMutation( obj interface{} ) bool {
  return isMutationArray(obj, "$array")
}

func IsTextMutation( obj interface{} ) bool {
  return isMutationArray(obj, "$text")
}

func isSkipOrDeleteMutation( obj interface{}, kind string ) bool {
  switch t := obj.(type) {
	case map[string]interface{}:
	  o, ok := obj.(map[string]interface{})[kind]
	  if !ok {
		return false
	  }
	  n, ok := o.(float64)
	  if !ok || n < 0 || math.Ceil(n) != n {
		return false
	  }
	  return true
  }
  return false
}

func IsSkipMutation( obj interface{} ) bool {
  return isSkipOrDeleteMutation( obj, "$skip" )
}

func IsDeleteMutation( obj interface{} ) bool {
  return isSkipOrDeleteMutation( obj, "$delete" )
}

func isSqueezeOrLiftMutation( obj interface{}, kind string ) bool {
  switch t := obj.(type) {
	case map[string]interface{}:
	  o, err := obj.(map[string]interface{})[kind]
	  if !err {
		return false
	  }
	  _, err = o.(string)
	  if !err {
		return false
	  }
	  return true
  }
  return false
}

func IsSqueezeMutation( obj interface{} ) bool {
  return isSqueezeOrLiftMutation( obj, "$squeeze" )
}

func IsLiftMutation( obj interface{} ) bool {
  return isSqueezeOrLiftMutation( obj, "$lift" )
}

func IsDocumentMutation( obj interface{} ) bool {  
  o, ok := obj.(map[string]interface{})
  if !ok {
	return false
  }
  _, ok1 := o["_rev"]
  _, ok2 := o["_data"]
  _, ok3 := o["_meta"]
  if !ok1 || (!ok2 && !ok3) {
	return false
  }
  return true
}

//----------------------------------------------------------------
// Mutation

type Mutation interface{}

type LiftMutation map[string]interface{}

func (self LiftMutation) Id() string {
  return self["$lift"].(string)
}

func (self LiftMutation) Mutation() interface{} {
  return self["$mutation"]
}

func (self LiftMutation) HasMutation() bool {
  _, err := self["$mutation"]
  return err
}

func toLiftMutation( m interface{} ) LiftMutation {
  return LiftMutation( m.(map[string]interface{}) )
}

type SqueezeMutation map[string]interface{}

func (self SqueezeMutation) Id() string {
  return self["$squeeze"].(string)
}

func toSqueezeMutation( m interface{} ) SqueezeMutation {
  return SqueezeMutation( m.(map[string]interface{}) )
}

type SkipMutation map[string]interface{}

func (self SkipMutation) Count() int {
  return int(self["$skip"].(float64))
}

func (self SkipMutation) SetCount(count int) {
  self["$skip"] = float64(count)
}

func toSkipMutation( m interface{} ) SkipMutation {
  return SkipMutation( m.(map[string]interface{}) )
}

func NewSkipMutation(count int) map[string]interface{} {
  s := make(map[string]interface{})
  s["$skip"] = float64(count)
  return s
}

type DeleteMutation map[string]interface{}

func toDeleteMutation( m interface{} ) DeleteMutation {
  return DeleteMutation( m.(map[string]interface{}) )
}

func (self DeleteMutation) Count() int {
  return int(self["$delete"].(float64))
}

func (self DeleteMutation) SetCount(count int) {
  self["$delete"] = float64(count)
}

func NewDeleteMutation(count int) map[string]interface{} {
  s := make(map[string]interface{})
  s["$delete"] = float64(count)
  return s
}

type ObjectMutation map[string]interface{}

func (self ObjectMutation) RemoveAttribute(attr string) {
  self[attr] = nil, false
}

func toObjectMutation( m interface{} ) ObjectMutation {
  return ObjectMutation( m.(map[string]interface{}) )
}

type TextMutation map[string]interface{}

func (self TextMutation) Array() []interface{} {
  return self["$text"].([]interface{})
}

func (self TextMutation) SetArray(arr []interface{} ) {
  self["$text"] = arr
}

func toTextMutation( m interface{} ) TextMutation {
  return TextMutation( m.(map[string]interface{}) )
}

type ArrayMutation map[string]interface{}

func (self ArrayMutation) Array() []interface{} {
  return self["$array"].([]interface{})
}

func (self ArrayMutation) SetArray(arr []interface{} ) {
  self["$array"] = arr
}

func toArrayMutation( m interface{} ) ArrayMutation {
  return ArrayMutation( m.(map[string]interface{}) )
}

type InsertMutation Mutation

//----------------------------------------------------------------
// Transformer

type Transformer struct {
  sLifts map[string]LiftMutation
  cLifts map[string]LiftMutation
  sLiftCounterpart map[string]Mutation
  cLiftCounterpart map[string]Mutation
}

func (self Transformer) Transform( s, c DocumentMutation ) os.Error {
  // Safety check
  if s.AppliedAtRevision() != c.AppliedAtRevision() {
	return os.NewError("Mutations must be applicable to the same document version to be transformed")
  }
  s_tmp, s_ok := s.DataMutation()
  c_tmp, c_ok := c.DataMutation()
  // Both have a data mutation?
  if s_ok && c_ok {
	if IsInsertMutation(s_tmp) {
	  // If the server has an insert mutation it will win
	  c["_data"] = nil, false
	} else if IsInsertMutation(c_tmp) {
	  // If the server has an insert mutation it will win
	  s["_data"] = nil, false
	} else {
	  if !IsObjectMutation(c_tmp) {
		return os.NewError("The client-side mutation is not an object or insert mutation")
	  }
	  log.Println("Translating ", s_tmp, " | ", c_tmp)
	  if err := self.transform(toObjectMutation(s_tmp), toObjectMutation(c_tmp)); err != nil {
		return err
	  }
	  log.Println("Translated ", s_tmp, " | ", c_tmp)
	}
  }

  s_tmp, s_ok = s.MetaMutation()
  c_tmp, c_ok = c.MetaMutation()
  // Both have a data mutation?
  if s_ok && c_ok {
	if IsInsertMutation(s_tmp) {
	  // If the server has an insert mutation it will win
	  c["_data"] = nil, false
	} else if IsInsertMutation(c_tmp) {
	  // If the server has an insert mutation it will win
	  s["_data"] = nil, false
	} else {
	  if !IsObjectMutation(c_tmp) {
		return os.NewError("The client-side mutation is not an object or insert mutation")
	  }
	  if err := self.transform(toObjectMutation(s_tmp), toObjectMutation(c_tmp)); err != nil {
		return err
	  }
	}
  }

  s["_rev"] = s["_rev"].(float64) + 1
  c["_rev"] = c["_rev"].(float64) + 1
  
  return nil
}

func (self Transformer) transform( s, c ObjectMutation ) os.Error {
  if err := self.transform_pass0_object( s, c ); err != nil {
	return err
  }
  if err := self.transform_pass1_object( s, c ); err != nil {
	return err
  }
  return nil
}

func (self Transformer) transform_pass0_object( sobj, cobj ObjectMutation ) os.Error {  
  for name, sval := range sobj {
	// Do not handle the $object attribute
	if name == "$object" {
	  continue;
	}

	cval, err := cobj[name]
	if !err {
	  // There is nothing to transform for this attribute
	  continue
	}
	
	if IsInsertMutation(sval) || IsInsertMutation(cval) {
	  continue
	} else if ( IsObjectMutation(sval) && IsObjectMutation(cval) ) {
	  if err := self.transform_pass0_object( toObjectMutation( sval ), toObjectMutation( cval ) ); err != nil {
		return err
	  }
	} else if ( IsArrayMutation(sval) && IsArrayMutation(cval) ) {
	  if err := self.transform_pass0_array( toArrayMutation( sval ), toArrayMutation( cval ) ); err != nil {
		return err
	  }
	} else if ( IsTextMutation(sval) && IsTextMutation(cval) ) {
	  continue
	} else {
	  return os.NewError("The two mutations of the object are not compatible")
	}
  }
  return nil
}

func (self Transformer) transform_pass0_array( sobj, cobj ArrayMutation ) os.Error {
  var (
    sindex int = 0
    cindex int = 0
    sinside int = 0
    cinside int = 0
	sarr []interface{} = sobj["$array"].([]interface{})
	carr []interface{} = cobj["$array"].([]interface{})
  )
  
  for sindex < len(sarr) || cindex < len(carr) {
	var smut, cmut interface{}
	if sindex < len(sarr) {
	  smut = sarr[sindex]
	}
	if cindex < len(carr) {
	  cmut = carr[cindex]
	}

	// Skip all server inserts
	for IsInsertMutation(smut) || IsSqueezeMutation(smut) {        
	  sindex++
	  if sindex < len(sarr) {
		smut = sarr[sindex]
	  } else {
		smut = nil
	  }
	}

	// Skip all client inserts
	for IsInsertMutation(cmut) || IsSqueezeMutation(cmut) {        
	  cindex++
	  if cindex < len(carr) {
		cmut = sarr[cindex]
	  } else {
		cmut = nil
	  }
	}

	// End of mutation reached?
	if sindex == len(sarr) || cindex == len(carr) {
	  break
	}
	
	if ( IsLiftMutation(smut) ) {
	  id := toLiftMutation(smut).Id()
	  self.sLifts[id] = toLiftMutation(smut);
	  self.cLiftCounterpart[id] = cmut;
	}
	if IsLiftMutation(cmut) {
	  id := toLiftMutation(cmut).Id()
	  self.cLifts[id] = toLiftMutation(cmut);
	  self.sLiftCounterpart[id] = smut;
	}
	
	if IsLiftMutation(smut) && IsLiftMutation(cmut) && toLiftMutation(smut).HasMutation() && toLiftMutation(cmut).HasMutation() {
	  self.transform_pass0_lift( toLiftMutation(smut).Mutation(), toLiftMutation(cmut).Mutation() )
	} else if IsLiftMutation(smut) && toLiftMutation(smut).HasMutation() && (IsArrayMutation(cmut) || IsTextMutation(cmut) || IsObjectMutation(cmut) ) {
	  self.transform_pass0_lift( toLiftMutation(smut).Mutation(), cmut )
	} else if IsLiftMutation(cmut) && toLiftMutation(cmut).HasMutation() && (IsArrayMutation(smut) || IsTextMutation(smut) || IsObjectMutation(smut) ) {
	  self.transform_pass0_lift( smut, toLiftMutation(cmut).Mutation() )
	}

	if ( IsDeleteMutation(smut) || IsSkipMutation(smut) ) && ( IsDeleteMutation(cmut) || IsSkipMutation(cmut)) {
	  var sdel, cdel int
	  if IsDeleteMutation(smut) {
		sdel = toDeleteMutation(smut).Count()
	  } else {
		sdel = toSkipMutation(smut).Count()
	  }
	  if IsDeleteMutation(cmut) {
		cdel = toDeleteMutation(cmut).Count()
	  } else {
		cdel = toSkipMutation(cmut).Count()
	  }
	  del := IntMin(sdel - sinside, cdel - cinside);
	  sinside += del;
	  cinside += del;
	  if sdel == del {
		sinside = 0;
		sindex++;
	  }
	  if cdel == del {
		cinside = 0;
		cindex++;
	  }
	} else if IsSkipMutation(smut) { // ... and mutation at the client
	  // Q_ASSERT(cmut.isArrayMutation() || cmut.isObjectMutation() || cmut.isRichTextMutation() || cmut.isTextMutation() || cmut.isLiftMutation() );
	  cindex++;
	  sinside++;
	  if toSkipMutation(smut).Count() == sinside {
		sinside = 0;
		sindex++;
	  }
	} else if IsSkipMutation(cmut) { // ... and mutation at the srver        
	  // Q_ASSERT(smut.isArrayMutation() || smut.isObjectMutation() || smut.isRichTextMutation() || smut.isTextMutation() || smut.isLiftMutation() );
	  sindex++;
	  cinside++;
	  if toSkipMutation(cmut).Count() == cinside {
		cinside = 0;
		cindex++;
	  }
	} else if IsDeleteMutation(smut) { // ... and mutation at the client
	  // Q_ASSERT(cmut.isArrayMutation() || cmut.isObjectMutation() || cmut.isRichTextMutation() || cmut.isTextMutation() || cmut.isLiftMutation() );
	  cindex++;
	  sinside++;
	  if toDeleteMutation(smut).Count() == sinside {            
		sinside = 0;
		sindex++;
	  }
	} else if IsDeleteMutation(cmut) { // ... and mutation at the server
	  // Q_ASSERT(smut.isArrayMutation() || smut.isObjectMutation() || smut.isRichTextMutation() || smut.isTextMutation() || smut.isLiftMutation() );
	  sindex++;
	  cinside++;
	  if toDeleteMutation(cmut).Count() == cinside {
		cinside = 0;
		cindex++;
	  }
	} else if IsLiftMutation(smut) {
	  // Q_ASSERT(cmut.isArrayMutation() || cmut.isObjectMutation() || cmut.isRichTextMutation() || cmut.isTextMutation() || cmut.isLiftMutation() );
	  sindex++;
	  cindex++;
	} else if IsLiftMutation(cmut) {
	  // Q_ASSERT(smut.isArrayMutation() || smut.isObjectMutation() || smut.isRichTextMutation() || smut.isTextMutation() || smut.isLiftMutation() );
	  sindex++;
	  cindex++;
	} else if IsArrayMutation(smut) && IsArrayMutation(cmut) {
	  self.transform_pass0_array( toArrayMutation(smut), toArrayMutation(cmut) )
	  cindex++;
	  sindex++;
	} else if IsObjectMutation(smut) && IsObjectMutation(cmut) {
	  self.transform_pass0_object( toObjectMutation(smut), toObjectMutation(cmut) )
	  cindex++;
	  sindex++;
	} else if IsTextMutation(smut) && IsTextMutation(cmut) {
	  cindex++;
	  sindex++;
	} else {
	  return os.NewError("The two mutations in the array do not match")
	}
  }
  
  return nil
}

func (self Transformer) transform_pass0_lift( s, c Mutation ) os.Error {
  if IsObjectMutation(s) && IsObjectMutation(c) {
	self.transform_pass0_object(toObjectMutation(s), toObjectMutation(c))
	self.transform_pass1_object(toObjectMutation(s), toObjectMutation(c))
  } else if IsArrayMutation(s) && IsArrayMutation(c) {
	self.transform_pass0_array(toArrayMutation(s), toArrayMutation(c))
	self.transform_pass1_array(toArrayMutation(s), toArrayMutation(c))
  } else if IsTextMutation(s) && IsTextMutation(c) {
	self.transform_pass1_text(toTextMutation(s), toTextMutation(c) )
  } else {
	return os.NewError("The two mutations are either incompatible or they are not allowed inside a lift")
  }
  return nil
}

func (self Transformer) transform_pass1_object( sobj, cobj ObjectMutation ) os.Error {
  for name, smut := range sobj {	
	if name == "$object" {
	  continue;
	}
	cmut, err := cobj[name]
	if !err {
	  continue
	}
	
	if IsInsertMutation(smut) {
	  cobj.RemoveAttribute(name);
	} else if IsObjectMutation(smut) {
	  if IsObjectMutation(cmut) {
		if err := self.transform_pass1_object(toObjectMutation(smut), toObjectMutation(cmut) ); err != nil {
		  return err
		}
	  } else if IsInsertMutation(cmut) {
		sobj.RemoveAttribute(name)
	  } else {
		return os.NewError("The two mutations are not compatible and/or not allowed inside an object mutation. This should have been detected in pass 0")
	  }
	} else if IsArrayMutation(smut) {
	  if IsArrayMutation(cmut) {
		if err := self.transform_pass1_array(toArrayMutation(smut), toArrayMutation(cmut) ); err != nil {
		  return err
		}
	  } else if IsInsertMutation(cmut) {
		sobj.RemoveAttribute(name)
	  } else {
		return os.NewError("The two mutations are not compatible and/or not allowed inside an object mutation. This should have been detected in pass 0")
	  }
	} else if IsTextMutation(smut) {
	  if IsTextMutation(cmut) {
		if err := self.transform_pass1_text(toTextMutation(smut), toTextMutation(cmut) ); err != nil {
		  return err
		}
	  } else if IsInsertMutation(cmut) {
		sobj.RemoveAttribute(name);
	  } else {
		return os.NewError("The two mutations are not compatible and/or not allowed inside an object mutation. This should have been detected in pass 0")
	  }
	} else {
	  return os.NewError("This mutation is not allowed inside an object mutation. This should have been detected in pass 0")
	}
  }
  
  return nil
}

func (self Transformer) transform_pass1_array( sobj, cobj ArrayMutation ) os.Error {
  var (
	sarr vec.Vector = sobj.Array()
	carr vec.Vector = cobj.Array()
    sindex int = 0
    cindex int = 0
    sinside int = 0
    cinside int = 0
  )
  
  // Loop until end of one mutation is reached
  for sindex < len(sarr) || cindex < len(carr) {
	var smut, cmut interface{}
	if sindex < len(sarr) {
	  smut = sarr[sindex]
	}
	if cindex < len(carr) {
	  cmut = carr[cindex]
	}

	// Server insert/squeeze go first
	for IsInsertMutation(smut) || IsSqueezeMutation(smut) {
	  if cinside > 0 {
		self.split(carr, cindex, cinside )
		cinside = 0
		cindex++
		cmut = carr[cindex]
	  }

	  if IsInsertMutation(smut) {
		// TODO: check the Insert mutation individually for correctness
		carr.Insert(cindex, NewSkipMutation(1))
		cindex++
		sindex++
	  } else if IsSqueezeMutation(smut) {
		sSqueeze := toSqueezeMutation(smut)
		// Which operation does the client side have for this object?
        c, hasCounterpart := self.cLiftCounterpart[sSqueeze.Id()]
        if !hasCounterpart || IsSkipMutation(c) {
		  // Server lift remains, client skips it at the new position
		  carr.Insert(cindex, NewSkipMutation(1))
		  cindex++
		  sindex++
		} else if IsDeleteMutation(c) {
		  // Client deletes the object at its new position
		  carr.Insert(cindex, NewDeleteMutation(1))
		  cindex++
		  // Server removes squeeze because it is already deleted by the client
		  sarr.Delete(sindex)
		} else if IsLiftMutation(c) {
		  // sLift := self.sLifts[sSqueeze.Id()]
		  cLift := toLiftMutation(c)
		  if cLift.HasMutation() {
			// The client does not lift it and therefore it mutates it at the new position
			carr.Insert(cindex, cLift.Mutation())
			cindex++
			// Server keeps its squeeze at this position
			sindex++
		  } else {
			// Client skips the squeezed object and does not lift it
			carr.Insert(cindex, NewSkipMutation(1))
			cindex++
			// Server keeps its squeeze at this position
			sindex++
		  }
		} else {
		  // The client somehow mutates the object. It must do so at its new position
		  carr.Insert(cindex, c)
		  cindex++
		  sindex++
		}
	  }
            
	  if sindex < len(sarr) {
		smut = sarr[sindex];
	  } else {
		smut = nil
	  }
	}
	
	// Client insert/squeeze go next
	for IsInsertMutation(cmut) || IsSqueezeMutation(cmut) {
	  if sinside > 0 {
		self.split(sarr, sindex, sinside )
		sinside = 0
		sindex++
		smut = sarr[sindex]
	  }

	  if IsInsertMutation(cmut) {
		// TODO: check the Insert mutation individually for correctness
		cindex++
		sarr.Insert(sindex, NewSkipMutation(1));
		sindex++
	  } else {
		cSqueeze := toSqueezeMutation(cmut)
		// Which operation does the server side have for this object?
		s, hasCounterpart := self.sLiftCounterpart[cSqueeze.Id()]
		if !hasCounterpart || IsSkipMutation(s) {
		  sarr.Insert(sindex, NewSkipMutation(1))
		  sindex++
		  cindex++
		} else if IsDeleteMutation(s) {
		  // Server deletes the object at its the new position
		  sarr.Insert(sindex, NewDeleteMutation(1))
		  sindex++
		  // Client removes squeeze because it is already deleted by the client
		  carr.Delete(cindex)
		} else if IsLiftMutation(s) {
		  // The server must lift the object here instead
		  sarr.Insert(sindex, s)
		  sindex++
		  // The server lifted this object as well -> the client cannot lift it ->
		  // then the client cannot squeeze it in here
		  carr.Delete(cindex)
		} else {
		  // The server mutates the object at its new position
		  sarr.Insert(sindex, s)
		  sindex++
		  // The client squeezes the object in here
		  cindex++
		}
	  }
	  
	  if cindex < len(carr) {
		cmut = carr[sindex];
	  } else {
		cmut = nil
	  }
	}

	if sindex == len(sarr) && cindex == len(carr) {
	  break
	}	
	if sindex == len(sarr) || cindex == len(carr) {
	  return os.NewError("The mutations do not have equal length")
	}

	//
	// Lift, Skip, Delete, mutations
	//

	if IsLiftMutation(smut) {
	  // Both are lifting the same object?
	  if IsLiftMutation(cmut) {
		// The client removes its lift. If it has a mutation then it will be moved to the corresponding squeeze
		carr.Delete(cindex)
		// The server removes its lift. It will be placed where the client moved it.
		sarr.Delete(sindex)
	  } else if IsDeleteMutation(cmut) {
		// The server does not lift the object because it is deleted
		sarr.Delete(sindex)
		// The client removes its delete. The delete is put where the server squeezed it.
		cindex, cinside = self.shorten(carr, cindex, cinside, 1)		
	  } else if IsSkipMutation(cmut) {
		// The client does not skip this element here. It is skipped where it is squeezed in
		cindex, cinside = self.shorten(carr, cindex, cinside, 1)
		// Server remains with its lift
		sindex++
	  } else {
		// Q_ASSERT(cmut.isArrayMutation() || cmut.isObjectMutation() || cmut.isTextMutation() || cmut.isRichTextMutation() );
		// The client removes its mutation here. It is shifted to the new position where the object is squeezed in.
		carr.Delete(cindex);
		// Server remains with its lift
		sindex++
	  }
	} else if IsLiftMutation(cmut) {
	  if IsDeleteMutation(smut) {
		// The client does not lift the object because it is deleted
		carr.Delete(cindex)
		// The server removes its delete. The delete is put where the client squeezed it.
		sindex, sinside = self.shorten(sarr, sindex, sinside, 1)
	  } else if IsSkipMutation(smut) {
		// The server does not skip this element here. It is skipped where it is squeezed in
		sindex, sinside = self.shorten(sarr, sindex, sinside, 1)
		// Client remains with its lift
		cindex++;
	  } else {
		// Q_ASSERT(smut.isArrayMutation() || smut.isObjectMutation() || smut.isTextMutation() || smut.isRichTextMutation() );
		// The server removes its mutation here. It is shifted to the new position where the object is squeezed in.
		sarr.Delete(sindex)
		// Client remains with its lift
		cindex++
	  }
	} else if IsDeleteMutation(smut) && IsDeleteMutation(cmut) {
	  sdel := toDeleteMutation(smut).Count()
	  cdel := toDeleteMutation(cmut).Count()
	  del := IntMin(sdel - sinside, cdel - cinside)
	  sindex, sinside = self.shorten(sarr, sindex, sinside, del)
	  cindex, cinside = self.shorten(carr, cindex, cinside, del)
	} else if IsSkipMutation(smut) && IsSkipMutation(cmut) {
	  sskip := toSkipMutation(smut).Count()
	  cskip := toSkipMutation(cmut).Count()
	  skip := IntMin(sskip - sinside, cskip - cinside)
	  sinside += skip;
	  cinside += skip;
	  if sinside == sskip {
		sinside = 0
		sindex++
	  }
	  if cinside == cskip {
		cinside = 0;
		cindex++
	  }
	} else if IsDeleteMutation(smut) && IsSkipMutation(cmut) {
	  sdel := toDeleteMutation(smut).Count()
	  cskip := toSkipMutation(cmut).Count()
	  count := IntMin(sdel - sinside, cskip - cinside)
	  sinside += count
	  if sinside == sdel {
		sinside = 0
		sindex++
	  }
	  cindex, cinside = self.shorten(carr, cindex, cinside, count)
	} else if IsSkipMutation(smut) && IsDeleteMutation(cmut) {
	  sskip := toSkipMutation(smut).Count()
	  cdel := toDeleteMutation(cmut).Count()
	  count := IntMin(sskip - sinside, cdel - cinside)
	  sindex, sinside = self.shorten(sarr, sindex, sinside, count)
	  cinside += count
	  if cinside == cdel {
		cinside = 0;
		cindex++
	  }
	} else if IsSkipMutation(smut) { // ... and mutation at the client
	  cindex++
	  sinside++
	  if toSkipMutation(smut).Count() == sinside {
		sinside = 0
		sindex++
	  }
	} else if IsSkipMutation(cmut) { // ... and mutation at the srver
	  sindex++
	  cinside++
	  if toSkipMutation(cmut).Count() == cinside {
		cinside = 0
		cindex++
	  }
	} else if IsDeleteMutation(smut) { // ... and mutation at the client
	  carr.Delete(cindex)
	  sinside++
	  if toDeleteMutation(smut).Count() == sinside {
		sinside = 0
		sindex++
	  }
	} else if IsDeleteMutation(cmut) { // ... and mutation at the server        
	  sarr.Delete(sindex)
	  cinside++
	  if toDeleteMutation(cmut).Count() == cinside {
		cinside = 0
		cindex++
	  }
	} else {
	  if IsObjectMutation(smut) && IsObjectMutation(cmut) {
		if err := self.transform_pass1_object(toObjectMutation(smut), toObjectMutation(cmut) ); err != nil {
		  return err
		}
	  } else if IsArrayMutation(smut) && IsArrayMutation(cmut) {
		if err := self.transform_pass1_array(toArrayMutation(smut), toArrayMutation(cmut) ); err != nil {
		  return err
		}
	  } else if IsTextMutation(smut) && IsTextMutation(cmut) {
		if err := self.transform_pass1_text(toTextMutation(smut), toTextMutation(cmut) ); err != nil {
		  return err
		}
	  } else {
		  return os.NewError("The mutations are not compatible or not allowed inside an array mutation")
	  }
	  sindex++
	  cindex++
	}
  }

  sobj.SetArray(sarr)
  cobj.SetArray(carr)
  
  return nil
}

func (self Transformer) transform_pass1_text( sobj, cobj TextMutation ) os.Error {
  var (
	sarr vec.Vector = sobj.Array()
	carr vec.Vector = cobj.Array()
    sindex int = 0
    cindex int = 0
    sinside int = 0
    cinside int = 0
  )
  
  // Loop until end of one mutation is reached
  for sindex < len(sarr) || cindex < len(carr) {
	var smut, cmut interface{}
	if sindex < len(sarr) {
	  smut = sarr[sindex]
	}
	if cindex < len(carr) {
	  cmut = carr[cindex]
	}
  
	// Server insertions go first
	if IsInsertMutation(smut) {
	  // In the middle of a client skip/delete? Split it
	  if cinside > 0 {
		self.split(carr, cindex, cinside )
		cinside = 0
		cindex++
		cmut = carr[cindex]
	  }
	  str, err := smut.(string)
	  if !err {
		panic("Only strings allowed inside a text mutation")
	  }
	  sindex++;
	  if len(str) > 0 {
		carr.Insert(cindex, NewSkipMutation(len(str)))
		cindex++
	  }
	  continue;
	}
	// Client insertions go next
	if IsInsertMutation(cmut)  {
	  // In the middle of a server skip/delete? Split it
	  if sinside > 0 {
		self.split(sarr, sindex, sinside )
		sinside = 0
		sindex++
		smut = sarr[sindex]
	  }

	  str, err := cmut.(string)
	  if !err {
		return os.NewError("Only strings allowed inside a text mutation")
	  }
	  cindex++
	  if len(str) > 0 {
		sarr.Insert(sindex, NewSkipMutation(len(str)))
		sindex++
	  }
	  continue;
	}

	if sindex == len(sarr) && cindex == len(carr) {
	  break
	}	
	if sindex == len(sarr) || cindex == len(carr) {
	  return os.NewError("The mutations do not have equal length")
	}

	if IsDeleteMutation(smut) && IsDeleteMutation(cmut) {
	  sdel := toDeleteMutation(smut).Count()
	  cdel := toDeleteMutation(cmut).Count()
	  del := IntMin(sdel - sinside, cdel - cinside)
	  sindex, sinside = self.shorten(sarr, sindex, sinside, del)
	  cindex, cinside = self.shorten(carr, cindex, cinside, del)
	} else if IsSkipMutation(smut) && IsSkipMutation(cmut) {
	  sskip := toSkipMutation(smut).Count()
	  cskip := toSkipMutation(cmut).Count()
	  skip := IntMin(sskip - sinside, cskip - cinside)
	  sinside += skip;
	  cinside += skip;
	  if sinside == sskip {
		sinside = 0
		sindex++
	  }
	  if cinside == cskip {
		cinside = 0;
		cindex++
	  }
	} else if IsDeleteMutation(smut) && IsSkipMutation(cmut) {
	  sdel := toDeleteMutation(smut).Count()
	  cskip := toSkipMutation(cmut).Count()
	  count := IntMin(sdel - sinside, cskip - cinside)
	  sinside += count
	  if sinside == sdel {
		sinside = 0
		sindex++
	  }
	  cindex, cinside = self.shorten(carr, cindex, cinside, count)
	} else if IsSkipMutation(smut) && IsDeleteMutation(cmut) {
	  sskip := toSkipMutation(smut).Count()
	  cdel := toDeleteMutation(cmut).Count()
	  count := IntMin(sskip - sinside, cdel - cinside)
	  sindex, sinside = self.shorten(sarr, sindex, sinside, count)
	  cinside += count
	  if cinside == cdel {
		cinside = 0;
		cindex++
	  }
	} else {
	  switch t := cmut.(type) {
		case map[string]interface{}:
		  o, err := cmut.(map[string]interface{})["$skip"]
		  if !err {
			panic("1")
		  }
		  _, err = o.(float64)
		  if !err {
			panic("2")
		  }
		default:
		  panic("3")
	  }
	  return os.NewError(fmt.Sprintf("Mutation not allowed in a text mutation:\n%v %v\n%v %v", smut, IsSkipMutation(smut), cmut, IsSkipMutation(cmut)))
	}
  }

  sobj.SetArray(sarr)
  cobj.SetArray(carr)
  
  return nil
}

func (self Transformer) split( arr vec.Vector, index, inside int ) {
  mut := arr[index]
  if IsDeleteMutation(mut) {
	arr.Insert(index+1, NewDeleteMutation( toDeleteMutation(mut).Count() - inside))
	toDeleteMutation(mut).SetCount(inside);
  } else if IsSkipMutation(mut) {
	arr.Insert(index+1, NewSkipMutation( toSkipMutation(mut).Count() - inside))
	toSkipMutation(mut).SetCount(inside);
  } else {
	panic("Unsupported mutation for split()")
  }
}

func (self Transformer) shorten( arr vec.Vector, index, inside, count int ) (rindex int, rinside int) {
  mut := arr[index]
  if IsDeleteMutation(mut) {
	del := toDeleteMutation(mut)
	del.SetCount( del.Count() - count )
	if inside == del.Count() {
	  inside = 0;
	  if del.Count() == 0 {
		arr.Delete(index)
	  } else {
		index++;
	  }
	}
  } else if IsSkipMutation(mut) {
	skip := toSkipMutation(mut)
	skip.SetCount( skip.Count() - 1 );
	if inside == skip.Count() {
	  inside = 0;
	  if skip.Count() == 0 {
		arr.Delete(index)
	  } else {
		index++
	  }
	}  
  } else {
	panic("Unsupported mutation for shorten()")
  }

  return index, inside
}

/*
func main() {

  text1 := "{\"$object\":true, \"abc\":\"xyz\", \"num\": 123, \"foo\":{\"$object\":true, \"a\":1,\"b\":3}, \"arr\":{\"$array\":[1,2,3]}}"
  dest1 := make(map[string]interface{})
  err := json.Unmarshal([]byte(text1), &dest1)
  if ( err != nil ) {
	log.Exitf("Cannot parse: %v", err);
  }

  text2 := "{\"$object\":true, \"torben\":\"weis\", \"num\": 321, \"foo\":{\"$object\":true, \"a\":6,\"c\":33}, \"arr\":{\"$array\":[4,5,6]}}"
  dest2 := make(map[string]interface{})
  err = json.Unmarshal([]byte(text2), &dest2)
  if ( err != nil ) {
	log.Exitf("Cannot parse: %v", err);
  }

  var t Transformer
  t.Transform( toObjectMutation(dest1), toObjectMutation(dest2) )
  
  enc1, _ := json.Marshal(dest1);
  fmt.Printf("t1=%v\n", string(enc1))
  enc2, _ := json.Marshal(dest2);
  fmt.Printf("t2=%v\n", string(enc2))
}
*/