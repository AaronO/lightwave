#include "jsonabstractobject.h"
#include "jsonarray.h"
#include "jsonobject.h"
#include "jsonconstant.h"

JSONAbstractObject::JSONAbstractObject()
    : m_data(0)
{
}

JSONAbstractObject::JSONAbstractObject(const JSONAbstractObject& obj)
    : m_data( obj.m_data)
{
    if ( m_data )
        m_data->counter++;
}

JSONAbstractObject::~JSONAbstractObject()
{
    if ( m_data )
    {
        m_data->counter--;
        if ( m_data->counter == 0 )
            delete m_data;
    }
}

JSONAbstractObject& JSONAbstractObject::operator=(const JSONAbstractObject& obj)
{
    if ( m_data == obj.m_data )
        return *this;
    if ( m_data )
    {
        m_data->counter--;
        if ( m_data->counter == 0 )
            delete m_data;
    }    
    m_data = obj.m_data;
    if ( m_data )
    {
        m_data->counter++;
    }
    return *this;
}

void JSONAbstractObject::clear()
{
    if ( m_data )
    {
        m_data->counter--;
        if ( m_data->counter == 0 )
            delete m_data;
        m_data = 0;
    }
}

bool JSONAbstractObject::isObject() const
{
    return dynamic_cast<JSONObject::ObjectData*>( m_data );
}

bool JSONAbstractObject::isArray() const
{
    return dynamic_cast<JSONArray::ArrayData*>( m_data );
}

bool JSONAbstractObject::isConstant() const
{
    return dynamic_cast<JSONConstant::ConstantData*>( m_data );
}

bool JSONAbstractObject::isString() const
{
    JSONConstant::ConstantData* c = dynamic_cast<JSONConstant::ConstantData*>( m_data );
    return ( c && c->variant.type() == QVariant::String );
}

bool JSONAbstractObject::isInt() const
{
    JSONConstant::ConstantData* c = dynamic_cast<JSONConstant::ConstantData*>( m_data );
    return ( c && c->variant.type() == QVariant::Int );
}

bool JSONAbstractObject::isDouble() const
{
    JSONConstant::ConstantData* c = dynamic_cast<JSONConstant::ConstantData*>( m_data );
    return ( c && c->variant.type() == QVariant::Double );
}

bool JSONAbstractObject::isBool() const
{
    JSONConstant::ConstantData* c = dynamic_cast<JSONConstant::ConstantData*>( m_data );
    return ( c && c->variant.type() == QVariant::Bool );
}

bool JSONAbstractObject::isNullValue() const
{
    JSONConstant::ConstantData* c = dynamic_cast<JSONConstant::ConstantData*>( m_data );
    return ( c && c->variant.isNull() );
}

JSONObject JSONAbstractObject::toObject() const
{
    return JSONObject(dynamic_cast<JSONObject::ObjectData*>( m_data ));
}

JSONArray JSONAbstractObject::toArray() const
{
    return JSONArray(dynamic_cast<JSONArray::ArrayData*>( m_data ));
}

JSONConstant JSONAbstractObject::toConstant() const
{
    return JSONConstant(dynamic_cast<JSONConstant::ConstantData*>( m_data ));
}

int JSONAbstractObject::toInt() const
{
    JSONConstant::ConstantData* c = dynamic_cast<JSONConstant::ConstantData*>( m_data );
    Q_ASSERT(c);
    return c->variant.toInt();
}

bool JSONAbstractObject::toBool() const
{
    JSONConstant::ConstantData* c = dynamic_cast<JSONConstant::ConstantData*>( m_data );
    Q_ASSERT(c);
    return c->variant.toBool();
}

QString JSONAbstractObject::toString() const
{
    JSONConstant::ConstantData* c = dynamic_cast<JSONConstant::ConstantData*>( m_data );
    if ( !c )
        return QString::null;
    return c->variant.toString();
}

double JSONAbstractObject::toDouble() const
{
    JSONConstant::ConstantData* c = dynamic_cast<JSONConstant::ConstantData*>( m_data );
    Q_ASSERT(c);
    return c->variant.toDouble();
}

JSONAbstractObject JSONAbstractObject::parent() const
{
    if ( !m_data)
        return JSONAbstractObject();
    return JSONAbstractObject(m_data->parent);
}

QString JSONAbstractObject::toJSON() const
{
    if ( !m_data )
        return "null";
    return m_data->toJSON();
}

JSONAbstractObject JSONAbstractObject::clone() const
{
    if ( !m_data )
        return JSONAbstractObject();
    return JSONAbstractObject(m_data->clone());
}

void JSONAbstractObject::becomeObject()
{
    if ( isObject() )
        return;
    if ( !isNull() )
        clear();
    m_data = new JSONObject::ObjectData();
}

bool JSONAbstractObject::equals(const JSONAbstractObject& obj) const
{
    if ( m_data == obj.m_data )
        return true;
    if ( !m_data || !obj.m_data )
        return false;
    return m_data->equals(obj.m_data);
}
