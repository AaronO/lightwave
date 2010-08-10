#include "waverootdocument.h"
#include "json/jsonarray.h"
#include "ot/objectmutation.h"
#include "ot/insertmutation.h"
#include "ot/documentmutation.h"
#include "wavecontainer.h"
#include <QUrl>
#include <QNetworkRequest>

WaveRootDocument::WaveRootDocument(WaveContainer* container, const QString& docId)
    : WaveDocument(docId), m_container(container)
{
}

bool WaveRootDocument::addDocument(FCGI::FCGIRequest* req, WaveDocument* wdoc, bool suppressReply)
{
    ObjectMutation rootop(true);
    if ( !jsonObject().hasAttribute("documents"))
    {
        JSONObject documentsOp(true);
        documentsOp.setAttribute(wdoc->docId().mid( docId().length() + 1), JSONObject(true));
        rootop.setMutation("documents", documentsOp);
    }
    else
    {
        ObjectMutation documentsOp(true);
        documentsOp.setMutation(wdoc->docId().mid( docId().length() + 1), InsertMutation(JSONObject(true)));
        rootop.setMutation("documents", documentsOp);
    }
    DocumentMutation docop(rootop);
    docop.setDocumentId(docId());
    docop.setRevision(revision());

    qDebug("OP=%s", qPrintable(docop.mutation().toJSON()));

    return processMutation(req, docop, suppressReply);
}

void WaveRootDocument::update()
{
    if ( m_container->isRemote() )
    {
        m_container->m_authors = jsonObject().attributeObject("authors").attributeNamesSet();
        return;
    }

    QSet<QString> authors = jsonObject().attributeObject("authors").attributeNamesSet();
    QSet<QString> new_authors = authors.subtract(m_container->m_authors);
    QSet<QString> removed_authors = m_container->m_authors.subtract(authors);

    foreach( QString a, new_authors )
    {
        JID jid(a);
        // Malformed
        if ( !jid.isValid())
            continue;
        m_container->m_authors.insert(a);
        // TODO: Inform the sessions of the author
    }

    foreach( QString a, removed_authors )
    {
        // TODO: Inform the sessions of the author
    }

    QSet<QString> remoteHosts;
    foreach(QString a, m_container->m_authors)
    {
        JID jid(a);
        if ( jid.domain() == m_container->host())
            continue;
        remoteHosts.insert(jid.domain());
    }
    // Any new remote hosts?
    foreach(QString h, remoteHosts.subtract(m_container->m_remoteHosts))
    {
        qDebug("ADDING remote host %s", qPrintable(h));
        // Send a snapshot to the remote host
        JSONObject obj(true);
        obj.setAttribute(docId(), jsonObject().clone());
        foreach( QString id, m_container->m_documents.keys())
        {
            WaveDocument* wdoc = m_container->m_documents[id];
            obj.setAttribute(id, wdoc->jsonObject().clone());
        }

        new SubmitToRemoteJob(m_container, h, obj.toJSON().toUtf8());
    }
}

