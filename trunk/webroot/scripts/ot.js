if ( !window.LW ) {
  LW = { };
}

if ( !LW.JsonOT ) {
  LW.JsonOT = { };
}

// Flags for mutation application
LW.JsonOT.CreateIDs = 1;

LW.JsonOT.uniqueIdCounter_ = 1;
LW.JsonOT.uniqueId_ = function() {
  return LW.JsonOT.uniqueIdCounter_++;
};

// Flags for mutation events
LW.JsonOT.AttributeModified = 1;
LW.JsonOT.ObjectModified = 2;
LW.JsonOT.DocumentModified = 3;
LW.JsonOT.ArrayModified = 4;
LW.JsonOT.ArrayElementModified = 5;
LW.JsonOT.ArrayElementDeleted = 6;
LW.JsonOT.ArrayElementInserted = 7;
LW.JsonOT.ArrayElementLifted = 8;
LW.JsonOT.ArrayElementSqueezed = 9;

LW.JsonOT.applyDocMutation = function( doc, docmutation, flags ) {
  if ( docmutation._data ) {
	doc._data = LW.JsonOT.applyMutation_(doc, doc._data, docmutation._data, flags );
	// Event
	if ( doc._data_cb ) {
	  doc._data_cb(doc, doc, "_data", docmutation._data, LW.JsonOT.AttributeModified );
	}
  }
  if ( docmutation._meta ) {
	doc._meta = LW.JsonOT.applyMutation_(doc, doc._meta, docmutation._meta, flags );
	// Event
	if ( doc._meta_cb ) {
	  doc._meta_cb(doc, doc, "_meta", docmutation._meta, LW.JsonOT.AttributeModified );
	}
  }
  // Event
  if ( doc._cb ) {
	doc._cb(doc, doc, nil, docmutation, LW.JsonOT.ObjectModified );
  }
};

LW.JsonOT.applyMutation_ = function( doc, val, mutation, flags ) {
  if ( mutation["$object"] == true ) {
	return LW.JsonOT.applyObjMutation_( doc, val, mutation, flags )
  } else if ( mutation["$array"] ) {
	return LW.JsonOT.applyArrayMutation_( doc, val, mutation, flags )
  } else if ( mutation["$text"] ) {
	return LW.JsonOT.applyTextMutation_( doc, val, mutation, flags )
  } else {
	return LW.JsonOT.applyInsertMutation_( doc, mutation, flags )
  }
};

LW.JsonOT.applyObjMutation_ = function( doc, obj, mutation, flags ) {
  for ( key in mutation ) {
	var m = mutation[key];
	if ( key[0] == '$' ) {
	  continue;
	}
	if ( flags & LW.JsonOT.CreateIDs == LW.JsonOT.CreateIDs ) {
	  if ( !obj._id ) {
		obj._id = LW.JsonOT.uniqueId_();
	  }
	  obj["_rev"] = doc._rev;
	}
	if ( m === null ) {
	  delete obj[key];
	} else {
	  obj[key] = LW.JsonOT.applyMutation_(doc, obj[key], m, flags);
	}
	// Event
	if ( obj["_cb_" + key] ) {
	  obj["_cb_" + key](doc, obj, key, m, LW.JsonOT.AttributeModified);
	}
  }
  // Event
  if ( obj._cb ) {
	obj._cb( doc, obj, nil, mutation, LW.JsonOT.ObjectModified );
  }
  return obj;
};

LW.JsonOT.applyArrayMutation_ = function( doc, arr, mutation, flags ) {
  var index = 0;

  // Find the lifts
  var lifts = {};
  for ( var i in mutation["$array"] ) {
	if ( i[0] == "_" ) continue;
	// Skip event handlers
	var mut = mutation["$array"][i];
	if ( mut["$delete"] != null ) {
	  index += mut["$delete"];
	} else if ( mut["$skip"] != null ) {
	  index += mut["$skip"];
	} else if ( mut["$object"] == true || mut["$array"] || mut["$text"] ) {
	  index++;
	} else if ( mut["$lift"] ) {
	  var val = arr[index];
	  if ( mut["$mutation"] ) {
		val = LW.JsonOT.applyMutation_(doc, val, mut["$mutation"], flags);
	  }
	  lifts[mut["$lift"]] = val;
	  index++;
	} else {
	  // Do nothing by intention: InsertMutation or SqueezeMutation
	}
  }

  index = 0;
  for ( var i in mutation["$array"] ) {
	// Skip event handlers
	if ( i[0] == "_" ) continue;
	var mut = mutation["$array"][i];
	if ( mut["$delete"] != null ) {
	  arr.splice(index, mut["$delete"]);
	  // Event
	  if ( arr._cb_deleted ) {
		arr._cb_deleted( doc, arr, index, mut, LW.JsonOT.ArrayElementDeleted );
	  }
	} else if ( mut["$skip"] != null ) {
	  index += mut["$skip"];
	} else if ( mut["$lift"] ) {
	  arr.splice(index, 1);
	  // Event
	  if ( arr._cb_lifted ) {
		arr._cb_lifted( doc, arr, index, mut, LW.JsonOT.ArrayElementLifted );
	  }
	} else if ( mut["$squeeze"] ) {
	  arr.splice(index, 0, lifts[mut["$squeeze"]]);
	  // Event
	  if ( arr._cb_squeezed ) {
		arr._cb_squeezed( doc, arr, index, mut, LW.JsonOT.ArrayElementSqueezed );
	  }	  
	  index++;
	} else if ( mut["$object"] == true || mut["$array"] || mut["$text"] ) {
	  arr[index] = LW.JsonOT.applyMutation_(doc, arr[index], mut, flags);
	  // Event
	  if ( arr._cb_deleted ) {
		arr._cb_deleted( doc, arr, index, mut, LW.JsonOT.ArrayElementModified );
	  }
	  index++;
	} else {
	  // Insert mutation
	  arr.splice(index, 0, LW.JsonOT.applyInsertMutation_(doc, mut, flags));
	  	  // Event
	  if ( arr._cb_inserted ) {
		arr._cb_inserted( doc, arr, index, mut, LW.JsonOT.ArrayElementInserted );
	  }
	  index++;
	}
  }
  
  // Event
  if ( arr._cb ) {
	arr._cb( doc, obj, nil, mutation, LW.JsonOT.ArrayModified );
  }

  return arr;
};

LW.JsonOT.applyTextMutation_ = function( doc, txt, mutation, flags ) {
  var index = 0;
  for ( var i in mutation["$text"] ) {
	mut = mutation["$text"][i];
	if ( mut["$delete"] != null ) {
	  txt = txt.substr(0, index) + txt.substring(index + mut["$delete"], txt.length);
	} else if ( mut["$skip"] != null ) {
	  index += mut["$skip"]
	} else {
	  // Must be an insert mutation, e.g. a string
	  txt = txt.substr(0, index) + mut + txt.substring(index, txt.length);
	  index += mut.length;
	}
  }
  return txt;
};

LW.JsonOT.applyInsertMutation_ = function( doc, mutation, flags ) {
  if ( Array.isArray(mutation) ) {
	return LW.JsonOT.applyInsertArrayMutation_( doc, mutation, flags );
  } else if ( mutation != null && typeof(mutation) == "object" ) {
	return LW.JsonOT.applyInsertObjectMutation_( doc, mutation, flags );
  } else {
	return mutation;
  }
};

LW.JsonOT.applyInsertObjectMutation_ = function( doc, mutation, flags ) {
  var m = {};
  for ( var key in mutation ) {
	m[key] = LW.JsonOT.applyInsertMutation_(doc, mutation[key], flags);
  }
  if (flags & LW.JsonOT.CreateIDs == LW.JsonOT.CreateIDs) {
	m._id = LW.JsonOT.uniqueId_();	
	m._rev = doc._rev;
  }
  return m;
};

LW.JsonOT.applyInsertArrayMutation_ = function( doc, mutation, flags ) {
  var a = [];
  for ( var i in mutation ) {
	a[i] = LW.JsonOT.applyInsertMutation_(doc, mutation[i], flags);
  }
  if (flags & LW.JsonOT.CreateIDs == LW.JsonOT.CreateIDs) {
	a._id = LW.JsonOT.uniqueId_();	
	a._rev = doc._rev;
  }
  return a;
};

Array.isArray = Array.isArray || function(o) { return Object.prototype.toString.call(o) === '[object Array]'; };