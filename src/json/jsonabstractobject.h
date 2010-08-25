#ifndef JSONABSTRACTOBJECT_H
#define JSONABSTRACTOBJECT_H

#include <QString>
#include <QtGlobal>

class JSONObject;
class JSONArray;
class JSONConstant;

class JSONAbstractObject
{
public:
    JSONAbstractObject();
    JSONAbstractObject(const JSONAbstractObject& obj);
    ~JSONAbstractObject();

    JSONAbstractObject& operator=(const JSONAbstractObject& obj);
    bool operator==(const JSONAbstractObject& obj) const { return m_data == obj.m_data; }
    bool operator!=(const JSONAbstractObject& obj) const { return m_data != obj.m_data; }
    bool equals(const JSONAbstractObject& obj) const;
    bool operator<(const JSONAbstractObject& obj) const { if ( m_data == obj.m_data ) return false; if ( !m_data ) return true; if ( !obj.m_data ) return false; return m_data->lessThan(obj.m_data); }
    bool operator<=(const JSONAbstractObject& obj) const { if ( equals(obj) ) return true; if ( !m_data ) return true; if ( !obj.m_data ) return false; return m_data->lessThan(obj.m_data); }
    bool operator>(const JSONAbstractObject& obj) const { return !equals(obj) && !(*this < obj); }
    bool operator>=(const JSONAbstractObject& obj) const { return !(*this < obj); }

    bool isNull() const { return m_data == 0; }
    bool isObject() const;
    bool isArray() const;
    bool isConstant() const;
    bool isString() const;
    bool isInt() const;
    bool isDouble() const;
    bool isBool() const;
    /**
      * Represents the JavaScript null value.
      */
    bool isNullValue() const;

    JSONObject toObject() const;
    JSONArray toArray() const;
    JSONConstant toConstant() const;
    int toInt() const;
    QString toString() const;
    double toDouble() const;
    bool toBool() const;

    JSONAbstractObject parent() const;

    QString toJSON() const;

    JSONAbstractObject clone() const;
    /**
      * The object looses its internal state and become null.
      */
    void clear();

    class Data
    {
    public:
        Data() : counter(1), parent(0) { };
        virtual ~Data() { }

        int counter;
        Data* parent;

        virtual Data* clone() const = 0;
        virtual void removeChild(Data* d) = 0;
        virtual QString toJSON() const = 0;
        virtual bool equals( Data* data ) = 0;
        virtual bool lessThan( const Data* data ) const = 0;
    };

    JSONAbstractObject(Data* data) : m_data(data) { if ( m_data) m_data->counter++; }

    Data* m_data;

    void becomeObject();
};

#endif // JSONABSTRACTOBJECT_H
