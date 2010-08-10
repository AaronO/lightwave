#include "wavecontainer.h"
#include "waverootdocument.h"
#include "waveprovider.h"
#include "json/jsonscanner.h"
#include "ot/abstractmutation.h"
#include "session.h"
#include "utils/settings.h"

#include <QNetworkAccessManager>
#include <QNetworkRequest>
#include <QUrl>
#include <string>

QNetworkAccessManager* WaveContainer::s_networkManager = 0;

WaveContainer::WaveContainer(const QString& host, const QString& waveId)
    : m_host(host), m_waveId(waveId)
{
    Q_ASSERT(!host.isEmpty());
    Q_ASSERT(waveId.length() > 2 && waveId.startsWith("w+"));

    m_rootDoc = new WaveRootDocument(this, host + "/" + waveId);
}

WaveContainer::~WaveContainer()
{
    delete m_rootDoc;
}

bool WaveContainer::putRootDocumentFromHost( FCGI::FCGIRequest* req, JSONObject doc )
{
    Q_ASSERT(isRemote());

    // Apply the delta
    AbstractMutation m(doc);
    DocumentMutation docop( m );
    if ( !m_rootDoc->processMutation(req, docop, true))
        return false;

    // Tell all sessions which are listing that the wave has changed
    notifySessions(m_rootDoc, false);

    return true;
}

bool WaveContainer::putRootDocument( FCGI::FCGIRequest* req )
{
    JSONObject doc(true);
    // No document has been sent? Then create the default document
    if ( req->m_stdinStream.size() == 0 )
    {
        doc.setAttribute("authors", JSONObject(true));
        doc.setAttribute("documents", JSONObject(true));
    }
    else
    {
        JSONScanner scanner(req->m_stdinStream.data(), req->m_stdinStream.size());
        bool ok = false;
        doc = scanner.scan(&ok);
        if ( !ok )
        {
            req->errorReply("JSON parsing error");
            return false;
        }
    }

    if ( isRemote() )
    {
        new SubmitToHostJob(this, req, QString::null, QByteArray(req->m_stdinStream.data(), req->m_stdinStream.size()));
        return true;
    }

    // Apply the delta
    AbstractMutation m(doc);
    DocumentMutation docop( m );
    if ( !m_rootDoc->processMutation(req, docop))
        return false;

    // Send a reply to the client

    JSONObject obj(true);
    obj.setAttribute("ok", true);
    obj.setAttribute("id", m_rootDoc->docId());
    obj.setAttribute("rev", m_rootDoc->revision());
    req->replyJson(obj.toJSON());

    // Tell all sessions which are listing that the wave has changed
    notifySessions(m_rootDoc, false);

    return true;
}

bool WaveContainer::putDocument( FCGI::FCGIRequest* req, const QString& docId )
{
    Q_ASSERT(docId.startsWith("d+"));

    JSONScanner scanner(req->m_stdinStream.data(), req->m_stdinStream.size());
    bool ok = false;
    JSONObject doc = scanner.scan(&ok);
    if ( !ok )
    {
        req->errorReply("JSON parsing error");
        return false;
    }

    if ( isRemote() )
    {
        new SubmitToHostJob(this, req, docId, QByteArray(req->m_stdinStream.data(), req->m_stdinStream.size()));
        return true;
    }

    //
    // Check the version, mutate the submitted document and persist it
    //

    WaveDocument* wdoc = m_documents.value(docId);

    bool is_initial = false;
    // Initial submission?
    if ( doc.attributeString("_rev").isEmpty())
    {
        is_initial = true;

        if ( wdoc )
        {
            req->errorReply("Error: Document already exists");
            return false;
        }
        // Create the document
        wdoc = new WaveDocument(m_host + "/" + m_waveId + "/" + docId);
        // if ( !wdoc->setContent(req, doc) )

        // Apply the initial mutation
        AbstractMutation m(doc);
        DocumentMutation docop( m );
        if ( !wdoc->processMutation(req, docop))
            return false;

        //
        // Add the document to the wave
        //

        if ( !m_rootDoc->addDocument(req, wdoc) )
            return false;

        m_documents[docId] = wdoc;
    }
    // Overwrite/mutate the document?
    else
    {
        if ( !wdoc )
        {
            req->errorReply("Error: Document does not exist");
            return false;
        }

        AbstractMutation m(doc);
        DocumentMutation docop( m );
        if ( !wdoc->processMutation(req, docop))
            return false;
    }

    //
    // Send a reply to the client
    //

    JSONObject obj(true);
    obj.setAttribute("ok", true);
    obj.setAttribute("id", wdoc->docId());
    obj.setAttribute("rev", wdoc->revision());
    req->replyJson(obj.toJSON());

    // Tell all sessions which are listing that the wave has changed
    notifySessions(wdoc, is_initial);

    return true;
}

bool WaveContainer::putDocumentFromHost( FCGI::FCGIRequest* req, const QString& docId, JSONObject doc )
{
    Q_ASSERT(docId.startsWith("d+"));
    Q_ASSERT(isRemote());

    //
    // Check the version, mutate the submitted document and persist it
    //

    WaveDocument* wdoc = m_documents.value(docId);

    bool is_initial = false;
    // Initial submission?
    if ( doc.attributeString("_rev").isEmpty())
    {
        is_initial = true;

        if ( wdoc )
            return false;

        // Create the document
        wdoc = new WaveDocument(m_host + "/" + m_waveId + "/" + docId);

        // Apply the initial mutation
        AbstractMutation m(doc);
        DocumentMutation docop( m );
        if ( !wdoc->processMutation(req, docop, true))
            return false;

        // Add the document to the wave
        if ( !m_rootDoc->addDocument(req, wdoc, true) )
            return false;

        m_documents[docId] = wdoc;
    }
    // Overwrite/mutate the document?
    else
    {
        if ( !wdoc )
            return false;

        AbstractMutation m(doc);
        DocumentMutation docop( m );
        if ( !wdoc->processMutation(req, docop, true))
            return false;
    }

    // Tell all sessions which are listing that the wave has changed
    notifySessions(wdoc, is_initial);

    return true;
}

bool WaveContainer::putDocumentFromRemote( FCGI::FCGIRequest* req, const QString& docId )
{
    Q_ASSERT(!isRemote());

    WaveDocument* wdoc;
    if ( docId.isEmpty() )
        wdoc = m_rootDoc;
    else
        wdoc = m_documents.value(docId);

    // Check signature
    // TODO

    // Check permissions, i.e. is this remote server perhaps blocked?
    // TODO

    return putDocument(req, docId);
}

void WaveContainer::getDocument( FCGI::FCGIRequest* req, const QString& docId )
{
    WaveDocument* wdoc = m_documents[docId];
    if ( !wdoc )
    {
        req->errorReply("Error: Document does not exist");
        return;
    }

    req->replyJson(wdoc->jsonObject().toJSON());
}

void WaveContainer::getRootDocument( FCGI::FCGIRequest* req )
{
    req->replyJson(m_rootDoc->jsonObject().toJSON());
}

void WaveContainer::registerSession( const QString& sessionId )
{
    // Get a pointer to the session
    Session *s = WaveProvider::self()->session(sessionId);
    if ( !s )
    {
        qDebug("Oooops, new session is already dead");
        return;
    }
    m_sessions.insert(sessionId);

    // Tell the session about all document versions in this wave
    QHash<QString,QString> revisions;
    foreach( WaveDocument* doc, m_documents.values() )
    {
        revisions[doc->docId()] = doc->revision();
    }
    revisions[m_rootDoc->docId()] = m_rootDoc->revision();
    s->notify(revisions);
}

void WaveContainer::deregisterSession( const QString& sessionId )
{
    m_sessions.remove(sessionId);
}

void WaveContainer::notifySessions(WaveDocument* doc, bool sendRootDoc)
{
    QHash<QString,QString> revisions;
    revisions[doc->docId()] = doc->revision();
    if ( sendRootDoc )
        revisions[m_rootDoc->docId()] = m_rootDoc->revision();

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
    if ( docId == m_rootDoc->docId() )
        return m_rootDoc->getMutations(sinceRevision);

    // Strip off the host and waveId part of the ID
    QString d = docId.mid(m_host.length() + 1 + m_waveId.length() + 1);

    WaveDocument* wdoc = m_documents[d];
    if ( !wdoc )
        return QList<DocumentMutation>();
    return wdoc->getMutations(sinceRevision);
}

bool WaveContainer::isRemote() const
{
    return m_host != Settings::settings()->domain();
}

QNetworkAccessManager* WaveContainer::networkManager()
{
    if ( !s_networkManager )
        s_networkManager = new QNetworkAccessManager();
    return s_networkManager;
}

////////////////////////////////////////////////////////
//
// SubmitToHostJob
//
////////////////////////////////////////////////////////

SubmitToHostJob::SubmitToHostJob(WaveContainer* parent, FCGI::FCGIRequest* req, const QString& docId, const QByteArray& data)
    : QObject(parent), m_clientRequest(req), m_docId(docId), m_data(data)
{
    QUrl url;
    if ( docId.isEmpty() )
        url = QUrl("http://" + parent->host() + "/wave/_host/" + parent->waveId() + "/" + docId);
    else
        url = QUrl("http://" + parent->host() + "/wave/_host/" + parent->waveId());
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
    if ( m_clientRequest )
    {
        m_clientRequest->errorReply("Communication with hosting server failed");
        m_clientRequest = 0;
    }
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
            qDebug("Failed parsing the response from the remote server");
        else
            qDebug("Answer from hosting server: %s", qPrintable(doc.toJSON()));
        // TODO make any use from the reply
    }
    deleteLater();
}

void SubmitToHostJob::onSslErrors( const QList<QSslError>& errors )
{
    Q_UNUSED(errors)
    if ( m_clientRequest )
    {
        m_clientRequest->errorReply("Communication with hosting server failed");
        m_clientRequest = 0;
    }
    deleteLater();
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
    QUrl url("http://" + host + "/wave/_remote/" + parent->host());
    QNetworkRequest serverRequest( url );
    m_serverReply = WaveContainer::networkManager()->put(serverRequest, m_data);

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
            qDebug("Failed parsing the response from the remote server");
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
