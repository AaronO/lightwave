#ifndef VIEW_H
#define VIEW_H

#include "wavecontainer.h"
#include <QScriptValue>
#include <QScriptEngine>

class ViewContainer;

class View : public WaveDocument
{
public:
    View(ViewContainer* parent, const QString& docId);

    void update();

private:
    QScriptValue parseFunction(const QString& js);

    QScriptValue m_mapFunction;
    QScriptValue m_reduceFunction;
    QScriptEngine m_scriptEngine;
};

#endif // VIEW_H
