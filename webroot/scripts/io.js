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
  sessionCreated : false,
  content : { },
  version : 0
};

LW.Session.init = function() {
  if ( !LW.Session.sessionCreated ) {
	LW.Session.createSession_();
  } else {
	LW.Session.sessionPoll_();
  }
};

// Create a new session at the server
LW.Session.createSession_ = function() {
  LW.Session.id = "s" + Math.random().toString();
  var mutation = {"_rev":0, "_data":{"filters":{}}};
  LW.JsonOT.applyDocMutation( LW.Session.content, mutation, 0);
  LW.Rpc.enqueue("/client/_session/" + LW.Rpc.user + "/" + LW.Session.id, JSON.stringify(mutation), LW.Session.createSessionCallback_, LW.Session.createSessionErrCallback_);
};

// Callback for the successful attempt of creating a session on the server
LW.Session.createSessionCallback_ = function(reply) {
  // console.log("create session: " + reply)
  var json = JSON.parse(reply)
  if ( json.ok == true ) {
	LW.Session.sessionCreated = true;
	LW.Session.version = json.version;
	// Open the inbox
	LW.Session.open("/" + LW.Rpc.domain + "/_user/" + LW.Rpc.user + "/inbox", true, false )
	// Start polling the session to get updates on the documents
	LW.Session.sessionPoll_();
  } else {
	alert("Failed to create a session");
  }
};

// Callback for the failed attempt of creating a session on the server
LW.Session.createSessionErrCallback_ = function() {
  alert("Offline", "You seem to be offline");
};
	
// HTTP long call to get updates on the documents
LW.Session.sessionPoll_ = function() {
  LW.Rpc.get("/client/_session/" + LW.Rpc.user + "/" + LW.Session.id + "/_poll", LW.Session.sessionPollCallback_, LW.Session.sessionPollErrCallback_);  
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
LW.Session.sessionPollErrCallback_ = function() {
  alert("Offline", "You seem to be offline");
};
	
// Adds a new document to the session
LW.Session.open = function(prefix, snapshot, recursive, mimeTypes, schemas) {
  if ( !mimeTypes ) {
	mimeTypes = [];
  }
  if ( !schemas ) {
	schemas = [];
  }
  // This document is already open?
  if ( LW.Session.content._data.filters[prefix] ) {
	return
  }
  var mutation = { "_rev":LW.Session.version, "_data":{ "$object":true, "filters":{"$object":true} } };
  mutation._data.filters[prefix] = {"recursive":recursive, "mimeTypes":mimeTypes, "schemas":schemas, "snapshot":snapshot};
  LW.JsonOT.applyDocMutation( LW.Session.content, mutation, 0);
  // console.log("-> " + JSON.stringify(mutation))
  LW.Rpc.enqueue("/client/_session/" + LW.Rpc.user + "/" + LW.Session.id, JSON.stringify(mutation), LW.Session.openCallback_);
};

// Response to the request of adding a document to the session
LW.Session.openCallback_ = function(reply) {
  // console.log("open: " + reply)
  var json = JSON.parse(reply)
  if ( json.ok == true ) {
	LW.Session.version = json.version;
  } else {
	alert("Failed to open a document");
  }  
};
