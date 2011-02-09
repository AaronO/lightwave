if ( !window.LW ) {
  LW = { };
}

// ----------------------------------------------------------
// Implementation of HTTP Post and Get using XMLHTTPRequest

LW.Rpc = {
    // TODO
    user : "weis",
    domain : "localhost",
    displayName : "Torben",
    logout : false,
    queue : [ ]
};

LW.Rpc.post = function(url, jsonData, callback, errCallback) {
  LW.Rpc.postOrGet_(url, "POST", jsonData, callback, errCallback);
}

LW.Rpc.get = function(url, callback, errCallback) {
  LW.Rpc.postOrGet_(url, "GET", null, callback, errCallback);
}

LW.Rpc.postOrGet_ = function(url, httpMethod, jsonData, callback, errCallback) {
  var xmlHttp = null;
  try {
	// Mozilla, Opera, Safari sowie Internet Explorer (ab v7)
	xmlHttp = new XMLHttpRequest();
  } 
  catch(e) {
	try {
	  // MS Internet Explorer (ab v6)
	  xmlHttp  = new ActiveXObject("Microsoft.XMLHTTP");
	} catch(e) {
	  try {
		// MS Internet Explorer (ab v5)
		xmlHttp  = new ActiveXObject("Msxml2.XMLHTTP");
	  } catch(e) {
		xmlHttp  = null;
	  }
	}
  }
  if (xmlHttp) {
	xmlHttp.open(httpMethod, url, true);
	xmlHttp.setRequestHeader("Content-type", "application/json")
	xmlHttp.onreadystatechange = function () {
	  if (xmlHttp.readyState == 4) {
		if( callback && xmlHttp.status == 200 ) {
		  callback(xmlHttp.responseText);
		} else if ( errCallback && xmlHttp.status != 200 ) {
		  errCallback();
		}
	  }
	};
	if ( jsonData ) {
	  console.log("Sending ...");
	  console.log(jsonData);
	  xmlHttp.send(jsonData);
	} else {
	  xmlHttp.send();
	}
	return true;
  }
  return false;
};

LW.Rpc.enqueueGet = function(url, callback, errCallback) {
    LW.Rpc.queue.push( { method:"get", url:url, callback:callback, errCallback:errCallback } );
  if ( LW.Rpc.queue.length == 1 ) {
	LW.Rpc.get(url, LW.Rpc.enqueueCallback_, LW.Rpc.enqueueErrCallback_);
  }
};

LW.Rpc.enqueue = function(url, jsonData, callback, errCallback) {
    LW.Rpc.queue.push( { method:"post", url:url, data:jsonData, callback:callback, errCallback:errCallback } );
  if ( LW.Rpc.queue.length == 1 ) {
	LW.Rpc.post(url, jsonData, LW.Rpc.enqueueCallback_, LW.Rpc.enqueueErrCallback_);
  }
};

LW.Rpc.enqueueCallback_ = function(reply) {
    console.log("Answer ...");
    console.log(reply);
    var msg = LW.Rpc.queue[0];
    if ( msg.callback ) {
	msg.callback(reply);
    }
    LW.Rpc.queue.splice(0,1);
    // Send more?
    if ( LW.Rpc.queue.length > 0 ) {
	var msg = LW.Rpc.queue[0];
        if ( msg.method == "get" ) {
	    LW.Rpc.get(msg.url, LW.Rpc.enqueueCallback_, LW.Rpc.enqueueErrCallback_);
        } else {
	    LW.Rpc.post(msg.url, msg.data, LW.Rpc.enqueueCallback_, LW.Rpc.enqueueErrCallback_);
        }
    }
};

LW.Rpc.enqueueErrCallback_ = function(reply) {
    if ( LW.Rpc.logout ) {
        return;
    }
    var msg = LW.Rpc.queue[0];
    if ( msg.errCallback ) {
	msg.errCallback(reply);
    }
    alert("Offline", "You seem to be offline");
};

// -------------------------------------------------------------
// Session

LW.Session = {
    id : null,
    /**
     * The keys are document URIs. The values are useless currently.
     */
    openDocs : { },
    openViews : { }
};

LW.Session.init = function() {
    // Open the inbox
    // LW.Session.open("/" + LW.Rpc.domain + "/_user/" + LW.Rpc.user + "/inbox")
// TODO: proper encoding
    LW.Session.openView('inbox', 'map=digest&with=with:' + LW.Rpc.user  + "@" + LW.Rpc.domain)
    // Start polling the session to get updates on the documents
    LW.Session.sessionPoll_();
};
	
// HTTP long call to get updates on the documents
LW.Session.sessionPoll_ = function() {
  LW.Rpc.get("/_sessionpoll", LW.Session.sessionPollCallback_, LW.Session.sessionErrCallback_);  
};

// Successful response from the HTTP long call
LW.Session.sessionPollCallback_ = function(reply) {
  console.log("Recv session ...")
  var json = JSON.parse(reply);
  console.log(json);
  for( var url in json ) {
	var doc = LW.Inbox.getOrCreateDoc(url);
	var mutations = json[url];
	for( var i in mutations ) {
	  doc.recvDocMutation(mutations[i]);
	}
  }
  LW.Session.sessionPoll_();
};

// Failed response from the HTTP long call
LW.Session.sessionErrCallback_ = function() {
    if ( LW.Rpc.logout ) {
        return;
    }
    alert("Offline", "You seem to be offline");
};
	
// Adds a new document to the session
LW.Session.open = function(uri, snapshot) {
    // This document is already open?
    if ( LW.Session.openDocs[uri] ) {
        return
    }
    // TODO: proper escaping of the URI
    LW.Rpc.get("/_open?uri=" + uri + "&snapshot=" + (snapshot ? "y" : "n"), function(reply) { LW.Session.openCallback_(reply, uri); }, LW.Session.sessionErrCallback_);
};

// Response to the request of adding a document to the session
LW.Session.openCallback_ = function(reply, uri) {
    // console.log("open: " + reply)
    var json = JSON.parse(reply);
    if ( json.ok == false ) {
        alert("Failed to open a document");
        return;
    }
    LW.Session.openDocs[uri] = true;
};

// Adds a new document to the session
LW.Session.close = function(uri) {
    // This document is already open?
    if ( !LW.Session.openDocs[uri] ) {
        return
    }
    // TODO: proper escaping of the URI
    LW.Rpc.get("/_close?uri=" + uri, function(reply) { LW.Session.closeCallback_(reply, uri); }, LW.Session.sessionErrCallback_);
};

// Response to the request of adding a document to the session
LW.Session.closeCallback_ = function(reply, uri) {
    // console.log("close: " + reply)
    var json = JSON.parse(reply);
    if ( json.ok == false ) {
        alert("Failed to close a document");
        return;
    }  
    delete LW.Session.openDocs[uri];
};

// Adds a new document to the session
LW.Session.openView = function(id, query) {
    // This document is already open?
    if ( LW.Session.openViews[id] ) {
        alert("View ID used multiple times")
    }
    LW.Rpc.get("/_session/_openView?id=" + id + "&" + query, function(reply) { LW.Session.openViewCallback_(reply, id); }, LW.Session.sessionErrCallback_);
};

// Response to the request of adding a document to the session
LW.Session.openViewCallback_ = function(reply, id) {
    console.log("openView: " + reply)
    var json = JSON.parse(reply);
    if ( json.ok == false ) {
        alert("Failed to open a view");
        return;
    }
    LW.Session.openViews[id] = true;
};

// Adds a new document to the session
LW.Session.closeView = function(id) {
    // This document is already open?
    if ( !LW.Session.openDocs[id] ) {
        return
    }
    // TODO: proper escaping of the URI
    LW.Rpc.get("/_session/_closeView?id=" + id, function(reply) { LW.Session.closeViewCallback_(reply, id); }, LW.Session.sessionErrCallback_);
};

// Response to the request of adding a document to the session
LW.Session.closeCallback_ = function(reply, id) {
    // console.log("close: " + reply)
    var json = JSON.parse(reply);
    if ( json.ok == false ) {
        alert("Failed to close a view");
        return;
    }  
    delete LW.Session.openViews[id];
};
