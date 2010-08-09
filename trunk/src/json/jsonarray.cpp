#include "jsonarray.h"
#include "jsonconstant.h"

JSONArray::JSONArray()
    : JSONAbstractObject()
{
}

JSONArray::JSONArray(bool create_empty)
    : JSONAbstractObject()
{
    if ( create_empty )
        ensureData();
}

JSONArray::JSONArray( const JSONArray& arr )
    : JSONAbstractObject( arr )
{
}

bool JSONArray::replace(int i, const JSONAbstractObject& obj)
{
    if ( !m_data )
        return false;
    if ( i < 0 || i >= data()->arr.count() )
        return false;
    data()->arr.replace(i, obj);
    return true;
}

bool JSONArray::insert(int i, const JSONAbstractObject& obj)
{
    if ( !m_data )
        return false;
    if ( i < 0 || i > data()->arr.count() )
        return false;
    data()->arr.insert(i, obj);
    return true;
}

bool JSONArray::removeAt(int i)
{
    if ( !m_data )
        return false;
    if ( i < 0 || i >= data()->arr.count() )
        return false;
    data()->arr.removeAt(i);
    return true;
}

void JSONArray::remove(const JSONAbstractObject& obj)
{
    if ( !m_data)
        return;
    data()->arr.removeOne(obj);
}

void JSONArray::append( const JSONAbstractObject& obj)
{
    ensureData();
    data()->arr.append(obj);
}

void JSONArray::append( const QString& str)
{
    ensureData();
    data()->arr.append(JSONConstant(str));
}

void JSONArray::append( const char* str)
{
    ensureData();
    data()->arr.append(JSONConstant(QString(str)));
}

void JSONArray::append( int i)
{
    ensureData();
    data()->arr.append(JSONConstant(i));
}

void JSONArray::append( double d)
{
    ensureData();
    data()->arr.append(JSONConstant(d));
}

void JSONArray::append( bool b)
{
    ensureData();
    data()->arr.append(JSONConstant(b));
}

int JSONArray::indexOf(const JSONAbstractObject& obj) const
{
    if ( !m_data)
        return -1;
    return data()->arr.indexOf(obj);
}

void JSONArray::ArrayData::removeChild(JSONAbstractObject::Data* d)
{
    for( int i = 0; i < arr.count(); ++i )
    {
        if ( arr[i].m_data == d )
        {
            arr.removeAt(i);
            return;
        }
    }
}

JSONArray::ArrayData::~ArrayData()
{
    for( int i = 0; i < arr.count(); ++i )
        if ( arr[i].m_data )
            arr[i].m_data->parent = 0;
}

QString JSONArray::ArrayData::toJSON() const
{
    QString result = "[";

    bool first = true;
    foreach(JSONAbstractObject obj, arr)
    {
        if ( !first )
            result += ",";
        else
            first = false;
        result += obj.toJSON();
    }

    result += "]";
    return result;
}

JSONArray::Data* JSONArray::ArrayData::clone() const
{
    ArrayData* d = new ArrayData();
    d->counter--;
    foreach( JSONAbstractObject obj, arr)
    {
        d->arr.append( obj.clone() );
    }
    return d;
}

bool JSONArray::ArrayData::equals( JSONAbstractObject::Data* data )
{
    JSONArray::ArrayData* d = dynamic_cast<JSONArray::ArrayData*>(data);
    if ( !d )
        return false;
    if ( d->arr.count() != arr.count() )
        return false;
    for( int i = 0; i < arr.count(); ++i )
    {
        if ( !arr[i].equals(d->arr[i]))
            return false;
    }
    return true;
}
