#ifndef JSENGINE_H
#define JSENGINE_H

#include <QScriptEngine>
#include <QScriptValue>
#include "json/jsonobject.h"

class JSEngine : public QScriptEngine
{
public:
    JSEngine( QObject* parent = 0);

    QScriptValue fromJSON(JSONAbstractObject obj);
    JSONAbstractObject toJSON(const QScriptValue& value);
};

#endif
