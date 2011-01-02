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
  // Points to an instance of LW.Doc
  currentDoc : null
};

// Called when a new column is shown
LW.Tensor.createColumnContent_ = function(id, list, isNewWave) {
  var comments = null;
  // Clicked on a conversation in the inbox?
  if ( id.indexOf('!') == -1 ) {
	LW.Tensor.currentDoc = LW.Inbox.getOrCreateDoc(id);
	LW.Tensor.currentDoc.content._data._cb_comments = function(doc, obj, key, mutations, event) {
	  if ( event == LW.JsonOT.AttributeInserted ) {
		document.getElementById("list-2").objectid = LW.Tensor.currentDoc.url + "!" + LW.Tensor.currentDoc.content._data.comments._id;
		LW.Tensor.currentDoc.content._data.comments._cb_inserted = LW.Tensor.commentInsertedCallback_;
	  }
	}
	// Open the document if it is not yet open
	// LW.Session.open(LW.Tensor.currentDoc.url, true, false);
	comments = LW.Tensor.currentDoc.content._data.comments;
  } else {
	comments = LW.Inbox.getElementById(id).comments;
  }
  // Clear the HTML for this columns
  list.innerHTML = "";
  // Is the document already loaded? -> Show it
  if ( comments ) {
	// Generate the HTML for the column
	list.objectid = LW.Tensor.currentDoc.url + "!" + comments._id;
	for( var i = 0; i < comments.length; i++ ) {
	  LW.Tensor.commentModifiedCallback_(comments, i)
	}
  }
  // Create the HTML for the edit box that allows users to type new comments
  var box = document.createElement("div");
  var html = '<div class="editor">';
  html += '<div class="title"><span class="label">Title:</span> <input type="text" class="titleInput"></input></div>';
  html += '<div><textarea></textarea></div>';
  html += '<div><span><a href="#">Submit</a> <a href="#">Cancel</a></span></div>';
  html += '</div>';
  html += '<div class="info"><a href="#">Click here to reply</a></div>';
  box.innerHTML = html;
  box.className = "infoBox";
  // Event handlers for the edit box.
  box.lastChild.firstChild.onclick = function() {
	box.className = "inputBox";
	if ( LW.Tensor.currentDoc.content._data.comments && LW.Tensor.currentDoc.content._data.comments.length == 0 ) {
	  box.firstChild.firstChild.lastChild.focus();
	} else {
	  var textarea = box.firstChild.children[1].firstChild;
	  textarea.focus();
	}
	return false;
  };
  // Cancel clicked
  box.firstChild.lastChild.firstChild.lastChild.onclick = function() {
	box.className = "infoBox";
	return false;
  };
  // Submit clicked
  box.firstChild.lastChild.firstChild.firstChild.onclick = function() {
	if ( LW.Tensor.currentDoc.url == id && LW.Tensor.currentDoc.content._data.comments && LW.Tensor.currentDoc.content._data.comments.length == 0 ) {
	  var title = box.firstChild.firstChild.lastChild;
	  LW.Tensor.setTitle(title.value);
	  title.value = "";
	}
	var textarea = box.firstChild.children[1].firstChild;
	LW.Tensor.createNewComment(list.objectid, textarea.value);
	textarea.value = "";
	textarea.focus();
	return false;
  };
  list.appendChild(box);
  if ( isNewWave ) {
	box.className = "inputBox";
	box.firstChild.firstChild.lastChild.focus();
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
	  list.fadeOut(0);
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
  LW.Tensor.createColumnContent_(item.context.id, nextlist.toArray()[0]);
  nextlist.fadeIn(240);
};

LW.Tensor.reload_ = function(nextlist, nextnextlist, selected, item) {
  nextnextlist.fadeOut(240);
  selected.removeClass('selected'); 
  item.addClass('selected');
  nextlist.fadeOut(240, function() {
	nextlist.children('.selected').removeClass('selected');
	nextlist.removeClass('grey');
    LW.Tensor.createColumnContent_(item.context.id, nextlist.toArray()[0]);
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

// Called when a user clicks on an entry in the inbox or a comment.
// "this" is the DIV element the user clicked on.
LW.Tensor.onDivClick_ = function() {
  list     = $(this).parent();
  col      = list.parent();
  selected = list.children('.selected');

  numlist  = parseInt(list.attr('id').substring(5));
  
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
  
  if ( numlist == 1 && $(this).hasClass('selected') ) {
	LW.Session.open(this.id, true, false);
  }
};

// Creates and sends a document mutation to insert a new comment.
//
// @param objectid is the ID denoting a JSON array that contains a list of comments.
LW.Tensor.createNewComment = function(objectid, text) {
  if ( !text ) {
	text = "Hallo Welt, das ist ein neuer Kommentar";
  }
  var i = objectid.indexOf("!");
  var id = objectid.substring(i + 1, objectid.length);
  // console.log("New Comment for " + objectid);
  // console.log(LW.Tensor.currentDoc.content);
  var comments = LW.Inbox.getElementById(objectid);
  var mutation = [{"content":text, "_meta":{"author":LW.Rpc.user + "@" + LW.Rpc.domain, "date":"Dec 4"}, "comments":[]}];
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

LW.Tensor.setTitle = function(title) {
  var datamut = {"$object":true, "title":title};
  var mut = LW.Tensor.currentDoc.createMutationForId(LW.Tensor.currentDoc.content._data._id, datamut);
  LW.Tensor.currentDoc.submitDocMutation( mut );
};

// Called when a comment has changed or been inserted
LW.Tensor.commentModifiedCallback_ = function(arr, index) {
  // console.log("Comment modified callback");
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
	// console.log("Comment has been inserted: " + JSON.stringify(comment) + " at position " + index.toString());
	var list = $('.list[objectid=' + LW.Tensor.currentDoc.url + "!" + arr._id + "]").toArray()[0];
	// Do not display the comment?
	if ( !list ) {
	  return;
	}
	div = document.createElement("div")
	div.className = "wave new";
	div.innerHTML = html;
	div.id = LW.Tensor.currentDoc.url + "!" + comment._id;
	div.onclick = LW.Tensor.onDivClick_;
	list.insertBefore( div, list.children[index] );
  } else {
	div.innerHTML = html;
  }
};

// Invoked when an item of the inbox has changed or been inserted
LW.Tensor.inboxModifiedCallback_ = function(entry) {
  var html = '<h3><span class="text">' + "Georg...Heinz,Fritz" + '</span> <span class="updates">(' + "0" + ')</span> <div style="margin-left:4px; float:right"><img src="http://www.uni-due.de/favicon.ico"></div><span class="date"> ' + "uni-due.de" + ' </span></h3>';
  html += '<h4>' + esc(entry.digest) + ' <span class="author">' + "Today" + '</span></h4>';
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
  document.getElementById("username").innerText = LW.Rpc.user + "@" + LW.Rpc.domain;
  LW.Inbox.init();
  LW.Inbox.self.content._data._cb_docs = function(doc, obj, key, mut, event) {
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
  LW.Session.init();
};

// UGLY
function esc(str) {
  // TODO
  return str;
}

$(function() {
  // Header pulldown
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
	if ( !LW.Session.sessionCreated ) {
	  alert("No session created yet");
	  return;
	}
	LW.Tensor.deselectAll_();  
	// Instantiate the document
	var url = "/" + LW.Rpc.domain + "/" + LW.Inbox.uniqueId();
	var doc = LW.Inbox.getOrCreateDoc(url);
	LW.Tensor.createColumnContent_(doc.url, document.getElementById("list-2"), true);
	$("#list-2").fadeIn();
	doc.submitDocMutation( {"_rev":0, "_meta":{"$object":true, "participants":[LW.Rpc.user + "@" + LW.Rpc.domain]},
							"_data":{"$object":true, "title":"", "comments":[]}});
//					  doc.submitDocMutation( {"_rev":0, "_meta":{"$object":true, "participants":[LW.Rpc.user + "@" + LW.Rpc.domain]},
//											"_data":{"$object":true, "title":"A new document", "comments":[
//																										   {"content":"Hallo Welt, das ist ein neues Dokument mit einem sehr langen Text, der eigentlich in der Inbox nicht komplett zu sehen sein sollte!",
//																										   "comments":[],
//																										   "_meta":{"author":LW.Rpc.user + "@" + LW.Rpc.domain, "date":"Dec 4"}
//																										   }]}});
					  LW.Session.open(LW.Tensor.currentDoc.url, false, false);
  } );
  /*
  $('.newcomment').click( function() {
	if ( !LW.Tensor.currentDoc ) {
	  alert("You must select a document first");
	}
	var col = $(this);
	while( !col.hasClass("col") ) {
	  col = col.parent();
	}
    var list = col.children('.list').toArray()[0];
	LW.Tensor.createNewComment(list.objectid);
  });
   */
});
