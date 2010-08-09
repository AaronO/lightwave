#include "randomdocgenerator.h"
#include "json/jsonarray.h"
#include "json/jsonconstant.h"
#include <QtGlobal>

RandomDocGenerator::RandomDocGenerator(int depth)
{
    m_depth = depth;
}

JSONObject RandomDocGenerator::createObject(int depth)
{
    JSONObject obj(true);

    // How many attributes should it have?
    int x = qrand() % 4;
    for( int i = 0; i < x; ++i )
    {
        QString name = "a" + QString::number(i);
        // Which type should it have?
        int y = qrand() % 8;
        if ( depth < m_depth && ( y == 0 || y == 1 || y == 2 || y == 3) )
            obj.setAttribute(name, createArray(depth + 1));
        else if ( depth < m_depth && y == 4 )
            obj.setAttribute(name, createObject(depth + 1));
        else if ( depth < m_depth && y == 5 )
            obj.setAttribute(name, createRichText(depth + 1));
        else
            obj.setAttribute(name, createString());
    }

    return obj;
}

QString RandomDocGenerator::createString()
{
    QString str = "";
    int x = qrand() % 4;
    for( int i = 0; i < x; ++i )
    {
        str += QChar('a' + qrand() % 26);
    }
    return str;
}

JSONArray RandomDocGenerator::createArray(int depth)
{
    JSONArray arr(true);
    int x = qrand() % 12;
    for( int i = 0; i < x; ++i )
    {
        // Which type should it have?
        int y = qrand() % 8;
        if ( depth < m_depth && ( y == 0 || y == 1 || y == 2 || y == 3) )
            arr.append(createArray(depth + 1));
        else if ( depth < m_depth && y == 4)
            arr.append(createObject(depth + 1));
        else if ( depth < m_depth && y == 5)
            arr.append(createRichText(depth + 1));
        else
            arr.append(createString());
    }
    return arr;
}

JSONObject RandomDocGenerator::createRichText(int depth)
{
    JSONObject obj(true);
    JSONArray arr(true);

    // How many objects should it have?
    int x = qrand() % 8;
    bool last_was_text = false;
    for( int i = 0; i < x; ++i )
    {
        // Which type should it have?
        int y = qrand() % 6;
        if ( !last_was_text && ( y == 0 || y == 1 || y == 2 || y == 3) )
        {
            last_was_text = true;
            arr.append( JSONConstant("r" + createString()));
        }
        else if ( y == 4 )
        {
            last_was_text = false;
            arr.append(createObject(depth + 1));
        }
        else if ( y == 5 )
        {
            last_was_text = false;
            arr.append(createRichText(depth + 1));
        }
        else
        {
            last_was_text = false;
            arr.append(createArray(depth + 1));
        }
    }   
    obj.setAttribute("_r", arr);

    return obj;
}
