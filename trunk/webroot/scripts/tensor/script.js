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
LW.Tensor.createColumnContent_ = function(item, nextlist) {
  var id = item.context.id;
  var list = nextlist.toArray()[0];
  var title = false;
  var comments = null;
  // Clicked on a conversation in the inbox?
  if ( id.indexOf('!') == -1 ) {
	LW.Tensor.currentDoc = LW.Inbox.getOrCreateDoc(id);
	title = true;
	comments = LW.Tensor.currentDoc.content._data.comments;
  } else {
	comments = LW.Inbox.getElementById(id).comments;
  }
  // Generate the HTML for the column
  console.log("Create Comments View for " + comments._id)
  list.innerHTML = "";
  list.objectid = LW.Tensor.currentDoc.url + "!" + comments._id;
  for( var i = 0; i < comments.length; i++ ) {
	LW.Tensor.commentModifiedCallback_(comments, i)
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
  LW.Tensor.createColumnContent_(item, nextlist);
  nextlist.fadeIn(235);
};

LW.Tensor.reload_ = function(nextlist, nextnextlist, selected, item) {
  nextnextlist.fadeOut(300);
  selected.removeClass('selected'); 
  item.addClass('selected');
  nextlist.fadeOut(300, function() {
	nextlist.children('.selected').removeClass('selected');
	nextlist.removeClass('grey');
    LW.Tensor.createColumnContent_(item, nextlist);
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

// Creates and sends an initial document mutation to create a new conversation document
LW.Tensor.createNewConversation = function()
{
  if ( !LW.Session.sessionCreated ) {
	alert("No session created yet");
	return;
  }
    
  // Instantiate the document
  var url = "/" + LW.Rpc.domain + "/" + LW.Inbox.uniqueId();
  var doc = LW.Inbox.getOrCreateDoc(url);
  LW.Tensor.currentDoc = doc;
  doc.content._data._cb_comments = function(d, obj, key, mutation, event ) {
	if ( event == LW.JsonOT.AttributeInserted ) {
	  var list = $('#list-2');
	  list.toArray()[0].innerHTML = "";
	  list.toArray()[0].objectid = LW.Tensor.currentDoc.url + "!" + d._data.comments._id;
	  list.fadeIn();
	  d._data.comments._cb_inserted = LW.Tensor.commentInsertedCallback_;
	}
  }
  doc.submitDocMutation( {"_rev":0, "_meta":{"$object":true, "participants":[LW.Rpc.user + "@" + LW.Rpc.domain]},
						"_data":{"$object":true, "title":"A new document", "comments":[
								  {"content":"Hallo Welt, das ist ein neues Dokument mit einem sehr langen Text, der eigentlich in der Inbox nicht komplett zu sehen sein sollte!",
								   "comments":[],
								   "_meta":{"author":LW.Rpc.user + "@" + LW.Rpc.domain, "date":"Dec 4"}
								  }]}});
  
  // Open the document in the session
  LW.Session.open(url, false);
};

// Creates and sends a document mutation to insert a new comment.
//
// @param objectid is the ID denoting a JSON array that contains a list of comments.
LW.Tensor.createNewComment = function(objectid) {
  if ( !LW.Session.sessionCreated ) {
	alert("No session created yet");
	return;
  }
  var i = objectid.indexOf("!");
  var id = objectid.substring(i + 1, objectid.length);
  console.log("New Comment for " + objectid);
  console.log(LW.Tensor.currentDoc.content);
  var comments = LW.Inbox.getElementById(objectid);
  var mutation = [{"content":"Hallo Welt, das ist ein neuer Kommentar", "_meta":{"author":LW.Rpc.user + "@" + LW.Rpc.domain, "date":"Dec 4"}, "comments":[]}];
  if ( comments.length > 0 ) {
	mutation.splice(0,0, {"$skip":comments.length});
  }
  var arrmut = {"$array":mutation};	
  var mut = LW.Tensor.currentDoc.createMutationForId(comments._id, arrmut);
  LW.Tensor.currentDoc.submitDocMutation( mut );
};

// Install the event handlers for a comment that has been inserted in an array at the specified index.
LW.Tensor.commentInsertedCallback_ = function(doc, arr, index, mut, event) {
  arr[index]._cb = function(d, obj, key, mut, event) {
	if ( event == LW.JsonOT.ObjectModified ) {
	  LW.Tensor.commentModifiedCallback_(arr, index);
	} else if ( event == LW.JsonOT.AttributeInserted && key == "comments" ) {
	  obj.comments._cb_inserted = LW.Tensor.commentInsertedCallback_;
	}
  }
};

// Called when a comment has changed or been inserted
LW.Tensor.commentModifiedCallback_ = function(arr, index) {
  var comment = arr[index];
  var newreplies = comment.comments.length;
  // TODO: newreplies
  var title = "";
  if ( arr == LW.Tensor.currentDoc.content._data.comments && index == 0 ) {
	title = LW.Tensor.currentDoc.content._data.title + " ";
  }
  var html = '<h3><span class="text">' + esc(comment._meta.author) + ": " + title + '</span> <span class="updates">' + (newreplies > 0 ? ('(' + newreplies.toString() + ')') : "") + '</span> <span class="date"> ' + comment._meta.date + ' </span></h3>';
  html += '<h4>' + esc(comment.content) + '</h4>';
  html += '<div class="tools">';
  html += '<span class="view"><a href="#">View</a></span><span class="reply"><a href="#">Reply</a></span><span class="history"><a href="#">History</a></span><span class="edit"><a href="#">Edit</a></span>';
  html += '</div>';
  var div = document.getElementById(LW.Tensor.currentDoc.url + "!" + comment._id);
  if ( !div ) {
	console.log("Comment has been inserted: " + JSON.stringify(comment) + " at position " + index.toString());
	var list = $('.list[objectid=' + LW.Tensor.currentDoc.url + "!" + arr._id + "]").toArray()[0];
	div = document.createElement("div")
	div.className = "wave new";
	div.innerHTML = html;
	div.id = LW.Tensor.currentDoc.url + "!" + comment._id;
	div.onclick = LW.Tensor.onDivClick_;
	list.insertBefore( div, list.children[index + 1] );
  } else {
	div.innerHTML = html;
  }
};

// Invoked when an item of the inbox has changed or been inserted
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
	  $(div).addClass("selected");
	}
  } else {
	div.innerHTML = html;
  }
};

// Install the event handlers for the Inbox
LW.Tensor.init = function() {
  LW.Inbox.init();
  // Wait for the data object
  LW.Inbox.self.content._cb_data = function(doc, obj, key, mut, event) {
	if ( event == LW.JsonOT.AttributeInserted ) {
	  // Wait for the "docs" object
	  doc._data._cb_docs = function(doc, obj, key, mut, event) {
		if ( event == LW.JsonOT.AttributeInserted ) {
		  // Wait for documents being inserted in the "docs" object
		  doc._data.docs._cb_inserted = function(doc, arr, index, mut, event) {
			// Wait for changes in the inserted docs object
			arr[index]._cb = function(doc, obj, key, mut, event) {
			  if ( event == LW.JsonOT.ObjectModified ) {
				LW.Tensor.inboxModifiedCallback_(obj);
			  }
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
  });

  $('.wave').click( LW.Tensor.onDivClick_ );
});
