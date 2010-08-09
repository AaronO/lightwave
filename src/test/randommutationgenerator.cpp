#include "randommutationgenerator.h"
#include "ot/objectmutation.h"
#include "ot/arraymutation.h"
#include "ot/textmutation.h"
#include "ot/insertmutation.h"
#include "ot/skipmutation.h"
#include "ot/deletemutation.h"
#include "json/jsonobject.h"
#include "json/jsonconstant.h"
#include "ot/documentmutation.h"
#include "ot/liftmutation.h"
#include "ot/squeezemutation.h"
#include "ot/richtextmutation.h"
#include <QtGlobal>
#include <QSet>

RandomMutationGenerator::RandomMutationGenerator(int lifts)
    : m_lifts(lifts)
{
}

DocumentMutation RandomMutationGenerator::createDocumentMutation(JSONObject obj)
{
    m_liftCount = 0;

    ObjectMutation m = createMutation(obj);

    return DocumentMutation(m);
}

AbstractMutation RandomMutationGenerator::createMutation(JSONAbstractObject obj)
{
    if ( obj.isObject())
    {
        if ( obj.toObject().hasAttribute("_r") )
            return createRichTextMutation(obj.toObject());
        else
            return createMutation(obj.toObject());
    }
    if ( obj.isArray())
        return createMutation(obj.toArray());
    if ( obj.isString() )
        return createMutation(obj.toString());
    qFatal("Unexpected value in document");
    return AbstractMutation();
}

ObjectMutation RandomMutationGenerator::createMutation(JSONObject obj)
{
    ObjectMutation m(true);
    QList<QString> names = obj.attributeNames();
    if ( names.count() == 0)
        return m;
    int x = qrand() % names.count();
    for( int i = 0; i < x; ++ i)
    {
        if ( names[i][0] == 'l')
            continue;

        if ( qrand() % 4 == 0 )
            m.setMutation(names[i], InsertMutation( JSONConstant::createNull() ));
        else if ( obj.attribute(names[i]).isString())
        {
            if ( qrand() % 2 == 0 )
                m.setMutation(names[i], createMutation(obj.attribute(names[i])));
            else
                m.setMutation(names[i], InsertMutation( createString() ));
        }
        else
        {
            m.setMutation(names[i], createMutation(obj.attribute(names[i])));
        }
    }

    if ( qrand() % 4 == 0 )
        m.setMutation("s", InsertMutation( createString() ));

    return m;
}

ArrayMutation RandomMutationGenerator::createMutation(JSONArray arr)
{
    ArrayMutation m(true);

    QList<int> liftPositions;
    QList<int> arrPositions;

    int i = 0;
    while( i < arr.count() )
    {
        int y = qrand() % 7;
        if ( y == 0 || y == 1 )
        {
            liftPositions.append( m.content().count() );
            arrPositions.append(i);
            m.content().append(SkipMutation(1));
            i++;
        }
        else if ( y == 2 || y == 3 )
        {
            liftPositions.append( m.content().count() );
            arrPositions.append(i);
            m.content().append(DeleteMutation(1));
            i++;
        }
        else if ( y == 4 || y == 5)
        {
            liftPositions.append( m.content().count() );
            arrPositions.append(i);
            m.content().append( createMutation(arr[i]));
            i++;
        }
        else
        {
            m.content().append( InsertMutation(createString()) );
        }
    }

    int counter = m_liftCount;
    int lifts = qrand() % qMax(1, liftPositions.count() - 2);
    m_liftCount += lifts;
//    qDebug("Fuck %i", lifts);
    for( int i = 0; i < lifts; i++ )
    {        
        int pos = qrand() % liftPositions.count();
        LiftMutation l( "AL" + QString::number(counter+i));
        m.content().replace(liftPositions[pos], l);

        if ( qrand() % 2 == 0 )
        {
            l.setMutation( createMutation(arr[arrPositions[pos]]) );
        }

        liftPositions.removeAt(pos);
        arrPositions.removeAt(pos);
    }

    for( int i = 0; i < lifts; i++ )
    {
        int pos =  qrand() % (m.content().count() + 1);
        m.content().insert(pos, SqueezeMutation( "AL" + QString::number(counter+i)));
    }

    return m;
}

TextMutation RandomMutationGenerator::createMutation(const QString& str)
{
    TextMutation m(true);

    int i = 0;
    while( i < str.count() )
    {
//        int y = qrand() % 5;
//        if ( y == 0 || y == 1 )
//        {
//            m.content().append(SkipMutation(1));
//            i++;
//        }
//        else if ( y == 2 || y == 3 )
//        {
//            m.content().append(DeleteMutation(1));
//            i++;
//        }
//        else
//            m.content().append( InsertMutation(createString()) );
        int y = qrand() % 3;
        if ( y == 0 )
        {
            m.content().append(SkipMutation(1));
            i++;
        }
        else if ( y == 1)
        {
            m.content().append(DeleteMutation(1));
            i++;
        }
        else
            m.content().append( InsertMutation(createString()) );
    }

    return m;
}

RichTextMutation RandomMutationGenerator::createRichTextMutation(JSONObject object)
{
    RichTextMutation m(true);

    JSONArray text = object.attributeArray("_r");
    // How many characters/objects are there?
    int count = 0;
    QHash<int, JSONAbstractObject> objects;
    for( int j = 0; j < text.count(); ++j )
    {
        if ( text[j].isString())
            count += text[j].toString().length();
        else if ( !text[j].toObject().hasAttribute("_format"))
        {
            objects[count] = text[j];
            count++;
        }
    }

    int i = 0;
    while( i < count )
    {
        int y = qrand() % 4;
        if ( y == 0 )
        {
            m.content().append(SkipMutation(1));
            i++;
        }
        else if ( y == 1)
        {
            m.content().append(DeleteMutation(1));
            i++;
        }
        else if ( y == 2 && objects.contains(i))
        {
            m.content().append( createMutation( objects[i]) );
            i++;
        }
        else
        {
            m.content().append( InsertMutation(createString()) );
        }
    }

    return m;
}

QString RandomMutationGenerator::createString()
{
    QString str = "";
    int x = qrand() % 4;
    for( int i = 0; i < x; ++i )
    {
        str += QChar('a' + qrand() % 26);
    }
    return str;
}
