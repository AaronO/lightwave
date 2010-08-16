#include "wavecontainer.h"
#include "waverootdocument.h"
#include "waveprovider.h"
#include "hostcontainer.h"
#include "json/jsonscanner.h"
#include "ot/abstractmutation.h"
#include "session.h"
#include "utils/settings.h"

#include <QNetworkAccessManager>
#include <QNetworkRequest>
#include <QUrl>
#include <string>

QNetworkAccessManager* WaveContainer::s_networkManager = 0;

WaveContainer::WaveContainer(WaveContainer* parent, const QString& name)
    : QObject(parent), m_name(name), m_isTemp(true)
{
    m_documents.insert("_meta", new WaveMetaDocument(this, "_meta"));
}

WaveContainer::~WaveContainer()
{
    foreach( WaveDocument* doc, m_documents.values())
        delete doc;
    foreach( WaveContainer* c, m_children.values())
        delete c;
}

JSONObject WaveContainer::get(FCGI::FCGIRequest* req, const QString& docKind)
{
    Q_ASSERT(docKind.startsWith("_"));
    Q_UNUSED(req);

    WaveDocument* wdoc = m_documents.value(docKind);
    if ( !wdoc )
    {
        JSONObject e(true);
        e.setAttribute("ok", false);
        e.setAttribute("error", "Document does not exist");
        return e;
    }

    return wdoc->jsonObject();
}

JSONObject WaveContainer::put(JSONObject doc, const QString& docKind, FCGI::FCGIRequest* req)
{
    qDebug("Put document %s", qPrintable(docKind));
    Q_ASSERT(docKind.startsWith("_"));

    if ( isRemote() )
    {
        Q_ASSERT(req);
        qDebug("Put document remote");
        new SubmitToHostJob(this, req, docKind, doc);
        return JSONObject();
    }

    //
    // Check the version, mutate the submitted document and persist it
    //

    WaveDocument* wdoc = m_documents.value(docKind);

    bool is_initial = false;
    // Initial submission?
    if ( doc.attributeString("_rev").isEmpty())
    {
        is_initial = true;

        if ( wdoc && ( wdoc != metaDocument() || wdoc->revisionNumber() > 0 ) )
        {
            JSONObject e(true);
            e.setAttribute("ok", false);
            e.setAttribute("error", "Document already exists");
            return e;
        }
        // Create the document
        if ( !wdoc )
            wdoc = new WaveDocument(this, docKind);

        // Apply the initial mutation
        AbstractMutation m(doc);
        DocumentMutation docop( m );
        if ( !wdoc->processMutation(docop))
        {
            JSONObject e(true);
            e.setAttribute("ok", false);
            e.setAttribute("error", "Failure when processing mutation");
            return e;
        }

        // Add the document to the wave
        if ( !metaDocument()->addDocument(wdoc) )
        {
            JSONObject e(true);
            e.setAttribute("ok", false);
            e.setAttribute("error", "Failed to add document to container");
            return e;
        }

        m_documents[docKind] = wdoc;
    }
    // Overwrite/mutate the document?
    else
    {
        if ( !wdoc )
        {
            JSONObject e(true);
            e.setAttribute("ok", false);
            e.setAttribute("error", "Document does not exist");
            return e;
        }

        AbstractMutation m(doc);
        DocumentMutation docop( m );
        if ( !wdoc->processMutation(docop))
        {
            JSONObject e(true);
            e.setAttribute("ok", false);
            e.setAttribute("error", "Failure when processing mutation");
            return e;
        }
    }

    // Tell all sessions which are listing that the wave has changed
    notifySessions(wdoc, is_initial);

    //
    // Send a message to all remote hosts
    //
    if ( !m_remoteHosts.isEmpty() )
    {
        JSONObject msg(true);
        JSONObject m = doc.clone().toObject();
        if ( is_initial )
            m.setAttribute("_snapshot", true);
        msg.setAttribute(wdoc->waveId().toString(), m);
        foreach(QString h, m_remoteHosts)
        {
            new SubmitToRemoteJob(this, h, msg.toJSON().toUtf8());
        }
    }

    if ( docKind == "_meta" )
        updateFromMetaDocument();

    onDocumentUpdate(wdoc);

    JSONObject obj(true);
    obj.setAttribute("ok", true);
    obj.setAttribute("id", wdoc->waveId().toString());
    obj.setAttribute("rev", wdoc->revision());
    return obj;
}

JSONObject WaveContainer::putFromHost(JSONObject doc, const QString& docKind)
{
    Q_ASSERT(isRemote());
    Q_ASSERT(docKind.startsWith("_"));

    WaveDocument* wdoc = m_documents.value(docKind);
    if ( !wdoc )
    {
        JSONObject e(true);
        e.setAttribute("ok", false);
        e.setAttribute("error", "Document does not exist");
        return e;
    }

    // Apply the delta
    AbstractMutation m(doc);
    DocumentMutation docop( m );

    if ( !wdoc->processMutationFromHost(docop))
    {
        JSONObject e(true);
        e.setAttribute("ok", false);
        e.setAttribute("error", "Failure when processing mutation");
        return e;
    }

    // Tell all sessions which are listing that the wave has changed
    notifySessions(wdoc, false);

    JSONObject obj(true);
    obj.setAttribute("ok", true);
    return obj;
}

JSONObject WaveContainer::putSnapshotFromHost( JSONObject doc, const QString& docKind )
{
    Q_ASSERT(docKind.startsWith("_"));
    Q_ASSERT(isRemote());

    WaveDocument* wdoc;
    if ( docKind == "_meta")
        wdoc = metaDocument();
    else
    {
        if ( m_documents.contains(docKind))
        {
            qDebug("Error: Cannot use snapshot because document already exists");
            JSONObject e(true);
            e.setAttribute("ok", false);
            e.setAttribute("error", "Cannot use snapshot because document already exists");
            return e;
        }

        // Create the document
        wdoc = new WaveDocument(this, docKind);
        m_documents[docKind] = wdoc;
    }
    wdoc->setSnapshotFromHost(doc);

    // Tell all sessions which are listing that the wave has changed
    notifySessions(wdoc, true);

    JSONObject obj(true);
    obj.setAttribute("ok", true);
    return obj;
}

JSONObject WaveContainer::putFromRemote( JSONObject doc, const QString& docKind )
{
    Q_ASSERT(docKind.startsWith("_"));
    Q_ASSERT(!isRemote());

    // Check signature
    // TODO

    // Check permissions, i.e. is this remote server perhaps blocked?
    // TODO

    // qDebug("Message from remote server: %s", req->m_stdinStream.data());
    return put(doc, docKind);
}

void WaveContainer::updateFromMetaDocument()
{
    WaveMetaDocument* metaDoc = metaDocument();

    //
    // Check for changes in the root document
    //

    QSet<QString> authors = metaDoc->jsonObject().attributeObject("authors").attributeNamesSet();
    QSet<QString> new_authors = authors.subtract(m_authors);
    QSet<QString> removed_authors = m_authors.subtract(authors);

    foreach( QString a, new_authors )
    {
        JID jid(a);
        // Malformed
        if ( !jid.isValid())
            continue;
        m_authors.insert(a);
        // TODO: Inform the sessions of the author
    }

    foreach( QString a, removed_authors )
    {
        // TODO: Inform the sessions of the author
    }

    // Determine the set of remote hosts
    QSet<QString> remoteHosts;
    foreach(QString a, m_authors)
    {
        JID jid(a);
        if ( jid.domain() == Settings::settings()->domain())
            continue;
        remoteHosts.insert(jid.domain());
    }

    // Send a snapshot to the new the remote hosts
    foreach(QString h, remoteHosts)
    {
        if ( !m_remoteHosts.contains(h))
        {
            JSONObject obj = snapshot();
            qDebug("ADDING remote host %s. Sending snapshot %s", qPrintable(h), qPrintable(obj.toJSON()));
            // Send a snapshot to the remote host
            new SubmitToRemoteJob(this, h, obj.toJSON().toUtf8());
        }
    }

    m_remoteHosts = remoteHosts;

}

JSONObject WaveContainer::snapshot()
{
    JSONObject obj(true);
    snapshot(obj);
    return obj;
}

void WaveContainer::snapshot(JSONObject obj)
{
    foreach( QString id, m_documents.keys())
    {
        WaveDocument* wdoc = m_documents[id];
        JSONObject doc = wdoc->jsonObject().clone().toObject();
        doc.setAttribute("_snapshot", true);
        obj.setAttribute(wdoc->waveId().toString(), doc);
    }

    foreach( WaveContainer* c, m_children.values())
    {
        c->snapshot(obj);
    }
}

WaveContainer* WaveContainer::createWaveContainer(const QString& name)
{
    Q_ASSERT(childContainer(name) == 0);
    WaveContainer* c = new WaveContainer(this, name);
    return c;
}

void WaveContainer::addContainer( WaveContainer* child )
{
    m_children.insert( child->name(), child );

    if ( !isRemote())
    {
        // Add the document to the wave
        if ( !metaDocument()->addContainer(child) )
            qDebug("Failed to add document to container");
    }
}

void WaveContainer::makePersistent()
{
    if ( !m_isTemp )
        return;

    parentContainer()->addContainer(this);
    m_isTemp = false;
}

void WaveContainer::registerSession( const QString& sessionId )
{
    // Get a pointer to the session
    Session *s = WaveProvider::self()->session(sessionId);
    if ( !s )
    {
        qDebug("Oooops, new session is already dead: %s", qPrintable(sessionId));
        return;
    }
    m_sessions.insert(sessionId);

    // Tell the session about all document versions in this wave
    QHash<QString,QString> revisions;
    foreach( WaveDocument* doc, m_documents.values() )
    {
        revisions[doc->waveId().toString()] = doc->revision();
    }
    s->notify(revisions);
}

void WaveContainer::deregisterSession( const QString& sessionId )
{
    m_sessions.remove(sessionId);
}

void WaveContainer::notifySessions(WaveDocument* doc, bool sendMetaDoc)
{
    QHash<QString,QString> revisions;
    revisions[doc->waveId().toString()] = doc->revision();
    if ( sendMetaDoc )
        revisions[metaDocument()->documentId()] = metaDocument()->revision();

    foreach( QString sid, m_sessions )
    {
        // Get a pointer to the session
        Session *s = WaveProvider::self()->session(sid);
        if ( !s )
        {
            qDebug("Session has died");
            m_sessions.remove(sid);
            continue;
        }
        s->notify(revisions);
    }
}

QList<DocumentMutation> WaveContainer::getMutations( const QString& docId, const QString& sinceRevision )
{
    WaveDocument* wdoc = m_documents.value(docId);
    if ( !wdoc )
        return QList<DocumentMutation>();
    return wdoc->getMutations(sinceRevision);
}

HostContainer* WaveContainer::hostContainer() const
{
    const WaveContainer* c = this;
    while(c)
    {
        if ( dynamic_cast<const HostContainer*>(c) )
            return const_cast<HostContainer*>(dynamic_cast<const HostContainer*>(c));
        c = c->parentContainer();
    }
    return 0;
}

bool WaveContainer::isRemote() const
{
    HostContainer* h = hostContainer();
    Q_ASSERT(h);
    return !h->isLocal();
}

QNetworkAccessManager* WaveContainer::networkManager()
{
    if ( !s_networkManager )
        s_networkManager = new QNetworkAccessManager();
    return s_networkManager;
}

void WaveContainer::onDocumentUpdate(WaveDocument* wdoc)
{
    Q_UNUSED(wdoc)
}

////////////////////////////////////////////////////////
//
// SubmitToHostJob
//
////////////////////////////////////////////////////////

SubmitToHostJob::SubmitToHostJob(WaveContainer* parent, FCGI::FCGIRequest* req, const WaveId& waveId, JSONObject data)
    : QObject(parent), m_clientRequest(req)
{
    QUrl url("http://" + parent->hostContainer()->name() + "/wave/_host");
    qDebug("Sending to host: %s", qPrintable(url.toString()));

    JSONObject msg(true);
    msg.setAttribute( waveId.toString(), data);
    m_data = msg.toJSON().toUtf8();

    QNetworkRequest serverRequest( url );
    m_serverReply = WaveContainer::networkManager()->put(serverRequest, m_data);

    bool ok = QObject::connect(m_serverReply, SIGNAL(finished()), this, SLOT(onFinished()));
    Q_ASSERT(ok);
    ok = QObject::connect(m_serverReply, SIGNAL(error(QNetworkReply::NetworkError)), this, SLOT(onError(QNetworkReply::NetworkError)));
    Q_ASSERT(ok);
    ok = QObject::connect(m_serverReply, SIGNAL(sslErrors(QList<QSslError>)), this, SLOT(onSslErrors(QList<QSslError>)));
    Q_ASSERT(ok);
}

SubmitToHostJob::~SubmitToHostJob()
{
    if ( m_serverReply )
    {
        m_serverReply->abort();
        m_serverReply->deleteLater();
    }
}

void SubmitToHostJob::onError(QNetworkReply::NetworkError code)
{
    Q_UNUSED(code);
    sendErrorToClient();
    deleteLater();
}

void SubmitToHostJob::onFinished()
{
    if ( m_serverReply )
    {
        QByteArray data = m_serverReply->readAll();
        JSONScanner scanner(data.constData(), data.count());
        bool ok = false;
        JSONObject doc = scanner.scan(&ok);
        if ( !ok )
            qDebug("Failed parsing the response from the hosting server: %s", qPrintable(doc.toJSON()));
        else
            qDebug("Answer from hosting server: %s", qPrintable(doc.toJSON()));

        if ( doc.attribute("ok").toBool() == true )
        {
            if ( m_clientRequest )
                m_clientRequest->replyJson(data);
        }
        else
        {
            if ( m_clientRequest )
                m_clientRequest->errorReply( doc.attributeString("error"));
        }
        m_clientRequest = 0;
    }    
    deleteLater();
}

void SubmitToHostJob::onSslErrors( const QList<QSslError>& errors )
{
    Q_UNUSED(errors)
    sendErrorToClient();
    deleteLater();
}

void SubmitToHostJob::sendErrorToClient()
{
    if ( m_clientRequest )
    {
        m_clientRequest->errorReply("Communication with hosting server failed");
        m_clientRequest = 0;
    }
}

///////////////////////////////////////////////////////
//
// SubmitToRemoteJob
//
///////////////////////////////////////////////////////

SubmitToRemoteJob::SubmitToRemoteJob(WaveContainer* parent, const QString& host, const QByteArray& data )
    : QObject(parent), m_data(data)
{
    Q_ASSERT(!parent->isRemote());
    QUrl url("http://" + host + "/wave/_remote");
    QNetworkRequest serverRequest( url );
    m_serverReply = WaveContainer::networkManager()->put(serverRequest, m_data);
    qDebug("Sending to remote: %s", qPrintable(url.toString()));

    bool ok = QObject::connect(m_serverReply, SIGNAL(finished()), this, SLOT(onFinished()));
    Q_ASSERT(ok);
    ok = QObject::connect(m_serverReply, SIGNAL(error(QNetworkReply::NetworkError)), this, SLOT(onError(QNetworkReply::NetworkError)));
    Q_ASSERT(ok);
    ok = QObject::connect(m_serverReply, SIGNAL(sslErrors(QList<QSslError>)), this, SLOT(onSslErrors(QList<QSslError>)));
    Q_ASSERT(ok);
}

SubmitToRemoteJob::~SubmitToRemoteJob()
{
    if ( m_serverReply )
    {
        m_serverReply->abort();
        m_serverReply->deleteLater();
    }
}

void SubmitToRemoteJob::onError(QNetworkReply::NetworkError code)
{
    Q_UNUSED(code);
    deleteLater();
}

void SubmitToRemoteJob::onFinished()
{
    if ( m_serverReply )
    {
        QByteArray data = m_serverReply->readAll();
        JSONScanner scanner(data.constData(), data.count());
        bool ok = false;
        JSONObject doc = scanner.scan(&ok);
        if ( !ok )
            qDebug("Failed parsing the response from the remote server: %s", qPrintable(doc.toJSON()));
        else
            qDebug("Answer from remote server: %s", qPrintable(doc.toJSON()));
        // TODO make any use from the reply
    }

    deleteLater();
}

void SubmitToRemoteJob::onSslErrors( const QList<QSslError>& errors )
{
    Q_UNUSED(errors)
    deleteLater();
}
