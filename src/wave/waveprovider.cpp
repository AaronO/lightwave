#include "waveprovider.h"
#include "wavecontainer.h"
#include "utils/settings.h"
#include "session.h"
#include "json/jsonobject.h"
#include "json/jsonscanner.h"

WaveProvider* WaveProvider::s_self = 0;

WaveProvider::WaveProvider()
    : m_waveUriRegExp("([A-Za-z0-9.-]+/)?(w\\+[A-Za-z0-9]+)"), m_docUriRegExp("([A-Za-z0-9.-]+/)?(w\\+[A-Za-z0-9]+)/(d\\+[A-Za-z0-9]+)"), m_sessionUriRegExp("(s\\+[A-Za-z0-9]+)"),
      m_sessionEventsUriRegExp("(s\\+[A-Za-z0-9]+)/_events"), m_sessionDeltasUriRegExp("(s\\+[A-Za-z0-9]+)/_deltas"),
      m_hostWaveUriRegExp("_host/(w\\+[A-Za-z0-9]+)"), m_hostDocUriRegExp("_host/(w\\+[A-Za-z0-9]+)/(d\\+[A-Za-z0-9]+)"), m_remoteWaveUriRegExp("_remote/([A-Za-z0-9.-]+)")
{
    m_waveUriRegExp.setPatternSyntax(QRegExp::RegExp2);
    m_docUriRegExp.setPatternSyntax(QRegExp::RegExp2);
    m_sessionUriRegExp.setPatternSyntax(QRegExp::RegExp2);
    m_sessionEventsUriRegExp.setPatternSyntax(QRegExp::RegExp2);
    m_sessionDeltasUriRegExp.setPatternSyntax(QRegExp::RegExp2);
    m_hostWaveUriRegExp.setPatternSyntax(QRegExp::RegExp2);
    m_hostDocUriRegExp.setPatternSyntax(QRegExp::RegExp2);
    m_remoteWaveUriRegExp.setPatternSyntax(QRegExp::RegExp2);
}

WaveProvider* WaveProvider::self()
{
    if ( !s_self )
        s_self = new WaveProvider();
    return s_self;
}

WaveContainer* WaveProvider::waveContainer(const QString& host, const QString& waveId)
{
    QString h = host;
    if ( h.isEmpty() )
        h = Settings::settings()->domain();

    QString id = h + "/" + waveId;

    if ( !m_container.contains(id))
        return 0;
    return m_container[id];
}

WaveContainer* WaveProvider::createWaveContainer(const QString& host, const QString& waveId, bool asRemoteWave)
{
    QString h = host;
    if ( h.isEmpty() )
        h = Settings::settings()->domain();
    else if ( !asRemoteWave && h != Settings::settings()->domain() )
    {
        qDebug("Cannot create wave on remote server");
        return 0;
    }

    QString id = h + "/" + waveId;

    if ( m_container.contains(id))
    {
        qDebug("Wave already exists");
        return 0;
    }

    WaveContainer* container = new WaveContainer(h, waveId);
    m_container[id] = container;
    return container;
}

Session* WaveProvider::createSession(FCGI::FCGIRequest* req, const QString& sessionId)
{
    if ( m_sessions.contains(sessionId))
    {
        req->errorReply("Session with this ID already exists");
        return 0;
    }

    Session* session = new Session(sessionId);
    m_sessions[sessionId] = session;
    return session;
}

Session* WaveProvider::session(const QString& sessionId)
{
    return m_sessions[sessionId];
}

void WaveProvider::put(FCGI::FCGIRequest* req)
{
    // http://host/wave/w+123
    if ( m_waveUriRegExp.exactMatch(req->requestUri()) )
    {
        // Find or create the wave
        WaveContainer* c = waveContainer(m_waveUriRegExp.cap(1), m_waveUriRegExp.cap(2));
        if ( !c )
            c = createWaveContainer(m_waveUriRegExp.cap(1), m_waveUriRegExp.cap(2), false);
        if ( !c )
        {
            req->errorReply("Error: Wave already exists");
            return;
        }
        // Change the wave
        c->putRootDocument(req);
    }
    // http://host/wave/w+123/d+abc
    else if ( m_docUriRegExp.exactMatch(req->requestUri()) )
    {
        QString host = m_docUriRegExp.cap(1);
        WaveContainer* c = waveContainer(host, m_docUriRegExp.cap(2));
        if ( !c )
        {
            req->errorReply("Error: Wave does not exist");
            return;
        }
        c->putDocument(req, m_docUriRegExp.cap(3));
    }
    // http://host/wave/s+123
    else if ( m_sessionUriRegExp.exactMatch(req->requestUri()) )
    {
        Session* s = m_sessions.value(m_sessionUriRegExp.cap(1));
        if ( !s )
        {
            s = createSession(req, m_sessionUriRegExp.cap(1));
            if ( !s )
                return;
        }
        s->put(req);
    }
    // http://host/wave/_host/w+123/d+abc
    else if ( m_hostDocUriRegExp.exactMatch(req->requestUri()) )
    {
        WaveContainer* c = waveContainer(QString::null, m_hostDocUriRegExp.cap(1));
        if ( !c )
        {
            req->errorReply("Error: Wave does not exist");
            return;
        }
        c->putDocumentFromRemote(req, m_hostDocUriRegExp.cap(2));
    }
    // http://host/wave/_host/w+123
    else if ( m_hostWaveUriRegExp.exactMatch(req->requestUri()) )
    {
        WaveContainer* c = waveContainer(QString::null, m_hostWaveUriRegExp.cap(1));
        if ( !c )
        {
            req->errorReply("Error: Wave does not exist");
            return;
        }
        c->putDocumentFromRemote(req);
    }
    // http://host/wave/_remote/some.host.com
    else if ( m_remoteWaveUriRegExp.exactMatch(req->requestUri()) )
    {
        QString host = m_remoteWaveUriRegExp.cap(1);
        JSONObject doc(true);
        JSONScanner scanner(req->m_stdinStream.data(), req->m_stdinStream.size());
        bool ok = false;
        doc = scanner.scan(&ok);
        if ( !ok )
        {
            req->errorReply("JSON parsing error");
            return;
        }

        qDebug("%s", qPrintable(doc.toJSON()));

        // Iterate over the deltas for all the documents
        foreach( QString id, doc.attributeNames() )
        {
            // host.com/w+123
            if ( m_waveUriRegExp.exactMatch(id))
            {
                if ( !m_waveUriRegExp.cap(1).isEmpty() && m_waveUriRegExp.cap(1) != host + "/" )
                {
                    qDebug("Includes deltas from a remote host. We do no accept them");
                    continue;
                }

                // Find or create the wave
                WaveContainer* c = waveContainer(host, m_waveUriRegExp.cap(2));
                if ( !c )
                    c = createWaveContainer(host, m_waveUriRegExp.cap(2), true);
                if ( !c )
                {
                    req->errorReply("Error: Wave does not exist and cannot be created");
                    return;
                }
                qDebug("Created remote wave %s/%s", qPrintable(c->host()), qPrintable(c->waveId()));

                if ( !c->putRootDocumentFromHost(req, doc.attributeObject(id) ) )
                {
                    qDebug("Applying a delta to the root failed");
                    continue;
                }
            }
            // host.com/w+123/d+abc
            else if ( m_docUriRegExp.exactMatch(id))
            {
                if ( !m_docUriRegExp.cap(1).isEmpty() && m_docUriRegExp.cap(1) != host + "/" )
                {
                    qDebug("Includes deltas from a remote host. We do no accept them");
                    continue;
                }

                // Find the wave
                WaveContainer* c = waveContainer(host, m_docUriRegExp.cap(2));
                if ( !c )
                {
                    req->errorReply("Error: Wave does not exist");
                    return;
                }

                // Change the wave document
                if ( !c->putDocumentFromHost(req, m_docUriRegExp.cap(3), doc.attributeObject(id)) )
                {
                    qDebug("Applying a delta failed");
                    continue;
                }
            }
            else
            {
                qDebug("Uninterpreted attribute %s when parsing message from hosting server", qPrintable(id));
            }
        }

        JSONObject obj(true);
        obj.setAttribute("ok", true);
        req->replyJson(obj.toJSON());
    }
    else
        req->errorReply("Error: URI syntax error");
}

void WaveProvider::get(FCGI::FCGIRequest* req)
{
    if ( m_waveUriRegExp.exactMatch(req->requestUri()) )
    {
        QString host = m_waveUriRegExp.cap(1);
        if ( !host.isEmpty() )
            host = host.left(host.length() - 1);
        WaveContainer* c = waveContainer(host, m_waveUriRegExp.cap(2));
        if ( !c )
        {
            req->errorReply("Error: Wave does not exist");
            return;
        }
        c->getRootDocument(req);
    }
    else if ( m_docUriRegExp.exactMatch(req->requestUri()) )
    {
        QString host = m_docUriRegExp.cap(1);
        if ( !host.isEmpty() )
            host = host.left(host.length() - 1);
        WaveContainer* c = waveContainer(host, m_docUriRegExp.cap(2));
        if ( !c )
        {
            req->errorReply("Error: Wave does not exist");
            return;
        }
        c->getDocument(req, m_docUriRegExp.cap(3));
    }
    else if ( m_sessionUriRegExp.exactMatch(req->requestUri()) )
    {        
        Session* s = m_sessions.value(m_sessionUriRegExp.cap(1));
        if ( !s )
        {
            req->errorReply("Error: Session does not exist");
            return;
        }
        s->get(req);
    }
    else if ( m_sessionEventsUriRegExp.exactMatch(req->requestUri()) )
    {
        Session* s = m_sessions.value(m_sessionEventsUriRegExp.cap(1));
        if ( !s )
        {
            req->errorReply("Error: Session does not exist");
            return;
        }
        s->sendEvents(req);
    }
    else if ( m_sessionDeltasUriRegExp.exactMatch(req->requestUri()) )
    {
        Session* s = m_sessions.value(m_sessionDeltasUriRegExp.cap(1));
        if ( !s )
        {
            req->errorReply("Error: Session does not exist");
            return;
        }
        s->sendDeltas(req);
    }
    else
        req->errorReply("Error: URI syntax error");
}
