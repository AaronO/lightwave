#include "jsonscanner.h"
#include "jsonconstant.h"
#include <ctype.h>
#include <string.h>
#include <QByteArray>
#include <QString>
#include <QtGlobal>

JSONScanner::Token JSONScanner::next()
{
    m_revertPtr = m_ptr;
    m_revertLen = m_len;

    while( true )
    {
        if ( m_len == 0 )
            return End;
        char ch = *m_ptr++;
        m_len--;
        if ( ch == '"' )
        {
            m_value = m_ptr;
            bool backslash = false;
            while( m_len > 0 )
            {
                char ch = *m_ptr;
                if ( backslash )
                {
                    backslash = false;
                    m_len--;
                    m_ptr++;
                }
                else if ( ch == '"' )
                {
                    m_len--;
                    m_ptr++;
                    break;
                }
                if ( ch == '\\' )
                    backslash = true;
                m_len--;
                m_ptr++;
            }
            if ( m_len == 0 )
                return Error;
            m_valueLen = (m_ptr - m_value - 1);
            return StringValue;
        }
        else if ( isdigit(ch) != 0 || ch == '-' || ch == '+' || ch == '.' )
        {
            m_value = m_ptr - 1;
            while( m_len > 0 )
            {
                char ch = *m_ptr;
                if ( isalnum(ch) != 0 || ch == '+' || ch == '-' || ch == '.' )
                {
                    m_len--;
                    m_ptr++;
                }
                else
                    break;
            }
            if ( m_len == 0 )
                return Error;
            m_valueLen = (int)(m_ptr - m_value);
            return NumberValue;
        }
        else if ( ch == 't' || ch == 'T' )
        {
            if ( m_len > 3 && strncmp( m_ptr, "rue", 3 ) == 0 )
            {
                m_ptr += 3;
                m_len -= 3;
                return TrueValue;
            }
            return Error;
        }
        else if ( ch == 'f' || ch == 'F' )
        {
            if ( m_len > 4 && strncmp( m_ptr, "alse", 4 ) == 0 )
            {
                m_ptr += 4;
                m_len -= 4;
                return FalseValue;
            }
            return Error;
        }
        else if ( ch == 'n' )
        {
            if ( m_len > 3 && strncmp( m_ptr, "ull", 3 ) == 0 )
            {
                m_ptr += 3;
                m_len -= 3;
                return NullValue;
            }
            return Error;
        }
        else if ( ch == '{' )
            return BeginObject;
        else if ( ch == '}' )
            return EndObject;
        else if ( ch == '[' )
            return BeginArray;
        else if ( ch == ']' )
            return EndArray;
        else if ( ch == ',' )
            return Comma;
        else if ( ch == ':' )
            return Colon;
        else if ( isspace(ch) != 0 )
            continue;
        return Error;
    }
}

QString JSONScanner::stringValue(bool* ok)
{
    QByteArray ba = QByteArray::fromRawData( m_value, m_valueLen );
    return unescape( ba, ok );
}

double JSONScanner::doubleValue(bool *ok)
{
    return QByteArray::fromRawData( m_value, m_valueLen ).toDouble(ok);
}

float JSONScanner::floatValue(bool *ok)
{
    return QByteArray::fromRawData( m_value, m_valueLen ).toFloat(ok);
}

qint32 JSONScanner::int32Value(bool *ok)
{
    return QByteArray::fromRawData( m_value, m_valueLen ).toInt(ok);
}

quint32 JSONScanner::uint32Value(bool *ok)
{
    return QByteArray::fromRawData( m_value, m_valueLen ).toUInt(ok);
}

qint64 JSONScanner::int64Value(bool *ok)
{
    return QByteArray::fromRawData( m_value, m_valueLen ).toLongLong(ok);
}

quint64 JSONScanner::uint64Value(bool *ok)
{
    return QByteArray::fromRawData( m_value, m_valueLen ).toULongLong(ok);
}

int JSONScanner::enumValue(bool *ok)
{
    return QByteArray::fromRawData( m_value, m_valueLen ).toInt(ok);
}

char JSONScanner::byteValue(bool *ok)
{
    int v = QByteArray::fromRawData( m_value, m_valueLen ).toInt(ok);
    if ( v < -128 || v > 128 )
        *ok = false;
    return (char)v;
}

int JSONScanner::tagValue()
{
    bool ok;
    int tag = QByteArray::fromRawData( m_value, m_valueLen ).toInt(&ok);
    if ( !ok )
        return -1;
    return tag;
}

bool JSONScanner::ishexnstring(const QString& string)
{
    for (int i = 0; i < string.length(); i++)
    {
        if (isxdigit(string[i] == 0))
            return false;
    }
    return true;
}

QString JSONScanner::unescape( const QByteArray& ba, bool* ok )
{
    Q_ASSERT( ok );
    *ok = false;
    QString res;
    QByteArray seg;
    bool bs = false;
    for ( int i = 0, size = ba.size(); i < size; ++i )
    {
        const char ch = ba[i];
        if ( !bs )
        {
            if ( ch == '\\' )
                bs = true;
            else
                seg += ch;
        }
        else
        {
            bs = false;
            switch ( ch )
            {
            case 'b':
                seg += '\b';
                break;
            case 'f':
                seg += '\f';
                break;
            case 'n':
                seg += '\n';
                break;
            case 'r':
                seg += '\r';
                break;
            case 't':
                seg += '\t';
                break;
            case 'u':
                {
                    res += QString::fromUtf8( seg );
                    seg.clear();

                    if ( i > size - 5 )
                    {
                        //error
                        return QString();
                    }

                    const QString hex_digit1 = QString::fromUtf8( ba.mid( i + 1, 2 ) );
                    const QString hex_digit2 = QString::fromUtf8( ba.mid( i + 3, 2 ) );
                    i += 4;

                    if ( !ishexnstring( hex_digit1 ) || !ishexnstring( hex_digit2 ) )
                        return QString::null;
                    bool hexOk;
                    const ushort hex_code1 = hex_digit1.toShort( &hexOk, 16 );
                    if (!hexOk)
                        return QString::null;
                    const ushort hex_code2 = hex_digit2.toShort( &hexOk, 16 );
                    if (!hexOk)
                        return QString::null;

                    res += QChar(hex_code2, hex_code1);
                    break;
                }
            case '\\':
                seg  += '\\';
                break;
            default:
                seg += ch;
                break;
            }
        }
    }
    res += QString::fromUtf8( seg );
    *ok = true;
    return res;
}

void JSONScanner::revert()
{
    m_ptr = m_revertPtr;
    m_len = m_revertLen;
}

JSONObject JSONScanner::scan(bool *ok)
{
    JSONObject obj = scanObject(ok);
    if ( !ok )
        return JSONObject();
    if ( next() != End )
    {
        *ok = false;
        return JSONObject();
    }
    return obj;
}

JSONArray JSONScanner::scanArray(bool *ok)
{
    *ok = false;
    if ( next() != BeginArray )
        return JSONArray();

    JSONArray arr(true);
    Token t;
    bool ok2;
    while( ( t = next()  ) != EndArray )
    {
        switch( t )
        {
        case StringValue:
            {
                QString val = stringValue(&ok2);
                if ( !ok2 )
                    return JSONArray();
                arr.append(val);
            }
            break;
        case NumberValue:
            {
                double val = doubleValue(&ok2);
                if ( !ok2 )
                    return JSONArray();
                arr.append(val);
            }
            break;
        case NullValue:
            arr.append( JSONConstant::createNull() );
            break;
        case TrueValue:
            arr.append(true);
            break;
        case FalseValue:
            arr.append(false);
            break;
        case BeginObject:
            revert();
            arr.append(scanObject(&ok2));
            if ( !ok2 )
                return JSONArray();
            break;
        case BeginArray:
            revert();
            arr.append(scanArray(&ok2));
            if ( !ok2 )
                return JSONArray();
            break;
        default:
            return JSONArray();
        }

        t = next();
        if ( t == EndArray )
            revert();
        else if ( t != Comma )
            return JSONArray();
    }

    *ok = true;
    return arr;
}

JSONObject JSONScanner::scanObject(bool *ok)
{
    *ok = false;
    if ( next() != BeginObject )
        return JSONObject();
    JSONObject obj(true);
    Token t;
    while( ( t = next() ) != EndObject )
    {
        if ( t != StringValue )
            return JSONObject();
        bool ok2;
        QString key = stringValue(&ok2);
        if ( !ok2 )
            return JSONObject();
        t = next();
        if ( t != Colon )
            return JSONObject();
        t = next();
        switch( t )
        {
        case StringValue:
            {
                QString val = stringValue(&ok2);
                if ( !ok2 )
                    return JSONObject();
                obj.setAttribute(key, val);
            }
            break;
        case NumberValue:
            {
                double val = doubleValue(&ok2);
                if ( !ok2 )
                    return JSONObject();
                obj.setAttribute(key, val);
            }
            break;
        case TrueValue:
            obj.setAttribute(key, true);
            break;
        case FalseValue:
            obj.setAttribute(key, false);
            break;
        case NullValue:
            obj.setAttribute(key, JSONConstant::createNull());
            break;
        case BeginObject:
            revert();
            obj.setAttribute(key, scanObject(&ok2));
            if ( !ok2 )
                return JSONObject();
            break;
        case BeginArray:
            revert();
            obj.setAttribute(key, scanArray(&ok2));
            if ( !ok2 )
                return JSONObject();
            break;
        default:
            return JSONObject();
        }
        // qDebug("%s", qPrintable(obj.toJSON()));

        t = next();
        if ( t == EndObject )
            revert();
        else if ( t != Comma )
            return JSONObject();
    }

    *ok = true;
    return obj;
}
