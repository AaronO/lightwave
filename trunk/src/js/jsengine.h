#ifndef JSENGINE_H
#define JSENGINE_H

#include <QScriptEngine>
#include <QScriptValue>
#include "json/jsonobject.h"
#include "jsemitclass.h"

class WaveContainer;

class JSEngine : public QScriptEngine
{
public:
    JSEngine( QObject* parent = 0);

    QScriptValue fromJSON(JSONAbstractObject obj);
    JSONAbstractObject toJSON(const QScriptValue& value);

    QScriptValue invokeMapOnContainer( const QScriptValue& func, WaveContainer* container );
    QScriptValue invokeReduceOnContainer( const QString& viewId, const QScriptValue& func, WaveContainer* container );
    QScriptValue invokeMapOnDigest( const QScriptValue& func, const QScriptValue& digest );

    static JSEngine* engine();

private:
    JSEmitClass m_emitClass;

    static JSEngine* s_self;
};

#endif
