#include "view.h"
#include "js/jsengine.h"

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
        QString map = digest.attributeString("map");
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
    return JSEngine::engine()->invokeOnContainer( m_digestMapFunction, c );
}

QScriptValue View::computeDigestReduce(WaveContainer* c)
{
    return JSEngine::engine()->invokeOnContainer( m_digestReduceFunction, c );
}
