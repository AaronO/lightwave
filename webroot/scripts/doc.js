if ( !window.LW ) {
  LW = { };
}

LW.Doc = function(url) {
  this.url = url;
  this.version = 0;
  this.hash = "TODOHASH";
  this.content = { };
  this.pendingSubmit = null;
  this.queue = [];
};

LW.Doc.prototype.recvDocMutation = function(mutation) {
  // Check that version and hash are as expected
  // ...
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
  this.version = mutation._endRev;
  this.hash = mutation._endHash;
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
  // TODO
};

// -------------------------------------------------------
// Inbox

LW.Inbox = {
  docs_ : {}
};

LW.Inbox.getOrCreateDoc = function(url) {
  if ( LW.Inbox.docs_[url] ) {
	return LW.Inbox.docs_[url];
  }
  
  var d = new LW.Doc(url);
  LW.Inbox.docs_[url] = d;
  return d;
};

// The id has the form "/localhost/conversation-id!object-id" where conversation-id has the form (/{identifier})*
LW.Inbox.getElementById = function(id) {
  var i = id.indexOf("!");
  var convid;
  var objectid;
  if ( i == -1 ) {
	convid = id;
  } else {
	url = id.substr(0, i);
	convid = id.substring(i + 1, len(id));
  }
  var d = new LW.Doc(convid);
  if ( !objectid ) {
	return d;
  }
  return d.getElementById(objectid);
};
