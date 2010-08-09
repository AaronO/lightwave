#ifndef JSONSCANNER_H
#define JSONSCANNER_H

#include <QString>
#include <QtGlobal>
#include "jsonobject.h"
#include "jsonarray.h"

class JSONScanner
{
public:
    enum Token
    {
        BeginObject = 1,
        EndObject = 2,
        BeginArray = 3,
        EndArray = 4,
        Comma = 5,
        TrueValue = 6,
        FalseValue = 7,
        StringValue = 8,
        NumberValue = 9,
        Colon = 10,
        NullValue = 11,
        End = 0,
        Error = 1000
    };

    JSONScanner(const char* ptr, int len) { m_ptr = ptr, m_len = len; }

    JSONObject scan(bool *ok);

    Token next();
    /**
      * Undos the previous next() call, but only once.
      */
    void revert();

    QString stringValue(bool *ok);
    double doubleValue(bool *ok);
    float floatValue(bool *ok);
    qint32 int32Value(bool *ok);
    quint32 uint32Value(bool *ok);
    qint64 int64Value(bool *ok);
    quint64 uint64Value(bool *ok);
    int enumValue(bool *ok);
    int tagValue();
    char byteValue(bool *ok);

private:    
    bool ishexnstring(const QString& string);
    QString unescape( const QByteArray& ba, bool* ok );

    JSONObject scanObject(bool *ok);
    JSONArray scanArray(bool *ok);

    const char* m_ptr;
    int m_len;
    const char* m_value;
    int m_valueLen;
    const char* m_revertPtr;
    int m_revertLen;
};

#endif // JSONSCANNER_H
