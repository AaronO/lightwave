#ifndef JSONARRAY_H
#define JSONARRAY_H

#include <QList>
#include "jsonabstractobject.h"

class JSONArray : public JSONAbstractObject
{
public:
    JSONArray();
    JSONArray(bool create_empty);
    JSONArray( const JSONArray& arr );

    inline int count() const { if ( m_data ) return data()->arr.count(); return 0; }
    inline JSONAbstractObject at( int i ) const { if ( m_data ) return data()->arr.at(i); return JSONAbstractObject(); }

    bool replace(int i, const JSONAbstractObject& obj);
    bool insert(int i, const JSONAbstractObject& obj);

    bool removeAt(int i);
    void remove(const JSONAbstractObject& obj);

    void append( const JSONAbstractObject& obj);
    void append( const QString& str);
    void append( const char* str);
    void append( int i);
    void append( double d);
    void append( bool b);

    int indexOf(const JSONAbstractObject& obj) const;

    QList<JSONAbstractObject> content() const { if (m_data) return data()->arr; return QList<JSONAbstractObject>(); }

    JSONAbstractObject operator[](int index) const { if (m_data) return data()->arr.at(index); return JSONAbstractObject(); }

    class ArrayData : public Data
    {
    public:
        ArrayData() { }
        ~ArrayData();

        QList<JSONAbstractObject> arr;

        void removeChild(Data* d);
        QString toJSON() const;
        Data* clone() const;
        bool equals( Data* data );
    };

    JSONArray(ArrayData* data) : JSONAbstractObject(data) { }

private:
    inline void ensureData() { if ( !m_data ) m_data = new ArrayData(); }
    inline ArrayData* data() const { return static_cast<ArrayData*>(m_data); }

};

#endif // JSONARRAY_H
