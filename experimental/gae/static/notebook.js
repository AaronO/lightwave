function Book(id, text) {
    this.id = id;
    this.text = text;
    // An array of Chapter objects
    this.chapters = [];
    this.currentChapter = null;

    this.inbox = new Chapter(this, "inbox", "Inbox", 0, null);
    this.addChapter(this.inbox, true);
    this.setActiveChapter(this.chapters[0]);
}

function Chapter(book, id, text, colorScheme, after) {
    this.book = book;
    this.id = id;
    this.text = text;
    this.after = after;
    // An array of Page objects
    this.pages = [];
    this.colorScheme = colorScheme;
    this.tab = null;
    this.currentPage = null;
}

function Page(chapter, id, text, after) {
    this.chapter = chapter;
    this.id = id;
    this.text = text;
    this.after = after;
    this.vtab = null;
    this.nextSeq = 0;
    this.layout = null;
    // An array of PageContent objects
    this.contents = [];
    // An array of Follower objects
    this.followers = [];
    // An array of Follower objects
    this.invitations = [];
}

function PageContent(page, id, text, cssClass, style) {
    this.page = page;
    this.id = id;
    this.text = text;
    this.cssClass = cssClass;
    this.style = {};
    this.applyStyle_(style);
    this.paragraphs = [];
    this.tombs = [text.length];
    this.buildParagraphs();
}

function PageLayout(page, id, style) {
    this.page = page;
    this.id = id;
    this.style = style;
}

function Follower(page, userid, username) {
    this.page = page;
    this.id = userid;
    this.name = username;
    if (!this.name) {
        this.name = this.id.substr(0, this.id.indexOf("@"));
    }
}

// Returns an entity based on its entity blobref
Book.prototype.getChapter = function(id) {
    for( var i = 0; i < this.chapters.length; i++) {
        if (this.chapters[i].id == id) {
            return this.chapters[i];
        }
    }
    return null;
};

Book.prototype.addChapter = function(c, make_active) {
    // Determine where to insert the chapter
    var pos = 0;
    if (c.after) {
        for( var i = 0; i < this.chapters.length; i++) {
            var x = this.chapters[i];
            if (x.id == c.after) {
                pos = i + 1;
                break;
            }
        }
    }
    var to_pos = pos;
    for( var i = pos; i < this.chapters.length; i++) {
        var x = this.chapters[i];
        if (x.id != c.after) {
            break;
        }
        to_pos++;
    }
    for( var i = pos; i < to_pos; i++) {
        var x = this.chapters[i];
        if (x.id < id) {
            pos++;
        } else {
            break;
        }
    }
    this.chapters.splice(pos, 0, c);
    // Show the chapter-tab in the UI
    c.renderTab();
    var tabs = document.getElementById("tabs");
    tabs.insertBefore(c.tab, tabs.children[pos]);
    this.positionChapterTabs_();
    if (make_active) {
        this.setActiveChapter(c);
    }
};

Book.prototype.positionChapterTabs_ = function() {
    // Shift all tabs on the right (including the '+' tab) ....
    var tabs = document.getElementById("tabs");
    for( var i = 0; i < tabs.children.length; i++) {
        var t = tabs.children[i];
        t.style.left = (-i * 3).toString() + "px";
        t.style.zIndex = 99 - i;
    }
};

Book.prototype.setActiveChapter = function(chapter) {
    if (chapter == this.currentChapter) {
        return;
    }
    // Deactivate a tab (if one is active)
    if (this.currentChapter) {
        this.currentChapter.close();
    }
    this.currentChapter = chapter;
    if (this.currentChapter) {
        this.currentChapter.open();
    }
};

// Get a map of all pages. The key is the page blobref.
Book.prototype.getPages = function() {
    var result = { };
    for (var i = 0; i < this.chapters.length; i++) {
        var chapter = this.chapters[i];
        for( var k = 0; k < chapter.pages.length; k++) {
            var page = chapter.pages[k];
            if (result[page.pageBlobRef]) {
                result[page.pageBlobRef] = result[page.pageBlobRef].concat(chapter);
            } else {
                result[page.pageBlobRef] = [chapter];
            }
        }
    }
    return result;
};

// Find a page by its page-entity ID
Book.prototype.getPage = function(id) {
   for (var i = 0; i < this.chapters.length; i++) {
        var chapter = this.chapters[i];
        for( var k = 0; k < chapter.pages.length; k++) {
            var page = chapter.pages[k];
            if (page.id == id) {
                return page;
            }
        }
    }
    return null;
};

Book.prototype.setUnreadInfo = function(unread) {
    for (var i = 0; i < this.chapters.length; i++) {
        var chapter = this.chapters[i];
        chapter.setUnreadInfo(unread);
    }
};

Book.prototype.setPageUnread = function(perma_blobref, unread) {
    for (var i = 0; i < this.chapters.length; i++) {
        var chapter = this.chapters[i];
        chapter.setPageUnread(perma_blobref, unread);
    }
};

// ============================================================================
// Chapter
// ============================================================================

Chapter.prototype.open = function() {
    $(this.tab).removeClass("inactivetab");
    $(this.tab).addClass("activetab");
    this.tab.style.zIndex = 100;
    var screen = document.getElementById("screen");
    screen.style.backgroundColor = colorSchemes[this.colorScheme].color;
    // Show a certain page in the inbox?
    if (this.id == "inbox") {
        if (!this.currentPage) {
            this.renderInbox_();
        }
    } else {
        // Create vtabs for pages
        var vtabs = document.getElementById("vtabs");
        for( var i = 0; i < this.pages.length; i++) {
            var p = this.pages[i];
            p.renderTab();
            vtabs.appendChild(p.vtab);
        }
        this.positionTabs_();
        if (!this.currentPage && this.pages.length > 0) {
            this.currentPage = this.pages[0];
        }
    }
    this.showUIElements_();
    if (this.currentPage) {
        this.currentPage.open();
    }
};

Chapter.prototype.close = function() {
    $(this.tab).removeClass("activetab");
    $(this.tab).addClass("inactivetab");
    if (this.id == "inbox") {
        this.closeInbox_();
    } else {
        // Remove all vtabs
        var vtabs = document.getElementById("vtabs");
        for( ; vtabs.children.length > 1;) {
            vtabs.removeChild( vtabs.children[1] );
        }
        for( var i = 0; i < this.pages.length; i++ ) {
            delete this.pages[i].tab;
        }
    }
    if (this.currentPage) {
        this.currentPage.close();
    }
};

Chapter.prototype.currentPageIndex = function() {
    for( var i = 0; i < this.pages.length; i++) {
        if (this.pages[i] == this.currentPage) {
            return i;
        }
    }
    return -1;
};

Chapter.prototype.getPageByPageBlobRef = function(blobref) {
    for( var i = 0; i < this.pages.length; i++) {
        if (this.pages[i].pageBlobRef == blobref) {
            return this.pages[i];
        }
    }
    return null;
};

/*
Chapter.prototype.removePage = function(page) {
    var index = -1
    for( var i = 0; i < this.pages.length; i++) {
        if (this.pages[i] == page) {
            index = i
            break;
        }
    }
    if (this.currentPage == page) {
        if (this.pages.length == 0) {
            this.setActivePage(null);
        } else if (index == this.pages.length - 1) {
            this.setActivePage(this.pages[index - 1]);
        } else {
            this.setActivePage(this.pages[index + 1]);
        }
        this.pages = this.pages.splice(i, 1);
    }
    if (this.id == "inbox") {
        // Remove from inbox item
        var div = document.getElementById("inboxitem-" + page.pageBlobRef);
        if (div) {
            div.parentNode.removeChild(div);
        }
    } else {
        page.vtab.parentNode.removeChild(page.vtab);
        this.positionTabs_();
    }
};
*/

Chapter.prototype.addPage = function(p, make_active) {
    // Determine where to insert the chapter
    var pos = 0;
    if (p.after) {
        for( var i = 0; i < this.pages.length; i++) {
            var x = this.pages[i];
            if (x.id == p.after) {
                pos = i + 1;
                break;
            }
        }
    }
    var to_pos = pos;
    for( var i = pos; i < this.pages.length; i++) {
        var x = this.pages[i];
        if (x.id != p.after) {
            break;
        }
        to_pos++;
    }
    for( var i = pos; i < to_pos; i++) {
        var x = this.pages[i];
        if (x.id < id) {
            pos++;
        } else {
            break;
        }
    }
    this.pages.splice(pos, 0, p);
    // Perhaps the page is unread? -> show it
    this.updateUnreadPagesCount();
    // Do not update the UI if the chapter is not visible at all
    if (this.book.currentChapter != this ) {
        return;
    }
    // Inbox?
    if ( this.id == "inbox") {
        if (!this.currentPage) {
            var div = this.renderInboxItem_(p);
            var inboxdiv = document.getElementById("inbox");
            inboxdiv.insertBefore(div, inboxdiv.children[1]);
        }
        return;
    }
    // Insert vtab
    p.renderTab();
    var vtabs = document.getElementById("vtabs");
    vtabs.insertBefore(p.vtab, vtabs.children[pos + 1]);
    this.positionTabs_();
    if (make_active) {
        this.setActivePage(p);
    }
};

Chapter.prototype.removePage = function(page) {
    var index = 0;
    for ( var i = 0; i < this.pages.length; i++) {
        if (this.pages[i] == page) {
            break;
        }
        index++;
    }
    if ( this.currentPage == page ) {
        if (this.id == "inbox") {
            this.setActivePage(null);
        } else if (this.pages.length > index + 1) {
            this.setActivePage(this.pages[index + 1]);
        } else if (index > 0) {
            this.setActivePage(this.pages[index - 1]);
        } else {
            this.setActivePage(null);
        }
    }
    if (page.vtab) {
        page.vtab.parentNode.removeChild(page.vtab);
        delete page.vtab;
    }
    if (page.inbox_div) {
        page.inbox_div.parentNode.removeChild(page.inbox_div);
        delete page.inbox_div;
    }
    this.pages.splice(index, 1);
};

Chapter.prototype.positionTabs_ = function() {
    // Shift all tabs on the right (including the '+' tab) ....
    var vtabs = document.getElementById("vtabs");
    for( var i = 1; i < vtabs.children.length; i++) {
        var t = vtabs.children[i];
        t.style.top = (-i + 1).toString() + "px";
    }
};

Chapter.prototype.showUIElements_ = function() {
    if (this.id == "inbox") {
        if (this.currentPage) {
            document.getElementById("back-to-inbox").style.display = "block";
            document.getElementById("inbox").style.display = "none";
            document.getElementById("pagecontainer").style.display = "block";
        } else {
            document.getElementById("inbox").style.display = "block";
            document.getElementById("pagecontainer").style.display = "none";
        }
        document.getElementById("newvtab").style.visibility = "hidden";
    } else {
        document.getElementById("back-to-inbox").style.display = "none";
        document.getElementById("newvtab").style.visibility = "visible";
        document.getElementById("inbox").style.display = "none";
        document.getElementById("pagecontainer").style.display = "block";
    }
};

Chapter.prototype.setActivePage = function(page) {
    if (this.currentPage == page) {
        return;
    }
    if (this.currentPage) {
        this.currentPage.close();
    } else if (this.id == "inbox") {
        this.closeInbox_();
    }
    this.currentPage = page;
    this.showUIElements_();
    if (this.currentPage) {
        this.currentPage.open();
    } else if (this.id == "inbox") {
        this.renderInbox_();
    }
};

Chapter.prototype.setText = function(text) {
    this.text = text;
    if (this.book != book || !this.tab) {
        return;
    }
    var child = this.tab.firstChild;
    // The name is currently being edited?
    if (child.nodeName == "INPUT") {
        return;
    }
    child.replaceData(0, child.nodeValue.length, this.text);
};
    
// Creates the horizontal tab for the chapter.
Chapter.prototype.renderTab = function() {
    // Insert tab
    this.tab = document.createElement("div");
    if (this.book.currentChapter == this) {
        this.tab.className = "tab tab" + this.colorScheme.toString() + " activetab";
    } else {
        this.tab.className = "tab tab" + this.colorScheme.toString() + " inactivetab";
    }
    this.tab.appendChild(document.createTextNode(this.text));
    this.showUnreadPagesCount();
    var chapter = this;
    this.tab.addEventListener("click", function(e) {
        store.waitForPageIO( function() {
            chapter.book.setActiveChapter(chapter);
        });
        return false;
    });
    this.tab.addEventListener("dblclick", function(e) {
        chapter.enableTabEditing_();
        return false;
    });
};

Chapter.prototype.enableTabEditing_ = function() {
    this.tab.removeChild(this.tab.firstChild);
    var input = document.createElement("input");
    input.type = "text";
    input.value = this.text;
    this.tab.insertBefore(input, this.tab.firstChild);
    input.focus();
    var chapter = this;
    var fixeventhandler = {done: false};
    var f = function() {
        if (fixeventhandler.done) {
            return;
        }
        fixeventhandler.done = true;
        if (chapter.text != input.value) {
            chapter.text = input.value;
            var msg = {perma: chapter.book.id, op: chapter.text, type: "mutation", entity: chapter.id, field: "title"};
            store.submit(msg, null, null, function(m) { m.entity = chapter.id });
        }
        chapter.tab.removeChild(input);
        chapter.tab.insertBefore(document.createTextNode(chapter.text), chapter.tab.firstChild);
    };
    input.addEventListener("change", f);
    input.addEventListener("blur", f);
};

Chapter.prototype.closeInbox_ = function(page) {
    var inboxdiv = document.getElementById("inbox");
    var items = document.getElementsByClassName("inboxitem");
    for( ; items.length > 0; ) {
        inboxdiv.removeChild(items[0]);
    }
};

Chapter.prototype.redrawInbox = function(page) {
    if (this.book.currentChapter != this || this.currentPage) {
        return;
    }
    for (var i = 0; i < this.pages.length; i++) {
        this.redrawInboxItem(this.pages[i]);
    }
};

Chapter.prototype.renderInbox_ = function(page) {
    var inboxdiv = document.getElementById("inbox");
    var prev;
    for (var i = 0; i < this.pages.length; i++) {
        var div = this.renderInboxItem_(this.pages[i]);
        inboxdiv.insertBefore(div, prev);
        prev = div;
    }
};

Chapter.prototype.redrawInboxItem = function(page) {
    if (this.book.currentChapter != this || this.currentPage) {
        return;
    }
    var div = document.getElementById("inboxitem-" + page.pageBlobRef);
    this.renderInboxItem_(page, div);
};

Chapter.prototype.renderInboxItem_ = function(page, div) {
    var isnew = true;
    if (div) {
        isnew = false;
        div.innerHTML = "";
    } else {
        div = document.createElement("div");
        page.inbox_div = div;
    }
    div.id = "inboxitem-" + page.pageBlobRef;
    if (page.inbox_latestauthors.length > 0) {
        div.className = "inboxitem inboxitemnew";
    } else {
        div.className = "inboxitem";
    }
    var input = document.createElement("input");
    input.className = "inboxcheckbox";
    input.type = "checkbox";
    if (page.inbox_selected) {
        input.checked = true;
        $(div).addClass("inboxitemselected");
    }
    input.addEventListener("click", function(e) {
        if (!e) var e = window.event;
 	e.cancelBubble = true;
	if (e.stopPropagation) e.stopPropagation();
        var targ;
	if (e.target) targ = e.target;
	else if (e.srcElement) targ = e.srcElement;
        if (targ && targ.checked) {
            page.inbox_selected = true;
            $(div).addClass("inboxitemselected");
        } else {
            page.inbox_selected = false;
            $(div).removeClass("inboxitemselected");
        }
    });
    div.appendChild(input);
    var span = document.createElement("span");
    span.className = "inboxauthor";
    var authors = [];
    for (var i = 0; i < page.inbox_latestauthors.length; i++) {
        // TODO: HTML escape
        authors.push("<b>" +  page.inbox_latestauthors[i] + "</b>")
    }
    for (var i = 0; i < page.inbox_authors.length; i++) {
        // TODO: HTML escape
        authors.push(page.inbox_authors[i])
    }
    for (var i = 0; i < page.inbox_followers.length; i++) {
        // TODO: HTML escape
        authors.push(page.inbox_followers[i])
    }
    span.innerHTML = authors.join(",");
    div.appendChild(span);
    var span = document.createElement("span");
    span.className = "inboxtime";
    span.innerText = "18:13";
    div.appendChild(span);
    var chapters = page.getChapters();
    for (var i = 0; i < chapters.length; i++) {
        var c = chapters[i];
        if (c.id == "inbox") {
            continue;
        }
        var label = document.createElement("div");
        label.className = "inboxlabel";
        label.innerText = c.text;
        label.style.backgroundColor = colorSchemes[c.colorScheme].color;
        div.appendChild(label);
    }
    div.appendChild(document.createTextNode(page.text));
    var chapter = this;
    if (isnew) {
        div.addEventListener("click", function() {
            book.setPageUnread(page.pageBlobRef);
            document.getElementById("inbox").style.display = "none";
            document.getElementById("pagecontainer").style.display = "block"; 
            chapter.setActivePage(page);
        });
        if (page.inbox_selected) {
            $(div).addClass("inboxitemselected");
            input.checked = true;
        }
    }
    return div;
};

Chapter.prototype.getSelectedPages = function() {
    if (this.id != "inbox") {
        if (this.currentPage) {
            return [this.currentPage];
        }
        return [];
    }
    var result = [];
    for (var i = 0; i < this.pages.length; i++) {
        var page = this.pages[i];
        if (page.inbox_selected) {
            result.push(page);
        }
    }
    return result;
};

Chapter.prototype.setUnreadInfo = function(unread) {
    for (var i = 0; i < this.pages.length; i++) {
        var page = this.pages[i];
        if (unread[page.pageBlobRef]) {
            page.setUnread(true);
            if (unread && this.id == "inbox") {
                this.redrawInboxItem(page);
            }
        } else {
            page.setUnread(false);
        }
    }
    this.updateUnreadPagesCount();
};

Chapter.prototype.setPageUnread = function(perma_blobref, unread) {
    var dirty = false
    for (var i = 0; i < this.pages.length; i++) {
        var page = this.pages[i];
        if (page.pageBlobRef == perma_blobref && page.unread != unread) {
            dirty = true
            // The current page cannot be marked unread
            if (this == this.book.currentChapter && page == this.currentPage && unread) {
                continue;
            }
            page.setUnread(unread);
            if (unread && this.id == "inbox") {
                this.redrawInboxItem(page);
            }
        }
    }
    if (dirty) {
        this.updateUnreadPagesCount();
    }
};

Chapter.prototype.updateUnreadPagesCount = function() {
    var old = this.unreadPagesCount;
    this.unreadPagesCount = 0;
    for( var i = 0; i < this.pages.length; i++ ) {
        var p = this.pages[i];
        if (p.unread) {
            this.unreadPagesCount++;
        }
    }
    if (!this.tab || old == this.unreadPagesCount) {
        return
    }
    this.showUnreadPagesCount()
};

Chapter.prototype.showUnreadPagesCount = function() {
    var span = this.tab.lastChild;
    if (!$(span).hasClass("pagesunread")) {
        span = null;
    }
    if (this.unreadPagesCount > 0) {
        if (!span) {
            span = document.createElement("span");
            span.className = "pagesunread";
            this.tab.appendChild(span);
        }
        span.innerText = this.unreadPagesCount.toString();
    } else {
        if (span) {
            this.tab.removeChild(span);
        }
    }
};

// =========================================================
// Page
// =========================================================

Page.prototype.open = function() {
    if (this.vtab) {
        $(this.vtab).removeClass("inactivevtab" + this.chapter.colorScheme.toString());
        $(this.vtab).addClass("activevtab");
    }
    this.applyLayout_();
    this.showContents();
    this.showFollowers();
    // If the page blobref is marked with "tmp-" then the page has just been created and there is no need to open it.
    if (this.pageBlobRef.substring(0,4) != "tmp-") {
        // If opened from the inbox, mark it as read immediately
        store.openPage(this, this.chapter.id == "inbox");
    }
};

Page.prototype.close = function() {
    if (this.vtab) {
        $(this.vtab).removeClass("activevtab");
        $(this.vtab).addClass("inactivevtab" + this.chapter.colorScheme.toString());
    }
    this.cleanup_();
    // Close the current page
    if (this.pageBlobRef && this.pageBlobRef.substr(0,4) != "tmp-") {
        store.closePage(this);
    }
};

Page.prototype.cleanup_ = function() {
    // Cleanup
    var pagediv = this.pageContentDiv_();
    for (var i = 0; i < this.contents.length; i++) {
        var content = this.contents[i];
        pagediv.removeChild(content.div);
        delete content.div;
    }
    // Cleanup
    var sharediv = document.getElementById("share");
    var friends = document.getElementsByClassName("friend");
    for( ; friends.length > 0; ) {
        sharediv.removeChild(friends[0]);
    }
};

Page.prototype.enterFullScreen = function() {
    this.cleanup_();
    fullscreen = true;
    this.applyLayout_();
    this.showContents();
    this.showFollowers();
};

Page.prototype.leaveFullScreen = function() {
    this.cleanup_();
    fullscreen = false;
    this.applyLayout_();
    this.showContents();
    this.showFollowers();
};

Page.prototype.pageContentDiv_ = function() {
    if (fullscreen) {
        return document.getElementById("pagecontent_fullscreen_scaled");
    }
    return document.getElementById("pagecontent");
};

Page.prototype.submitText = function(text) {
    this.setText(text);
    if (this.submitTimer) {
        return;
    }
    var page = this;
    var f = function() {
        delete page.submitTimer;
        var msg = {perma: page.chapter.book.id, op: page.text, type: "mutation", entity: page.id, field: "title"};
        store.submit(msg, null, null, function(m) { m.entity = page.id });
    };
    this.submitTimer = setTimeout( f, 2000 );
};

Page.prototype.setText = function(text) {
    this.text = text;
    if (this.chapter != book.currentChapter || !this.vtab) {
        return;
    }
    var child = this.vtab.firstChild;
    child.replaceData(0, child.nodeValue.length, this.text);
};

Page.prototype.renderTab = function() {
    this.vtab = document.createElement("div");
    if (this == this.chapter.currentPage) {
        this.vtab.className = "vtab activevtab";
    } else {
        this.vtab.className = "vtab inactivevtab" + this.chapter.colorScheme.toString();
    }
    this.vtab.appendChild(document.createTextNode(this.text));
    var p = this;
    this.vtab.addEventListener("click", function() {
        store.waitForPageIO( function() {
            p.chapter.setActivePage(p);
        });
    });
    if (this.unread) {
        var span = document.createElement("span");
        span.className = "pageunread";
        span.innerHTML = "&nbsp";
        this.vtab.appendChild(span);
    }
};

Page.prototype.addContent = function(content) {
    this.contents.push(content);
    if (!this.isVisible()) {
        return;
    }
    this.showContent(content);
};

Page.prototype.showContents = function() {
    for( var i = 0; i < this.contents.length; i++ ) {
        this.showContent(this.contents[i]);
    }
};

Page.prototype.showContent = function(content) {
    var pagecontentdiv = this.pageContentDiv_();
    if (content.cssClass == "image") {
        var f_move = function(e) {
            return LW.movableMouseDown(e, content);
        };
        var f_resize = function(e) {
            return LW.resizeMouseDown(e, content);
        };
        var div = document.createElement("div");
        content.div = div;
        div.addEventListener("mousedown", f_move, false);
        div.className = "movable picture";
        var img = document.createElement("img");
        img.src = content.text;
        div.appendChild(img);
        var r = document.createElement("div");
        r.className = "resize resize-nw";
        r.innerText = " ";
        r.addEventListener("mousedown", f_resize, false);
        div.appendChild(r);
        var r = document.createElement("div");
        r.className = "resize resize-ne";
        r.innerText = " ";
        r.addEventListener("mousedown", f_resize, false);
        div.appendChild(r);
        var r = document.createElement("div");
        r.className = "resize resize-sw";
        r.innerText = " ";
        r.addEventListener("mousedown", f_resize, false);
        div.appendChild(r);
        var r = document.createElement("div");
        r.className = "resize resize-se";
        r.innerText = " ";
        r.addEventListener("mousedown", f_resize, false);
        div.appendChild(r);
        pagecontentdiv.appendChild(div);
        content.applyStyle_(content.style);
        return;
    }
    var div = document.createElement("div");
    content.div = div;
    div.className = "content" + (content.cssClass ? " " + content.cssClass : "");
    div.appendChild(document.createTextNode(content.text));
    div.contentEditable = true;
    pagecontentdiv.appendChild(div);
    content.applyStyle_(content.style);
    var editor = new LW.Editor(content, "text", div);
};

Page.prototype.addFollower = function(follower) {
    this.followers.push(follower);
    // Remove from invitations
    for( var i = 0; i < this.invitations.length; i++ ) {
        if (this.invitations[i].id == follower.id) {
            this.invitations.splice(i, 1);
            if (this.isVisible()) {
                var div = document.getElementById("invitee-" + follower.id);
                if (div ) {
                    var sharediv = document.getElementById("share");
                    sharediv.removeChild(div);
                }
            }
            break;
        }
    }
    if (!this.isVisible()) {
        return;
    }
    this.showFollower(follower, false);
};

Page.prototype.getFollower = function(userid) {
    for( var i = 0; i < this.followers.length; i++ ) {
        if (this.followers[i].id == userid) {
            return this.followers[i];
        }
    }
    return null;
};

Page.prototype.addInvitation = function(follower) {
    this.invitations.push(follower);
    if (!this.isVisible()) {
        return;
    }
    this.showFollower(follower, true);
};

Page.prototype.getInvitation = function(userid) {
    for( var i = 0; i < this.invitations.length; i++ ) {
        if (this.invitations[i].id == userid) {
            return this.invitations[i];
        }
    }
    return null;
};

Page.prototype.showFollowers = function() {
    for( var i = 0; i < this.followers.length; i++ ) {
        this.showFollower(this.followers[i], false);
    }
    for( var i = 0; i < this.invitations.length; i++ ) {
        this.showFollower(this.invitations[i], true);
    }
};

Page.prototype.showFollower = function(follower, inviteOnly) {
    if (fullscreen) {
        return;
    }
    var div = document.createElement("div");
    // HACK
    if (!inviteOnly) {
        div.id = "follower-" + follower.id;
        div.className = "friend friendonline";
    } else {
        div.className = "friend friendaway";
        div.id = "invitee-" + follower.id;
    }
    var img = document.createElement("img");
    img.className = "friend-image";
    img.src = "unknown.png";
    div.appendChild(img);
    var div2 = document.createElement("div");
    var span = document.createElement("span");
    span.className = "friend-name";
    span.innerText = follower.name;
    div2.appendChild(span);
    div2.appendChild(document.createElement("br"));
    var span = document.createElement("span");
    span.className = "friend-id";
    span.innerText = follower.id;
    div2.appendChild(span);    
    div.appendChild(div2);
    var sharediv = document.getElementById("share");
    var invitesdiv = document.getElementById("invitations");
    if (inviteOnly) {
        sharediv.appendChild(div);
    } else {
        var sharediv = document.getElementById("share");
        sharediv.insertBefore(div, invitesdiv);
    }
};

Page.prototype.isVisible = function(content) {
    if (this.chapter.currentPage != this) {
        return false;
    }
    if (this.chapter.book.currentChapter != this.chapter) {
        return false;
    }
    return true;
};

Page.prototype.getContent = function(id) {
    for( var i = 0; i < this.contents.length; i++) {
        if (this.contents[i].id == id) {
            return this.contents[i];
        }
    }
    return null;
};

Page.prototype.setUnread = function(unread) {
    if (this.unread == unread) {
        return;
    }
    this.unread = unread;
    if (!this.unread && this.inbox_latestauthors) {
        this.inbox_authors = this.inbox_latestauthors.concat(this.inbox_authors);
        this.inbox_latestauthors = [];
    }
    if (!this.vtab) {
        return;
    }
    var span = this.vtab.lastChild;
    if (!$(span).hasClass("pageunread")) {
        span = null;
    }
    if (this.unread) {
        if (!span) {
            span = document.createElement("span");
            span.className = "pageunread";
            this.vtab.appendChild(span);
        }
        span.innerHTML = "&nbsp";
    } else {
        if (span) {
            this.vtab.removeChild(span);
        }
    }
};

Page.prototype.setLayout = function(layout) {
    this.layout = layout;
    this.applyLayout_();
};

Page.prototype.applyLayout_ = function() {
    if (fullscreen) {
        var pagediv = document.getElementById("pagecontent_fullscreen");
        var pagecontentdiv = document.getElementById("pagecontent_fullscreen_scaled");
        var scale = 1;
        if (this.layout && this.layout.style && this.layout.style["width"]) {
            scale = pagediv.offsetWidth / parseInt(this.layout.style["width"]);
        }
        pagecontentdiv.style["-webkit-transform"] = "scale(" + scale.toString() + ")";
        pagecontentdiv.style.width = (pagediv.offsetWidth / scale).toString() + "px";
        return;
    }

    var pagecontentdiv = document.getElementById("pagecontent");
    var pagediv = document.getElementById("page");
    var scale = 1;
    if (this.layout && this.layout.style && this.layout.style["width"]) {
        scale = (pagediv.offsetWidth - 232) / parseInt(this.layout.style["width"]);
        pagecontentdiv.style["-webkit-transform"] = "scale(" + scale.toString() + ")";
        pagecontentdiv.style.width = parseInt(this.layout.style["width"]).toString() + "px";
    } else {
        pagecontentdiv.style["-webkit-transform"] = "scale(1)";
        pagecontentdiv.style.width = (pagediv.offsetWidth - 232).toString() + "px";
    }
    if (this.layout && this.layout.style && this.layout.style["height"]) {
        pagecontentdiv.style.height = (parseInt(this.layout.style["height"])).toString() + "px";
        pagediv.style.height = (pagecontentdiv.offsetTop + 12 + parseInt(this.layout.style["height"]) * scale).toString() + "px";
    } else {
        pagecontentdiv.style.height = "auto";
        pagediv.style.height = "auto";
    }
};

// Returns a list of chapters which contain the same permanode as this page
Page.prototype.getChapters = function() {
  // Find out in which chapter this page is
  return this.chapter.book.getPages()[this.pageBlobRef];
};

// =================================================================
// PageContent
// =================================================================

PageContent.prototype.rotate = function(rotation) {
    this.style.rotate = rotation;
    this.div.style["-webkit-transform"] = "rotate(" + rotation.toString() + "deg)";
};

PageContent.prototype.submitRotate = function(x, y, w, h, rotation) {
    this.rotate(rotation);
    var op = {"rotate": Math.round(rotation).toString(), "left": Math.round(x).toString(), "top": Math.round(y).toString(), "width": Math.round(w).toString(), "height": Math.round(h).toString()};
    var msg = {perma: this.page.pageBlobRef, op: op, type: "mutation", entity: this.id, field: "style"};
    var pageContent = this;
    store.submit(msg, null, null, function(m) { m.entity = pageContent.id });
};

PageContent.prototype.move = function(x, y) {
    this.style.left = x;
    this.style.top = y;
    this.div.style.left = x.toString() + "px";
    this.div.style.top = y.toString() + "px";
};

PageContent.prototype.submitMove = function(x, y) {
    this.move(x, y);
    var op = {"left": x.toString(), "top": y.toString()};
    var msg = {perma: this.page.pageBlobRef, op: op, type: "mutation", entity: this.id, field: "style"};
    var pageContent = this;
    store.submit(msg, null, null, function(m) { m.entity = pageContent.id });
};

PageContent.prototype.resize = function(w, h) {
    this.style.width = w;
    this.style.height = h;
    this.div.firstChild.style.width = w.toString() + "px";
    this.div.firstChild.style.height = h.toString() + "px";
};

PageContent.prototype.submitResize = function(w, h) {
    this.resize(w, h);
    var op = {"width": w.toString(), "height": h.toString()};
    var msg = {perma: this.page.pageBlobRef, op: op, type: "mutation", entity: this.id, field: "style"};
    var pageContent = this;
    store.submit(msg, null, null, function(m) { m.entity = pageContent.id });
};

PageContent.prototype.applyStyle_ = function(style) {
    var resize = false;
    var move = false;
    var rotate = false;
    var newstyle = { }
    for (var key in this.style) {
        newstyle[key] = this.style[key]
    }
    for (var key in style) {
        try {
            if (key == "width" || key == "height") {
                newstyle[key] = parseInt(style[key]);
                resize = true;
            } else if (key =="left" || key == "top") {
                newstyle[key] = parseInt(style[key]);
                move = true;
            } else if (key == "rotate") {
                newstyle[key] = parseInt(style[key]);
                rotate = true;
            } else {
                console.log("Unsupported property: " + key);
            }
        } catch(e) {
            console.log("Type mismatch in property");
        }
    }
        
    if (this.div) {
        if (resize) {
            this.resize(newstyle["width"], newstyle["height"]);
        }
        if (move) {
            this.move(newstyle["left"], newstyle["top"]);
        }
        if (rotate) {
            this.rotate(newstyle["rotate"]);
        }
    }
};

PageContent.prototype.mutate = function(mutation, islocal) {
    if (mutation.field == "text") {
        lightwave.ot.executeStringOperations(this, mutation.op);
        if (this.cssClass == "title") {
            if ( islocal ) {
                this.page.submitText(this.firstTextLine());
            } else {
                this.page.setText(this.firstTextLine());
            }
        }
    } else if (mutation.field == "style") {
        this.applyStyle_(mutation.op);
    } else {
        console.log("Err: Unknown mutation field: " + mutation.field)
    }
};

PageContent.prototype.firstTextLine = function() {
    if ( !this.paragraphs) {
        var parags = this.text.split("\n");
        return parags[0];
    }
    return this.paragraphs[0].text;
};

PageContent.prototype.buildParagraphs = function() {
    this.paragraphs = [];
    var parags = this.text.split("\n");
    // Start by 1 because the string starts with a newline char.
    for( var i = 1; i < parags.length; i++ ) {
        var p = parags[i];
        this.paragraphs.push({text:p, style:{}});
    }
};

PageContent.prototype.Begin = function() {
    this.tombStream = new lightwave.ot.TombStream(this.tombs);
    this.mut_charCount = 0;
    this.mut_paragIndex = -1;
    this.mut_paragModified = false;
};

PageContent.prototype.InsertChars = function(str, style) {
    if (this.mut_paragIndex == -1) {
        console.log("ERR: Cannot insert at position 0");
        return;
    }
    this.tombStream.InsertChars(str.length);
    var parags = str.split("\n");
    for( var i = 0; i < parags.length; i++ ) {
        var s = parags[i];
        var parag = this.paragraphs[this.mut_paragIndex];
        if ( i > 0 ) {
            this.paragraphs.splice(this.mut_paragIndex + 1, 0, {text:s + parag.text.substring(this.mut_charCount, parag.text.length), style: style ? style : {} });
            parag.text = parag.text.substring(0, this.mut_charCount);
            if ( this.listener ) {
                this.listener.viewRenderParagraph(this.mut_paragIndex);
                this.listener.viewInsertParagraph(this.mut_paragIndex + 1);
            }   
            this.mut_paragIndex++;
            this.mut_charCount = s.length;
            this.mut_paragModified = false;
        } else {
            parag.text = parag.text.substring(0, this.mut_charCount) + s + parag.text.substring(this.mut_charCount, parag.text.length);
            this.mut_charCount += s.length;
            this.mut_paragModified = true;
        }
    }
};

PageContent.prototype.InsertTombs = function(count) {
    this.tombStream.InsertTombs(count);
};

PageContent.prototype.Delete = function(count) {
    if (this.mut_paragIndex == -1) {
        console.log("ERR: Cannot delete at position 0");
        return;
    }
    var burried, err;
    var result = this.tombStream.Bury(count);
    burried = result[0];
    err = result[1];
    if (err) {
        return err;
    }
    while( burried > 0 ) {
        this.mut_paragModified = true;
        var parag = this.paragraphs[this.mut_paragIndex];
        // Delete a line break?
        if (this.mut_charCount == parag.text.length) {
            parag.text = parag.text + this.paragraphs[this.mut_paragIndex + 1].text;
            this.paragraphs.splice(this.mut_paragIndex + 1, 1);
            burried--;
            if (this.listener) {
                this.listener.viewDeleteParagraph(this.mut_paragIndex + 1);
            }
        } else {
            var l = Math.min(burried, parag.text.length - this.mut_charCount);
            burried -= l;
            parag.text = parag.text.substring(0, this.mut_charCount) + parag.text.substring(this.mut_charCount + l, parag.text.length);
        }
        if (this.mut_paragIndex >= this.paragraphs.length) {
            throw "Error in delete";
        }
    }
    return null;
};

PageContent.prototype.Skip = function(count, style) {
    var s = decodeStyle(style);
    var chars = 0, err;
    var result = this.tombStream.Skip(count);
    chars = result[0];
    err = result[1];
    if (err) {
        return err;
    }
    while( chars > 0 ) {
        var len = 0;
        if (this.mut_paragIndex != -1 ) {
            len = this.paragraphs[this.mut_paragIndex].text.length;
        }
        // Skip a line break?
        if (this.mut_charCount == len) {
            if (this.mut_paragModified) {
                if (this.listener) {
                    this.listener.viewRenderParagraph(this.mut_paragIndex);
                }
                this.mut_paragModified = false;
            }
            this.mut_paragIndex++;
            this.mut_charCount = 0;
            chars--;
            if (s) {
                this.paragraphs[this.mut_paragIndex].style = mergeStyles(this.paragraphs[this.mut_paragIndex].style, s);
                this.mut_paragModified = true;
            }
        } else {
            var l = Math.min(chars, len - this.mut_charCount);
            chars -= l;
            this.mut_charCount += l;
        }
        if (this.mut_paragIndex >= this.paragraphs.length) {
            throw "Error in skip";
        }
    }
    return null;
};

PageContent.prototype.End = function() {
    if (this.listener && this.mut_paragModified) {
        this.listener.viewRenderParagraph(this.mut_paragIndex);
    }
    delete this.tombStream;
};

// ===============================================
// Colors
// ===============================================

var colorSchemes = [
    {"color":"#bbbbff"},
    {"color":"#ff8888"},
    {"color":"#99ff99"},
    {"color":"#ffda70"}
];

// ===============================================
// Helper functions
// ===============================================

function decodeStyle(style) {
    if (!style || style == "") {
        return null;
    }
    var result = { };
    var lst = style.split(";");
    for( var i = 0; i < lst.length; i++ ) {
        lst2 = lst[i].split(":");
        if (lst2.length == 2) {
            result[lst2[0]] = lst2[1];
        }
    }
    return result;
}

function mergeStyles(style1, style2) {
    if (!style1) {
        return style2;
    }
    if (!style2) {
        return style1;
    }
    for (var key in style2) {
        style1[key] = style2[key];
    }
    return style1;
}
