#ifndef JSONOBJECT_H
#define JSONOBJECT_H

#include "jsonabstractobject.h"
#include <QHash>
#include <QList>
#include <QString>
#include <QSet>

class JSONArray;

class JSONObject : public JSONAbstractObject
{
public:
    JSONObject();
    JSONObject(bool create_empty_object);
    JSONObject( const JSONObject& obj );

    JSONAbstractObject attribute( const QString& name ) const;
    JSONObject attributeObject( const QString& name ) const;
    JSONArray attributeArray( const QString& name ) const;
    QString attributeString( const QString& name ) const;

    bool hasAttribute( const QString& name ) { if ( m_data ) return data()->objects.contains(name); return false; }

    void setAttribute( const QString& name, const JSONAbstractObject& obj );
    void setAttribute( const QString& name, const QString& str );
    void setAttribute( const QString& name, const char* str );
    void setAttribute( const QString& name, int i );
    void setAttribute( const QString& name, bool b );
    void setAttribute( const QString& name, double d );

    void removeAttribute( const QString& name );

    QList<QString> attributeNames() const { if ( m_data) return data()->objects.keys(); return QList<QString>(); }
    QSet<QString> attributeNamesSet() const;
    QList<JSONAbstractObject> attributeValues() const { if ( m_data) return data()->objects.values(); return QList<JSONAbstractObject>(); }

    class ObjectData : public Data
    {
    public:
        ObjectData() { }
        ~ObjectData();

        QHash<QString,JSONAbstractObject> objects;

        void removeChild(Data* d);
        QString toJSON() const;
        Data* clone() const;
        bool equals( Data* data );
        virtual bool lessThan( const Data* data ) const;

        QString stringify( const QString& str ) const;
    };

    JSONObject(ObjectData* data) : JSONAbstractObject(data) { }

private:
    inline void ensureData() { if ( !m_data ) m_data = new ObjectData(); }
    inline ObjectData* data() const { return static_cast<ObjectData*>(m_data); }
};

#endif // JSONOBJECT_H
