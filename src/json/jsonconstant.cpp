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

bool JSONConstant::ConstantData::lessThan( const JSONAbstractObject::Data* data ) const
{
    const JSONConstant::ConstantData* d = dynamic_cast<const JSONConstant::ConstantData*>(data);
    if ( !d )
        return this->toJSON() < data->toJSON();
    if ( variant.isNull() && d->variant.isNull())
        return false;
    if ( variant.isNull() )
        return true;
    if ( d->variant.isNull() )
        return false;
    QVariant::Type t1 = variant.type();
    if ( t1 == QVariant::Int )
        t1 = QVariant::Double;
    QVariant::Type t2 = d->variant.type();
    if ( t2 == QVariant::Int )
        t2 = QVariant::Double;
    if ( t1 != t2 )
        return this->toJSON() < data->toJSON();
    switch( variant.type() )
    {
    case QVariant::String:
        return variant.toString() < d->variant.toString();
    case QVariant::Int:
    case QVariant::Double:
        return variant.toDouble() < d->variant.toDouble();
    case QVariant::Bool:
        return variant.toBool() < d->variant.toBool();
    default:
        break;
    }
    return this->toJSON() < data->toJSON();
}
