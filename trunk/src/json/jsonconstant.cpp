#include "jsonconstant.h"

JSONConstant::JSONConstant()
{
}

JSONConstant::JSONConstant( const JSONConstant& c )
    : JSONAbstractObject(c)
{
}

JSONConstant::JSONConstant( const QString& str )
    : JSONAbstractObject()
{
    if ( !str.isNull() )
    {
        ensureData();
        data()->variant.setValue(str);
    }
}

JSONConstant::JSONConstant( int i )
    : JSONAbstractObject()
{
    ensureData();
    data()->variant.setValue(i);
}

JSONConstant::JSONConstant( double d )
    : JSONAbstractObject()
{
    ensureData();
    data()->variant.setValue(d);
}

JSONConstant::JSONConstant( bool b )
    : JSONAbstractObject()
{
    ensureData();
    data()->variant.setValue(b);
}

JSONConstant::JSONConstant( const char* str )
    : JSONAbstractObject()
{
    if ( str )
    {
        ensureData();
        data()->variant.setValue(QString(str));
    }
}

JSONConstant::JSONConstant( const QVariant& value )
{
    ensureData();
    data()->variant = value;
}

QString JSONConstant::ConstantData::toJSON() const
{
    switch( variant.type() )
    {
    case QVariant::String:
        {
            QString result = variant.toString();
            result.replace( '\\', "\\\\" );
            result.replace( '"', "\\\"" );
            result.replace( '\b', "\\b" );
            result.replace( '\f', "\\f" );
            result.replace( '\n', "\\n" );
            result.replace( '\r', "\\r" );
            result.replace( '\t', "\\t" );
            return "\"" + result + "\"";
        }
    case QVariant::Int:
        return QString::number(variant.toInt());
    case QVariant::Double:
        return QString::number(variant.toDouble());
    case QVariant::Bool:
        if (variant.toBool())
            return "true";
        return "false";
    default:
        return "null";
    }      
}

JSONConstant::Data* JSONConstant::ConstantData::clone() const
{
    ConstantData* d = new ConstantData();
    d->counter--;
    d->variant = variant;
    return d;
}

bool JSONConstant::ConstantData::equals( JSONAbstractObject::Data* data )
{
    JSONConstant::ConstantData* d = dynamic_cast<JSONConstant::ConstantData*>(data);
    if ( !d )
        return false;
    return ( d->variant == variant );
}
