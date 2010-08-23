#include "waverootdocument.h"
#include "json/jsonarray.h"
#include "ot/objectmutation.h"
#include "ot/insertmutation.h"
#include "ot/documentmutation.h"
#include "wavecontainer.h"
#include <QUrl>
#include <QNetworkRequest>

WaveMetaDocument::WaveMetaDocument(WaveContainer* container, const QString& docId)
    : WaveDocument(container, docId)
{
}

bool WaveMetaDocument::addDocument(WaveDocument* wdoc)
{
    // Strip the "_" in front.
    QString docid = wdoc->documentId().mid(1);

    ObjectMutation rootop(true);
    if ( !jsonObject().hasAttribute("documents"))
    {
        JSONObject documentsOp(true);
        documentsOp.setAttribute(docid, JSONObject(true));
        rootop.setMutation("documents", documentsOp);
    }
    else
    {
        ObjectMutation documentsOp(true);
        documentsOp.setMutation(docid, InsertMutation(JSONObject(true)));
        rootop.setMutation("documents", documentsOp);
    }
    DocumentMutation docop(rootop);
    docop.setDocumentId(documentId());
    docop.setRevision(revision());

    // qDebug("OP=%s", qPrintable(docop.mutation().toJSON()));

    JSONObject result = container()->put(docop.mutation().toObject(), documentId());
    return result.attribute("ok").toBool();
}

bool WaveMetaDocument::addContainer(WaveContainer* c)
{
    Q_ASSERT(container() == c->parentContainer());

    ObjectMutation rootop(true);
    if ( !jsonObject().hasAttribute("containers"))
    {
        JSONObject documentsOp(true);
        documentsOp.setAttribute(c->name(), JSONObject(true));
        rootop.setMutation("containers", documentsOp);
    }
    else
    {
        ObjectMutation documentsOp(true);
        documentsOp.setMutation(c->name(), InsertMutation(JSONObject(true)));
        rootop.setMutation("containers", documentsOp);
    }
    DocumentMutation docop(rootop);
    docop.setDocumentId(documentId());
    docop.setRevision(revision());

    // qDebug("OP=%s", qPrintable(docop.mutation().toJSON()));

    JSONObject result = container()->put(docop.mutation().toObject(), documentId());
    return result.attribute("ok").toBool();
}
