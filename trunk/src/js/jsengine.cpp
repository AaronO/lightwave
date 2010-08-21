#include "jsengine.h"
#include "json/jsonarray.h"
#include "json/jsonconstant.h"
#include <QScriptValueIterator>

JSEngine::JSEngine( QObject* parent)
    : QScriptEngine(parent)
{
}

QScriptValue JSEngine::fromJSON(JSONAbstractObject obj)
{
    if ( obj.isNull() )
        return QScriptValue(QScriptValue::UndefinedValue);
    if ( obj.isNullValue() )
        return QScriptValue(QScriptValue::NullValue);
    if ( obj.isBool() )
        return QScriptValue(obj.toBool());
    if ( obj.isDouble() )
        return QScriptValue((qsreal)obj.toDouble());
    if ( obj.isInt() )
        return QScriptValue(obj.toInt());
    if ( obj.isString() )
        return QScriptValue(obj.toString());
    if ( obj.isArray())
    {
        JSONArray arr = obj.toArray();
        QScriptValue v = newArray(arr.count());
        for( int i = 0; i < arr.count(); ++i )
            v.setProperty(i, fromJSON(arr.at(i)));
        return v;
    }
    if ( obj.isObject())
    {
        JSONObject o = obj.toObject();
        QScriptValue v = newObject();
        foreach(QString key, o.attributeNames())
        {
            v.setProperty(key, fromJSON(o.attribute(key)));
        }
        return v;
    }

    Q_ASSERT(false);
    return QScriptValue();
}

JSONAbstractObject JSEngine::toJSON(const QScriptValue& value)
{
    if ( value.isUndefined())
        return JSONAbstractObject();
    if ( value.isNull())
        return JSONConstant::createNull();
    if ( value.isBool() )
        return JSONConstant(value.toBool());
    if ( value.isNumber() )
        return JSONConstant(value.toNumber());
    if ( value.isString() )
        return JSONConstant(value.toString());
    if ( value.isArray() )
    {
        JSONArray arr(true);
        quint32 i = 0;
        while(true)
        {
            QScriptValue v = value.property(i);
            if ( v.isUndefined() )
                break;
            arr.append( toJSON(v));
        }
        return arr;
    }
    if ( value.isObject() )
    {
        JSONObject o(true);
        QScriptValueIterator it(value);
        while (it.hasNext())
        {
            it.next();
            o.setAttribute(it.name(), toJSON( it.value()));
        }
        return o;
    }

    Q_ASSERT(false);
    return JSONAbstractObject();
}

