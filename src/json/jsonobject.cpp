#include "jsonobject.h"
#include "jsonarray.h"
#include "jsonconstant.h"

JSONObject::JSONObject()
    : JSONAbstractObject()
{
}

JSONObject::JSONObject(bool create_empty_object)
    : JSONAbstractObject()
{
    if ( create_empty_object )
        becomeObject();
}

JSONObject::JSONObject( const JSONObject& obj )
    : JSONAbstractObject( obj.m_data )
{
}

JSONAbstractObject JSONObject::attribute( const QString& name ) const
{
    if ( !m_data )
        return JSONAbstractObject();
    return data()->objects.value(name);
}

JSONObject JSONObject::attributeObject( const QString& name ) const
{
    if ( !m_data )
        return JSONObject();
    return data()->objects.value(name).toObject();
}

JSONArray JSONObject::attributeArray( const QString& name ) const
{
    if ( !m_data )
        return JSONArray();
    return data()->objects.value(name).toArray();
}

QString JSONObject::attributeString( const QString& name ) const
{
    if ( !m_data )
        return QString::null;
    return data()->objects.value(name).toString();
}

void JSONObject::setAttribute( const QString& name, const JSONAbstractObject& obj )
{
    if ( !obj.m_data )
    {
        if ( m_data )
            data()->objects.remove(name);
        return;
    }
    ensureData();
    if ( obj.m_data->parent != m_data )
    {
        if ( obj.m_data->parent )
        {
            obj.m_data->parent->removeChild(obj.m_data);
        }
        obj.m_data->parent = m_data;
    }

    data()->objects.insert(name, obj);
}

void JSONObject::setAttribute( const QString& name, const QString& str )
{
    if ( str.isNull() )
    {
        if ( m_data )
            data()->objects.remove(name);
        return;
    }
    ensureData();
    data()->objects.insert(name, JSONConstant(str));
}

void JSONObject::setAttribute( const QString& name, const char* str )
{
    if ( !str )
    {
        if ( m_data )
            data()->objects.remove(name);
        return;
    }
    ensureData();
    data()->objects.insert(name, JSONConstant(str));
}

void JSONObject::setAttribute( const QString& name, int i )
{
    ensureData();
    data()->objects.insert(name, JSONConstant(i));
}

void JSONObject::setAttribute( const QString& name, bool i )
{
    ensureData();
    data()->objects.insert(name, JSONConstant(i));
}

void JSONObject::setAttribute( const QString& name, double i )
{
    ensureData();
    data()->objects.insert(name, JSONConstant(i));
}

void JSONObject::removeAttribute( const QString& name )
{
    if ( m_data ) data()->objects.remove(name);
}

QSet<QString> JSONObject::attributeNamesSet() const
{
    QSet<QString> s;
    if ( m_data )
    {
        foreach( QString key, data()->objects.keys() )
            s.insert(key);
    }
    return s;
}

void JSONObject::ObjectData::removeChild(JSONAbstractObject::Data* d)
{
    foreach( QString key, objects.keys())
    {
        JSONAbstractObject obj = objects[key];
        if ( obj.m_data == d)
        {
            objects.remove(key);
            return;
        }
    }
}

JSONObject::ObjectData::~ObjectData()
{
    foreach( JSONAbstractObject obj, objects.values())
    {
        if (obj.m_data)
            obj.m_data->parent = 0;
    }
}

QString JSONObject::ObjectData::stringify( const QString& str ) const
{
    QString result("x" + str);
    result.replace( '\\', "\\\\" );
    result.replace( '"', "\\\"" );
    result.replace( '\b', "\\b" );
    result.replace( '\f', "\\f" );
    result.replace( '\n', "\\n" );
    result.replace( '\r', "\\r" );
    result.replace( '\t', "\\t" );
    result[0] = '"';
    result.append('"');
    return result;
}

QString JSONObject::ObjectData::toJSON() const
{
    QString result = "{";

    bool first = true;
    foreach(QString key, objects.keys())
    {
        if ( !first )
            result += ",";
        else
            first = false;
        JSONAbstractObject obj = objects[key];
        result += stringify(key) + ":" + obj.toJSON();
    }

    result += "}";
    return result;
}

JSONObject::Data* JSONObject::ObjectData::clone() const
{
    ObjectData* d = new ObjectData();
    d->counter--;
    foreach( QString key, objects.keys())
    {
        JSONAbstractObject obj = objects[key];
        d->objects.insert( key, obj.clone() );
    }
    return d;
}

bool JSONObject::ObjectData::equals( JSONAbstractObject::Data* data )
{
    JSONObject::ObjectData* d = dynamic_cast<JSONObject::ObjectData*>(data);
    if ( !d )
        return false;
    if ( d->objects.count() != objects.count() )
        return false;
    foreach( QString name, objects.keys())
    {
        if ( !objects[name].equals(d->objects[name]))
            return false;
    }
    return true;
}

bool JSONObject::ObjectData::lessThan( const JSONAbstractObject::Data* data ) const
{
    return this->toJSON() < data->toJSON();
}
