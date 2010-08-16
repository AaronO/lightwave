#include "rootcontainer.h"
#include "hostcontainer.h"
#include "json/jsonscanner.h"

RootContainer::RootContainer()
    : WaveContainer(0, "")
{
}

JSONObject RootContainer::put(FCGI::FCGIRequest* req, const WaveId& waveId)
{
    WaveContainer* c = getOrCreateContainer(waveId, true);
    if ( !c )
    {
        JSONObject e(true);
        e.setAttribute("ok", false);
        e.setAttribute("error", "Could not create wave " + waveId.toString());
        return e;
    }

    JSONScanner scanner(req->m_stdinStream.data(), req->m_stdinStream.size());
    bool ok = false;
    JSONObject doc = scanner.scan(&ok);
    if ( !ok )
    {
        JSONObject e(true);
        e.setAttribute("ok", false);
        e.setAttribute("error", "JSON parsing error");
        return e;
    }

    if ( doc.hasAttribute("_snapshot") && doc.attribute("_snapshot").toBool() )
    {
        qDebug("Error: A client must not send snapshots");
        JSONObject e(true);
        e.setAttribute("ok", false);
        e.setAttribute("error", "A client must not send snapshots");
        return e;
    }

    JSONObject result = c->put(doc, waveId.documentId(), req);
    if ( c->isTemporary() && !result.attribute("ok").toBool() )
        delete c;
    else if ( c->isTemporary() )
        c->makePersistent();
    return result;
}

JSONObject RootContainer::get(FCGI::FCGIRequest* req, const WaveId& waveId)
{
    WaveContainer* c = getOrCreateContainer(waveId, false);
    if ( !c )
    {
        JSONObject e(true);
        e.setAttribute("ok", false);
        e.setAttribute("error", "Could not find wave");
        return e;
    }

    return c->get(req, waveId.documentId());
}

JSONObject RootContainer::putFromHostingServer(JSONObject data, const WaveId& waveId)
{
    WaveContainer* c = getOrCreateContainer(waveId, true);
    if ( !c )
    {
        JSONObject e(true);
        e.setAttribute("ok", false);
        e.setAttribute("error", "Could not create remote wave");
        return e;
    }

    JSONObject result;
    if ( data.hasAttribute("_snapshot") && data.attribute("_snapshot").toBool() )
        result = c->putSnapshotFromHost(data, waveId.documentId());
    else
        result = c->putFromHost(data, waveId.documentId());

    if ( c->isTemporary() && !result.attribute("ok").toBool() )
        delete c;
    else if ( c->isTemporary() )
        c->makePersistent();
    return result;
}

JSONObject RootContainer::putFromRemoteServer(JSONObject data, const WaveId& waveId)
{
    if ( data.hasAttribute("_snapshot") && data.attribute("_snapshot").toBool() )
    {
        qDebug("Error: A remote host must not send snapshots");
        JSONObject e(true);
        e.setAttribute("ok", false);
        e.setAttribute("error", "Remote server must not send snapshots");
        return e;
    }

    WaveContainer* c = getOrCreateContainer(waveId, true);
    if ( !c )
    {
        JSONObject e(true);
        e.setAttribute("ok", false);
        e.setAttribute("error", "Could not create wave on behalf of remote server");
        return e;
    }

    JSONObject result = c->putFromRemote(data, waveId.documentId());
    if ( c->isTemporary() && !result.attribute("ok").toBool() )
        delete c;
    else if ( c->isTemporary() )
        c->makePersistent();
    return result;
}


WaveContainer* RootContainer::getOrCreateContainer(const WaveId& waveId, bool allow_creation)
{
    if ( waveId.isNull() )
        return 0;
    WaveContainer* c = childContainer(waveId.host());
    if ( !c && allow_creation )
    {
        c = createWaveContainer(waveId.host());
        if ( c )
            c->makePersistent();
    }
    if ( !c )
        return 0;
    for( int i = 0; i < waveId.pathItemCount(); ++i )
    {
        QString name = waveId.pathItem(i);
        WaveContainer* c2 = c->childContainer(name);
        if ( !c2 && i == waveId.pathItemCount() - 1 && allow_creation )
            c2 = c->createWaveContainer(name);
        if ( !c2 )
            return 0;
        c = c2;
    }
    return c;
}

WaveContainer* RootContainer::createWaveContainer(const QString& name)
{
    Q_ASSERT(container(name) == 0);
    HostContainer* c = new HostContainer(this, name, false);
    return c;
}
