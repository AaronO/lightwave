/*
 * Author: Kai Chang, Torben Weis
 * 
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

if ( !window.LW ) {
  LW = { };
}

LW.Tensor = {
  currentDoc : null
};

// Called when a new column is shown
LW.Tensor.createContent_ = function(item, nextlist) {
  var id = item.context.id;
  // Clicked on a conversation in the inbox?
  if ( id.indexOf('!') == -1 ) {
	LW.Tensor.currentDoc = LW.Inbox.getOrCreateDoc(id);
	LW.Tensor.createConversationView();
  } else {
	var reply = LW.Inbox.getElementById(id);
	LW.Tensor.createCommentsView(nextlist.toArray()[0], reply);
  }
};

LW.Tensor.deselectAll_ = function(list) {
  LW.Tensor.currentDoc = null;
  var content = $('#content');
  content.animate({left : '+=' + content.position().left}, 240);
  var i = 1;
  while( true ) {
	var list = $('#list-' + i.toString());
	if ( list.length == 0 ) {
	  break;
	}
	list.removeClass('grey');
	list.children('.selected').removeClass('selected');
	if ( i > 1 ) {
	  list.fadeOut();
	}
	i++;
  }
};

LW.Tensor.deselect_ = function(nextlist, nextnextlist, list, selected) {
  nextnextlist.fadeOut();
  list.removeClass('grey');
  selected.removeClass('selected');
  nextlist.fadeOut(240, function() {
	nextlist.removeClass('grey');
	nextlist.children('.selected').removeClass('selected');
  });
};

LW.Tensor.select_ = function(nextlist, list, selected, item) {
  list.addClass('grey');
  selected.removeClass('selected'); 
  item.addClass('selected');
  $('#container').scrollTop(0);
  LW.Tensor.createContent_(item, nextlist);
  nextlist.fadeIn(235);
};

LW.Tensor.reload_ = function(nextlist, nextnextlist, selected, item) {
  nextnextlist.fadeOut(300);
  selected.removeClass('selected'); 
  item.addClass('selected');
  nextlist.fadeOut(300, function() {
	nextlist.children('.selected').removeClass('selected');
	nextlist.removeClass('grey');
    LW.Tensor.createContent_(item, nextlist);
  } );
  $('html').scrollTop(0);
  nextlist.fadeIn();
};

LW.Tensor.shiftleft_ = function(col, nextlist) {
  nextcol  = nextlist.parent();
  colwidth = col.width();
  $('#leftcol').attr('id', '');
  $('#midcol').attr('id', 'leftcol');
  $('#rightcol').attr('id', 'midcol');
  nextcol.attr('id', 'rightcol');
  $('#content').animate({left : '-=' + colwidth}, 340);
};

LW.Tensor.shiftright_ = function(col, prevlist) {
  prevcol  = prevlist.parent();
  colwidth = col.width();
  $('#rightcol').attr('id', '');
  $('#midcol').attr('id', 'rightcol');
  $('#leftcol').attr('id', 'midcol');
  prevcol.attr('id', 'leftcol');
  $('#content').animate({left : '+=' + colwidth}, 340);
};

LW.Tensor.onDivClick_ = function() {
  list     = $(this).parent();
  col      = list.parent();
  selected = list.children('.selected');

  numlist  = parseInt(list.attr('id').substring(5));

  // Click in the inbox?
  if ( numlist == 1 ) {
	LW.Session.open(this.id, true, false);
  }
  
  nextnum  =  numlist + 1;
  nextlist = $('#list-' + nextnum);
  nextnextnum =  numlist + 2;
  nextnextlist = $('#list-' + nextnextnum);
    
  if ( numlist > 1 ) {
	prevnum  = numlist - 1;
	prevlist = $('#list-' + prevnum);
	if ( numlist > 2 ) {
	  prevprevnum  = numlist - 2;
	  prevprevlist = $('#list-' + prevprevnum);
	}
  }

  if ( col.attr('id') == 'rightcol' ) {
	LW.Tensor.shiftleft_(col, nextlist);
	LW.Tensor.select_(nextlist, list, selected, $(this));
  }
  else if ( col.attr('id') == 'midcol' ) {
	if ( $(this).hasClass('selected') ) {
	  LW.Tensor.deselect_(nextlist, nextnextlist, list, selected);
	  if ( numlist > 2) {
		LW.Tensor.shiftright_(col, prevprevlist);
	  }
	} else if ( list.children().hasClass('selected') ) {
	  LW.Tensor.reload_(nextlist, nextnextlist, selected, $(this));
	} else {
	  LW.Tensor.select_(nextlist, list, selected, $(this));
	}
  }
  else if ( col.attr('id') == 'leftcol' ) {
	if ( $(this).hasClass('selected') ) {
	  LW.Tensor.deselect_(nextlist, nextnextlist, list, selected);
	  if ( numlist > 1) {
		LW.Tensor.shiftright_(col, prevlist);
	  }
	} else if ( list.children().hasClass('selected') ) {
	  LW.Tensor.reload_(nextlist, nextnextlist, selected, $(this));
	  if ( numlist > 1) {
		LW.Tensor.shiftright_(col, prevlist);
	  }
	} else {
	  LW.Tensor.select_(nextlist, list, selected, $(this));
	}
  }
};

LW.Tensor.createNewConversation = function()
{
  if ( !LW.Session.sessionCreated ) {
	alert("No session created yet");
	return;
  }
    
  // Instantiate the document
  var url = "/" + LW.Rpc.domain + "/" + LW.Inbox.uniqueId();
  var doc = LW.Inbox.getOrCreateDoc(url);
  doc.submitDocMutation( {"_rev":0, "_meta":{"participants":[LW.Rpc.user + "@" + LW.Rpc.domain]}, "_data":{"title":"A new document", "content":"Hallo Welt, das ist ein neues Dokument mit einem sehr langen Text, der eigentlich in der Inbox nicht komplett zu sehen sein sollte!", "author":LW.Rpc.user, "date":"Dec 4", "replies":[]}} )
  doc.content._data.replies._cb_inserted = LW.Tensor.commentInsertedCallback_;
  LW.Tensor.currentDoc = doc;
  
  // Open the document in the session
  LW.Session.open(url, false);
};

LW.Tensor.createNewComment = function(objectid) {
  if ( !LW.Session.sessionCreated ) {
	alert("No session created yet");
	return;
  }
  var i = objectid.indexOf("!");
  var id = objectid.substring(i + 1, objectid.length);
  console.log("New Comment for " + objectid);
  console.log(LW.Tensor.currentDoc.content);
  var arr = LW.Inbox.getElementById(objectid);
  var mutation = [{"content":"Hallo Welt, das ist ein neuer Kommentar", "author":LW.Rpc.user, "date":"Dec 4", "replies":[]}];
  if ( arr.length > 0 ) {
	mutation.splice(0,0, {"$skip":arr.length});
  }
  var arrmut = {"$array":mutation};
  var mut = LW.Tensor.currentDoc.createMutationForId(id, arrmut);
  LW.Tensor.currentDoc.submitDocMutation( mut );
};

LW.Tensor.commentInsertedCallback_ = function(doc, arr, index, mut, event) {
  arr[index]._cb = function(d, obj, key, mut, event) {
	if ( event == LW.JsonOT.ObjectModified ) {
	  LW.Tensor.commentModifiedCallback_(arr, index);
	} else if ( event == LW.JsonOT.AttributeInserted && key == "replies" ) {
	  obj.replies._cb_inserted = LW.Tensor.commentInsertedCallback_;
	}
  }
};

LW.Tensor.commentModifiedCallback_ = function(arr, index) {
  var reply = arr[index];
  var html = '<h3><span class="text">' + esc(reply.author) + ":" + '</span> <span class="updates">' + (newreplies > 0 ? ('(' + newreplies.toString() + ')') : "") + '</span> <span class="date"> ' + reply.date + ' </span></h3>';
  html += '<h4>' + esc(reply.content) + '</h4>';
  html += '<div class="tools">';
  html += '<span class="view"><a href="#">View</a></span><span class="reply"><a href="#">Reply</a></span><span class="history"><a href="#">History</a></span><span class="edit"><a href="#">Edit</a></span>';
  html += '</div>';
  var div = document.getElementById(LW.Tensor.currentDoc.url + "!" + reply._id);
  if ( !div ) {
	console.log("Comment has been inserted: " + JSON.stringify(arr[index]) + " at position " + index.toString());
	var lists = $('.list[objectid=' + LW.Tensor.currentDoc.url + "!" + arr._id + "]");
	var list = lists.toArray()[0];
	var newreplies = reply.replies.length;
	div = document.createElement("div")
	div.className = "wave new";
	div.innerHTML = html;
	div.id = LW.Tensor.currentDoc.url + "!" + reply._id;
	div.onclick = LW.Tensor.onDivClick_;
	list.insertBefore( div, list.children[index + 1] );
  } else {
	div.innerHTML = html;
  }
};

LW.Tensor.inboxModifiedCallback_ = function(entry) {
  var html = '<h3><span class="text">' + esc(entry.digest) + '</span> <span class="updates">(' + "0" + ')</span> <span class="date"> ' + "today" + ' </span></h3>';
  html += '<h4>' + "Some text Bla bla bla" + ' <span class="author">' + "torben" + '</span></h4>';
  var div = document.getElementById(entry.uri);
  if ( !div ) {
	var list = document.getElementById("list-1");
	div = document.createElement("div")
	div.className = "wave new";
	div.id = entry.uri;
	div.onclick = LW.Tensor.onDivClick_;
	div.innerHTML = html;
	list.insertBefore(div, list.firstChild);
	// Is this the current document?
	if ( LW.Tensor.currentDoc && LW.Tensor.currentDoc.url == entry.uri ) {
	  div.onclick();
	}
  } else {
	div.innerHTML = html;
  }
};

LW.Tensor.createConversationView = function() {
  var doc = LW.Tensor.currentDoc;
  console.log("Create Conversation View for " + doc.content._data.replies._id)
  var list = document.getElementById("list-2");
  list.innerHTML = "";
  var html = '<h3><span class="text">' + esc(doc.content._data.author) + ": " + esc(doc.content._data.title) + '</span> <span class="updates">(' + doc.content._data.replies.length.toString() + ')</span> <span class="date"> ' + doc.content._data.date + ' </span></h3>';
  html += '<h4>' + esc(doc.content._data.content) + '</h4>';
  html += '<div class="tools">';
  html += '<span class="view"><a href="#">View</a></span><span class="reply"><a href="#">Reply</a></span><span class="history"><a href="#">History</a></span><span class="edit"><a href="#">Edit</a></span>';
  html += '</div>';
  var div = document.createElement("div")
  div.className = "wave new";
  div.innerHTML = html;
  div.id = doc.url + "!" + doc.content._data._id;
  list.appendChild(div);
  list.objectid = doc.url + "!" + doc.content._data.replies._id;
  
  for( var i = 0; i < doc.content._data.replies.length; i++ ) {
	var reply = doc.content._data.replies[i];
	var newreplies = reply.replies.length;
	html = '<h3><span class="text">' + esc(reply.author) + ":" + '</span> <span class="updates">' + (newreplies > 0 ? ('(' + newreplies.toString() + ')') : "") + '</span> <span class="date"> ' + reply.date + ' </span></h3>';
	html += '<h4>' + esc(reply.content) + '</h4>';
	html += '<div class="tools">';
	html += '<span class="view"><a href="#">View</a></span><span class="reply"><a href="#">Reply</a></span><span class="history"><a href="#">History</a></span><span class="edit"><a href="#">Edit</a></span>';
	html += '</div>';
	var div = document.createElement("div")
	div.className = "wave new";
	div.innerHTML = html;
	div.id = doc.url + "!" + reply._id;
	div.onclick = LW.Tensor.onDivClick_;
	list.appendChild(div);
  }
};

LW.Tensor.createCommentsView = function(list, obj) {
  console.log("Create Comments View for " + obj.replies._id)
  console.log(list)
  list.innerHTML = "";
  list.objectid = LW.Tensor.currentDoc.url + "!" + obj.replies._id;
  for( var i = 0; i < obj.replies.length; i++ ) {
	LW.Tensor.commentModifiedCallback_(obj.replies, i)
  }
};

LW.Tensor.init = function() {
  LW.Inbox.init();
  LW.Inbox.self.content._cb_data = function(doc, obj, key, mut, event) {
	doc._data._cb_docs = function(doc, obj, key, mut, event) {
	  if ( event == LW.JsonOT.AttributeInserted ) {
		doc._data.docs._cb_inserted = function(doc, arr, index, mut, event) {
		  console.log("A1");
		  arr[index]._cb = function(doc, obj, key, mut, event) {
			if ( event == LW.JsonOT.ObjectModified ) {
			  console.log("A3");
			  console.log(obj);
			  LW.Tensor.inboxModifiedCallback_(obj);
			}
		  }
		}
	  }
	}
  };
  LW.Session.init();
};

// UGLY
function esc(str) {
  // TODO
  return str;
}

$(function() {
  // Header pulldown
  // TODO: Abstract this code to keep things DRY
  var headspace = 0;
  $('header a').toggle(
    function() {
      if (headspace != 1) {
        $(this).addClass('active');
        $('header').animate({
          top : "+=150"
        }, 500, function() {
          headspace++;
        })
      }
    }, 
    function() {
      if (headspace != 0) {
        $(this).removeClass('active');
        $('header').animate({
          top : "-=150"
        }, 500, function() {
          headspace--;
        });
      }
  });

  // Footer pullup
  var footspace = 0;
  $('footer a').toggle(
    function() {
      if (footspace != 1) {
        $(this).addClass('active');
        $('footer').animate({
          bottom : "+=150"
        }, 500, function() {
          footspace++;
        })
      }
    }, 
    function() {
      if (footspace != 0) {
        $(this).removeClass('active');
        $('footer').animate({
          bottom : "-=150"
        }, 500, function() {
          footspace--;
        });
      }
  });

  // Meta pullup
  var metaspace = 0;
  $('.meta a').toggle(
    function() {
      if (metaspace != 1) {
        $(this).addClass('active');
        $('.meta').animate({
          bottom : "+=100"
        }, 500, function() {
          metaspace++;
        })
      }
    }, 
    function() {
      if (metaspace != 0) {
        $(this).removeClass('active');
        $('.meta').animate({
          bottom : "-=100"
        }, 500, function() {
          metaspace--;
        });
      }
  });

  // ------------------------------------------------
  // Added by Torben
  
  $('.newwave').click( function() {
	LW.Tensor.deselectAll_();
	LW.Tensor.createNewConversation();
  } );

  $('.newcomment').click( function() {
	var col = $(this);
	while( !col.hasClass("col") ) {
	  col = col.parent();
	}
    var list = col.children('.list').toArray()[0];
	LW.Tensor.createNewComment(list.objectid);
//						 LW.Tensor.createNewComment(LW.Tensor.currentDoc.url + "!" + LW.Tensor.currentDoc.content._data._id);
  });

  $('.wave').click( LW.Tensor.onDivClick_ );
});
