if ( !window.LW ) {
  LW = { };
}

LW.Doc = function(url) {
    this.url = url;
    this.version = 0;
    this.hash = "TODOHASH";
    this.content = { "_rev":0, "_data":{"_rev":0, "_id":LW.JsonOT.uniqueId_()}, "_meta":{"_rev":0, "_id":LW.JsonOT.uniqueId_()} };
    this.content._data._parent = this.content;
    this.content._meta._parent = this.content;
    this.pendingSubmit = null;
    this.queue = [];
};

// Called when a new mutation arrived for this document
LW.Doc.prototype.recvDocMutation = function(mutation) {
  // Check that version and hash are as expected
  // ...
  // Get the new revision and hash
  this.version = mutation._endRev;
  this.hash = mutation._endHash;
  // Is this our own mutation?
  if ( this.pendingSubmit && mutation._token == this.pendingSubmit ) {
	console.log("Filtered out my own delta");
	this.queue.splice(0,1);
  } else {
	// Transform the mutation against our queued mutations
	for( var i in this.queue ) {
		var own = this.queue[i];
		// ...
	}
	LW.JsonOT.applyDocMutation( this.content, mutation, LW.JsonOT.CreateIDs )
  }
  console.log("New version of " + this.url + " is " + this.version);
  
  // Is this our own mutation?
  if ( this.pendingSubmit && mutation._token == this.pendingSubmit ) {
	delete mutation._token;
	this.pendingSubmit = null;
	// Send the next mutation from the queue if there is one.
	if ( this.queue.length > 0 ) {
	  this.sendDocMutation_(this.queue[0]);
	}
  }
};

LW.Doc.prototype.submitDocMutation = function(mutation) {
  LW.JsonOT.applyDocMutation( this.content, mutation, LW.JsonOT.CreateIDs )
  // Enqueue such that incoming messages can be transformed against it
  this.queue.push(mutation);
  if ( this.pendingSubmit ) {
	// Do nothing by intention
  } else {
	this.sendDocMutation_(mutation);
  }
};

LW.Doc.prototype.sendDocMutation_ = function(mutation) {
  this.pendingSubmit = Math.random().toString();
  mutation._token = this.pendingSubmit;
  mutation._rev = this.version;
  mutation._hash = this.hash;
  LW.Rpc.enqueue("/client" + this.url, JSON.stringify(mutation));
};

LW.Doc.prototype.getElementById = function(id) {
  return this.getElementById_(this.content, id);
};

LW.Doc.prototype.getElementById_ = function(obj, id) {
    if ( obj == null ) {
	return null;
    }
    if ( Array.isArray(obj) ) {
	if ( obj._id == id ) {
	    return obj;
	}
	for( var i = 0; i < obj.length; ++i ) {
	    var element = this.getElementById_(obj[i], id);
	    if ( element ) {
		return element;
	    }	  
	}
    } else if ( typeof(obj) == "object" ) {
	if ( obj._id == id ) {
	    return obj;
	}
        if ( obj["$rtf"] ) {
            return null;
        }
	for( var i in obj ) {
            if ( i == "_parent" ) {
                continue;
            }
	    var element = this.getElementById_(obj[i], id);
	    if ( element ) {
		return element;
	    }	  
	}
    }
    return null;
};

// Creats a mutation for this document that mutates the object specified by the id,
// but that mutates nothing else. This is just a convenience function.
// 
// @param id is a string denoting an object-id, i.e. a JSON object or array inside this document
LW.Doc.prototype.createMutationForId = function(id, mutation) {
  return this.createMutationForId_(this.content, id, mutation);
};

LW.Doc.prototype.createMutationForId_ = function(obj, id, mutation) {
    if ( obj == null ) {
	return null;
    }
    if ( Array.isArray(obj) ) {
	if ( obj._id == id ) {
	    return mutation;
	}
	for( var i = 0; i < obj.length; ++i ) {
	    var mut = this.createMutationForId_(obj[i], id, mutation);
	    if ( mut ) {
		arr = [ mut ];
		if ( i > 0 ) {
		    arr.splice(0, 0, {"$skip":i});
		}
		if ( i + 1 < obj.length ) {
		    arr.push({"$skip":(obj.length - i - 1)});
		}
		var r = {"$array":arr};
		return r;
	    }	  
	}
    } else if ( typeof(obj) == "object" ) {
	if ( obj._id == id ) {
	    return mutation;
	}
	for( var i in obj ) {
            if ( i == "_parent" ) {
                continue;
            }
            if ( obj["$rtf"] ) {
                return null;
            }
	    var mut = this.createMutationForId_(obj[i], id, mutation);
	    if ( mut ) {
		var r = {"$object":true};
		r[i] = mut;
		return r;
	    }
	}
    }
    return null;
};

