#include "view.h"
#include "viewcontainer.h"

View::View(ViewContainer* parent, const QString& docId)
    : WaveDocument(parent, docId)
{
}

void View::update()
{
    QString map = jsonObject().attributeString("map");
    if ( map.isEmpty() )
        m_mapFunction = QScriptValue();
    else
        m_mapFunction = parseFunction(map);

    QString reduce = jsonObject().attributeString("reduce");
    if ( reduce.isEmpty() )
        m_reduceFunction = QScriptValue();
    else
        m_reduceFunction = parseFunction(reduce);
}

QScriptValue View::parseFunction(const QString& js)
{
    QScriptValue v = m_scriptEngine.evaluate(js, waveId().toString());
    if ( m_scriptEngine.hasUncaughtException() )
    {
        qDebug("Malformed JS");
        return QScriptValue();
    }
    qDebug("Javascript parsed successful");
    return v;
}
