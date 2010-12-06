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

LW.Tensor.createContent_ = function(item, nextlist) {
  var id = item.context.id;
  // Clicked on a conversation in the inbox?
  if ( id.indexOf('!') == -1 ) {
	LW.Tensor.currentDoc = LW.Inbox.getOrCreateDoc(id);
	// TODO: The doc should be opened in a session
	LW.Tensor.createConversationView();
  } else {
	// TODO
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
  LW.Tensor.createContent_(item);
  nextlist.fadeIn(235);
};

LW.Tensor.reload_ = function(nextlist, nextnextlist, selected, item) {
  nextnextlist.fadeOut(300);
  selected.removeClass('selected'); 
  item.addClass('selected');
  nextlist.fadeOut(300, function() {
	nextlist.children('.selected').removeClass('selected');
	nextlist.removeClass('grey');
    LW.Tensor.createContent_(item);
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
  var doc = LW.Inbox.getOrCreateDoc("/localhost/foo");
  doc.submitDocMutation( {"_rev":0, "_data":{"title":"A new document", "content":"Hallo Welt, das ist ein neues Dokument", "author":LW.Rpc.user, "date":"Dec 4", "replies":[
	{content:"This does not make much sense", author:"Joe", date:"Dec 5", replies:[]},
	{content:"I like it never the less", author:"Jack", date:"Dec 6", replies:[]},
  ]}} )
  doc.content._data._cb_content = LW.Tensor.contentCallback_;
  
  // Show the document in the inbox
  LW.Tensor.createDocumentInboxView(doc);
  
  // Open the document in the session
  LW.Session.open("/localhost/foo", false);
};

LW.Tensor.contentCallback_ = function(doc, obj, attr, mutation, action ) {
  console.log("Content has been modified: " + obj[attr]);
  // TODO
  doc._dom.getElementsByTagName("span")[0].innerHTML = obj[attr];
};

LW.Tensor.createDocumentInboxView = function(doc) {
  var list = document.getElementById("list-1");
  var html = '<h3><span class="text">' + esc(doc.content._data.title) + '</span> <span class="updates">(' + doc.content._data.replies.length.toString() + ')</span> <span class="date"> ' + doc.content._data.date + ' </span></h3>';
  html += '<h4>' + esc(doc.content._data.content) + ' <span class="author">' + doc.content._data.author + '</span></h4>';
  var div = document.createElement("div")
  div.className = "wave new";
  div.innerHTML = html;
  div.id = doc.url;
  div.onclick = LW.Tensor.onDivClick_;
  list.insertBefore(div, list.firstChild);
  doc.content._dom = div;
};

LW.Tensor.createConversationView = function() {
  var doc = LW.Tensor.currentDoc;
  var list = document.getElementById("list-2");
  var html = '<div id="' + doc.url + "!" + doc.content._data._id + '" class="wave new"><h3><span class="text">' + esc(doc.content._data.author) + ": " + esc(doc.content._data.title) + '</span> <span class="updates">(' + doc.content._data.replies.length.toString() + ')</span> <span class="date"> ' + doc.content._data.date + ' </span></h3>';
  html += '<h4>' + esc(doc.content._data.content) + '</h4>';
  html += '<div class="tools">';
  html += '<span class="view"><a href="#">View</a></span><span class="reply"><a href="#">Reply</a></span><span class="history"><a href="#">History</a></span><span class="edit"><a href="#">Edit</a></span>';
  html += '</div></div>';
  for( var i in doc.content._data.replies ) {
	var reply = doc.content._data.replies[i];
	var newreplies = reply.replies.length;
	html += '<div id="' + doc.url + "/" + doc.content._data._id + '"class="wave new"><h3><span class="text">' + esc(reply.author) + ":" + '</span> <span class="updates">' + (newreplies > 0 ? ('(' + newreplies.toString() + ')') : "") + '</span> <span class="date"> ' + reply.date + ' </span></h3>';
	html += '<h4>' + esc(reply.content) + '</h4>';
	html += '<div class="tools">';
	html += '<span class="view"><a href="#">View</a></span><span class="reply"><a href="#">Reply</a></span><span class="history"><a href="#">History</a></span><span class="edit"><a href="#">Edit</a></span>';
	html += '</div></div>';
  }
  list.innerHTML = html;
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
  
  $('.newwave').click( LW.Tensor.createNewConversation );

  $('.wave').click( LW.Tensor.onDivClick_ );
});
