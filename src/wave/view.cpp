#include "view.h"
#include "js/jsengine.h"
#include "session.h"
#include "waveprovider.h"
#include <QUuid>

View::View(ViewContainer* parent, const QString& docId)
    : WaveDocument(parent, docId), m_malformed(true)
{
}

View::~View()
{
    clearIndices();
}

void View::update()
{
    clearIndices();
    m_digestMapFunction = QScriptValue();
    m_digestReduceFunction = QScriptValue();
    m_malformed = true;

    JSONObject digest = jsonObject().attributeObject("digest");
    if ( digest.isNull() )
        return;
    QString map = digest.attributeString("map");
    if ( map.isEmpty() )
        return;
    m_digestMapFunction = parseFunction(map);
    if ( m_digestMapFunction.isUndefined())
        return;
    QString reduce = digest.attributeString("reduce");
    if ( reduce.isEmpty() )
        return;
    m_digestReduceFunction = parseFunction(reduce);
    if ( m_digestReduceFunction.isUndefined() )
        return;
    JSONObject index = jsonObject().attributeObject("index");
    foreach( QString name, index.attributeNames() )
    {
        QString map = index.attributeString("map");
        if ( map.isEmpty() )
            return;
        QScriptValue func = parseFunction(map);
        if ( func.isUndefined())
            return;
        Index* i = new Index(func);
        m_indices.insert(name, i);
    }

    m_malformed = false;
}

QScriptValue View::parseFunction(const QString& js)
{
    qDebug("Parsing '%s'", qPrintable(js));
    JSEngine* engine = JSEngine::engine();
    QScriptValue v = engine->evaluate("(" + js + ");", waveId().toString());
    if ( engine->hasUncaughtException() )
    {
        qDebug("Malformed JS: %s", qPrintable(engine->uncaughtException().toString()));
        return QScriptValue();
    }
    qDebug("Javascript parsed successful");
    return v;
}

QStringList View::indexNames() const
{
    QStringList result;
    foreach( QString s, m_indices.keys())
        result.append(s);
    return result;
}

void View::clearIndices()
{
    foreach( Index* i, m_indices.values())
        delete i;
    m_indices.clear();
}

QScriptValue View::computeDigestMap(WaveContainer* c)
{
    return JSEngine::engine()->invokeMapOnContainer( m_digestMapFunction, c );
}

QScriptValue View::computeDigestReduce(WaveContainer* c)
{
    return JSEngine::engine()->invokeReduceOnContainer( this->documentId(), m_digestReduceFunction, c );
}

QString View::registerSessionQuery( const Query& query )
{
    QString qid = QUuid::createUuid().toString();
    m_sessionQueries.insert(qid, query);
    return qid;
}

void View::notifySessionQueries( QHash<QString,IndexItemList> newIndexItems )
{
    foreach( QString qid, m_sessionQueries.keys() )
    {
        QHash<QString,IndexItemList> news;
        foreach( QString dbId, newIndexItems.keys() )
        {
            foreach( IndexItem item, newIndexItems.value(dbId) )
            {
                Query q = m_sessionQueries.value(qid);
                if ( q.userJID() == item.key().at(0).toString())
                {
                    if ( !news.contains(dbId))
                        news.insert(dbId, IndexItemList());
                    news[dbId].append(item);
                }
            }
        }

        if ( !news.isEmpty() )
        {
            QString sessionId = m_sessionQueries.value(qid).sessionID();
            // Get a pointer to the session
            Session *s = WaveProvider::self()->session(sessionId);
            if ( !s )
            {
                qDebug("Oooops, new session is already dead: %s", qPrintable(sessionId));
                continue;
            }
            s->notify(documentId(), qid, newIndexItems);
        }
    }
}
