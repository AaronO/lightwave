#ifndef JSWAVECONTAINER_H
#define JSWAVECONTAINER_H

#include <QScriptClass>
#include <QScriptValue>

class WaveId;
class JSEngine;

/**
  * A JavaScript wrapper around a WaveContainer.
  */
class JSWaveContainerClass : public QScriptClass
{
public:
    JSWaveContainerClass(JSEngine* engine);

    QueryFlags queryProperty ( const QScriptValue & object, const QScriptString & name, QueryFlags flags, uint * id );
    QScriptValue property( const QScriptValue & object, const QScriptString& name, uint id );

    static QScriptValue createWrapper(JSEngine* engine, const WaveId& waveid);

private:
    static JSWaveContainerClass* s_self;
};

#endif // JSWAVECONTAINER_H
