/*
  Text = { paragraphs:[ {paragFormat{fontSize:10}, text:"Hallo Welt", format:[{len:6},{len:4, bold:true}] }, ... ] }
*/

if ( !window.LW ) {
    LW = { };
}

/**
 * Implements the data structure for a rich text object that can be modified by OT mutations.
 */
LW.Richtext = function() {
    this.paragraphs = [];
    this["$rtf"] = true;
};

LW.Richtext.ParagRenderAction = 1;
LW.Richtext.ParagDeleteAction = 2;
LW.Richtext.ParagInsertAction = 3;

LW.Richtext.prototype.startUpdate = function() {
    this.updateQueue = [];
};

LW.Richtext.prototype.insertReturn = function(pos, paragFormat) {
    if ( pos.paragIndex == -1 ) {
        var pnew = { paragFormat:paragFormat, text:"", format:[] };
        this.paragraphs.unshift(pnew);
        this.enqueueUpdate_(this.paragraphs.length - 1, LW.Richtext.ParagInsertAction);
    } else if ( pos.paragIndex == this.paragraphs.length ) {
        var pnew = { paragFormat:paragFormat, text:"", format:[] };
        this.paragraphs.push(pnew);
        this.enqueueUpdate_(this.paragraphs.length - 1, LW.Richtext.ParagInsertAction);
    } else {
        var parag = this.paragraphs[pos.paragIndex];
        var pnew = { paragFormat:paragFormat };
        pnew.text = parag.text.substring(pos.charCount, parag.text.length)
        pnew.format = parag.format.slice(pos.charCount, parag.text.length)
        parag.text = parag.text.substr(0, pos.charCount);
        parag.format = parag.format.slice(0, pos.charCount);
        this.paragraphs.splice(pos.paragIndex + 1, 0, pnew );
        this.enqueueUpdate_(pos.paragIndex, LW.Richtext.ParagRenderAction);
        this.enqueueUpdate_(pos.paragIndex + 1, LW.Richtext.ParagInsertAction);
    }
};

LW.Richtext.prototype.deleteRange = function(pos1, pos2) {
    var parag1 = this.paragraphs[pos1.paragIndex];
    if ( pos1.paragIndex == pos2.paragIndex ) {
        parag1.text = parag1.text.substr(0, pos1.charCount) + parag1.text.substring(pos2.charCount, parag1.text.length);
        parag1.format = parag1.format.slice(0, pos1.charCount) + parag1.format.slice(pos2.charCount, parag1.text.length);
    } else {
        var parag2 = this.paragraphs[pos2.paragIndex];
        this.paragraphs.splice(pos1.paragIndex + 1, pos2.paragIndex - pos1.paragIndex);
        parag1.text = parag1.text.substr(0, pos1.charCount) + parag2.text.substring(pos2.charCount, parag2.text.length);
        parag1.format = parag1.format.slice(0, pos1.charCount) + parag2.format.slice(pos2.charCount, parag2.text.length);
        for( var i = pos1.paragIndex + 1; i <= pos2.paragIndex; ++i ) {
            this.enqueueUpdate_(pos1.paragIndex + 1, LW.Richtext.ParagDeleteAction);
        }
    }
    this.enqueueUpdate_(pos1.paragIndex, LW.Richtext.ParagRenderAction);
};

LW.Richtext.prototype.insertText = function(pos, text) {
    var parag = this.paragraphs[pos.paragIndex];
    parag.text = parag.text.substr(0, pos.charCount) + text + parag.text.substring(pos.charCount, parag.text.length);
    var offset = text.length;
    for(var i = parag.text.length - 1; i >= pos.charCount; --i ) {
        if ( parag.format[i] ) {
            parag.format[i+offset] = format[i];
            delete format[i];
        }
    }
    this.enqueueUpdate_(pos.paragIndex, LW.Richtext.ParagRenderAction);
};

LW.Richtext.prototype.insertObject = function(pos, obj) {
    var parag = this.paragraphs[pos.paragIndex];
    parag.text = parag.text.substr(0, pos.charCount) + "x" + parag.text.substring(pos.charCount, parag.text.length);
    parag.format.splice(pos.charCount, 0, obj);
    this.enqueueUpdate_(pos.paragIndex, LW.Richtext.ParagRenderAction);
};

LW.Richtext.prototype.enqueueUpdate_ = function(paragIndex, action) {
    var len = this.updateQueue.length;
    if ( action == LW.Richtext.ParagRenderAction && len > 0 && this.updateQueue[len-1].paragIndex == paragIndex &&
         (this.updateQueue[len-1].action == LW.Richtext.ParagRenderAction || this.updateQueue[len-1].action == LW.Richtext.ParagInsertAction ) ) {
        return;
    }
    this.updateQueue.push({paragIndex:paragIndex, action:action});
};

LW.Richtext.prototype.finishUpdate = function() {
    if ( !this.view ) {
        return;
    }
    for( var i = 0; i < this.updateQueue.length; ++i ) {
        var u = this.updateQueue[i];
        if ( u.action == LW.Richtext.ParagDeleteAction ) {
            this.view.viewDeleteParagraph( u.paragIndex );
        } else if ( u.action == LW.Richtext.ParagRenderAction ) {
            this.view.viewRenderParagraph( u.paragIndex );
        } else if ( u.action == LW.Richtext.ParagInsertAction ) {
            this.view.viewInsertParagraph( u.paragIndex );
        }
    }
    console.log(this.paragraphs);
};
