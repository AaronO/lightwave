if ( !window.LW ) {
  LW = { };
}

LW.Doc = function(url) {
  this.url = url;
  this.version = 0;
  this.hash = "TODOHASH";
  this.content = { "_rev":0, "_data":{"_rev":0, "_id":LW.JsonOT.uniqueId_()}, "_meta":{"_rev":0, "_id":LW.JsonOT.uniqueId_()} };
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
	for( var i in obj ) {
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

// -------------------------------------------------------
// Inbox

LW.Inbox = {
  docs_ : {},
  self : null
};

LW.Inbox.uniqueId = function() {
  var str = Math.random().toString();
  return str.substr(2, str.length - 2);
};

LW.Inbox.init = function() {
  // Create the json document that contains the inbox
  LW.Inbox.self = LW.Inbox.getOrCreateDoc("/" + LW.Rpc.domain + "/_user/" + LW.Rpc.user + "/inbox");
};

// @param url has the format "/host-name/conversation-id"
LW.Inbox.getOrCreateDoc = function(url) {
  if ( LW.Inbox.docs_[url] ) {
	return LW.Inbox.docs_[url];
  }
  
  var d = new LW.Doc(url);
  LW.Inbox.docs_[url] = d;
  return d;
};

// The id has the form "/host-name/conversation-id!object-id" where conversation-id has the form (/{identifier})*
LW.Inbox.getElementById = function(id) {
  var i = id.indexOf("!");
  var convid;
  var objectid;
  if ( i == -1 ) {
	convid = id;
  } else {
	convid = id.substr(0, i);
	objectid = id.substring(i + 1, id.length);
  }
  var d = LW.Inbox.docs_[convid];
  if ( !objectid ) {
	return d;
  }
  return d.getElementById(objectid);
};

