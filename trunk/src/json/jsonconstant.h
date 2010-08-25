#ifndef JSONCONSTANT_H
#define JSONCONSTANT_H

#include "jsonabstractobject.h"
#include <QVariant>

class JSONConstant : public JSONAbstractObject
{
public:
    JSONConstant();
    JSONConstant( const JSONConstant& c );
    JSONConstant( const QString& str );
    JSONConstant( const char* str );
    JSONConstant( int i );
    JSONConstant( double d );
    JSONConstant( bool b );
    JSONConstant( const QVariant& value );

    QVariant value() const { if ( !m_data ) return QVariant(); return data()->variant; }

    class ConstantData : public Data
    {
    public:
        QVariant variant;

        void removeChild(Data*) { return; }
        QString toJSON() const;
        Data* clone() const;
        bool equals( Data* data );
        virtual bool lessThan( const Data* data ) const;
    };

    JSONConstant(ConstantData* data) : JSONAbstractObject(data) { }

    static JSONConstant createNull() { return JSONConstant(new ConstantData()); }

private:
    inline void ensureData() { if ( !m_data ) m_data = new ConstantData(); }
    inline ConstantData* data() const { return static_cast<ConstantData*>(m_data); }

};

#endif // JSONCONSTANT_H
