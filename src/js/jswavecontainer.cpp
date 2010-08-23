#include "jswavecontainer.h"
#include "jsengine.h"
#include "wave/wavedocument.h"
#include "wave/wavecontainer.h"
#include "wave/waveprovider.h"
#include "wave/waveid.h"

JSWaveContainerClass* JSWaveContainerClass::s_self = 0;

JSWaveContainerClass::JSWaveContainerClass(JSEngine* engine)
    : QScriptClass(engine)
{
}

QScriptClass::QueryFlags JSWaveContainerClass::queryProperty ( const QScriptValue & object, const QScriptString & name, QueryFlags flags, uint * id )
{
    Q_UNUSED(object);
    Q_UNUSED(name);
    Q_UNUSED(flags);
    Q_UNUSED(id);

    return HandlesReadAccess | HandlesWriteAccess;
}

QScriptValue JSWaveContainerClass::property( const QScriptValue& object, const QScriptString& name, uint id )
{
    qDebug("... asking for property %s", qPrintable(name.toString()));

    Q_UNUSED(id);

    QScriptValue data = object.data();
    Q_ASSERT ( data.isObject() );
    QScriptValue cache = data.property(name, QScriptValue::ResolveLocal);
    if ( cache.isValid() )
        return cache;

    WaveContainer* c = WaveProvider::self()->container( WaveId( data.property("waveid", QScriptValue::ResolveLocal).toString() ));
    if ( !c )
        return QScriptValue();
    WaveDocument* doc = c->document(name);
    if ( !doc )
        return QScriptValue();
    QScriptValue v = JSEngine::engine()->fromJSON(doc->jsonObject());
    if ( data.isUndefined() )
        data = object.engine()->newObject();
    data.setProperty(name, v);
    const_cast<QScriptValue&>(object).setData(data);
    return v;
}

QScriptValue JSWaveContainerClass::createWrapper(JSEngine* engine, const WaveId& waveid)
{
    if ( s_self == 0 )
        s_self = new JSWaveContainerClass(engine);

    QScriptValue obj = engine->newObject(s_self);
    QScriptValue data = engine->newObject();
    data.setProperty("waveid", waveid.toString());
    obj.setData(data);
    return obj;
}
