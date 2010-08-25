#include "documentmutation.h"
#include "json/jsonobject.h"
#include "json/jsonarray.h"
#include "objectmutation.h"
#include "liftmutation.h"
#include "arraymutation.h"
#include "deletemutation.h"
#include "skipmutation.h"
#include "textmutation.h"
#include "squeezemutation.h"
#include "insertmutation.h"
#include "richtextmutation.h"
#include "json/jsonconstant.h"

DocumentMutation::DocumentMutation()
{
}

DocumentMutation::DocumentMutation(const DocumentMutation& docmutation)
    : m_mutation(docmutation.m_mutation)
{
}

DocumentMutation::DocumentMutation(const AbstractMutation& mutation)
    : m_mutation(mutation)
{
}

JSONObject DocumentMutation::apply(JSONObject obj, bool* ok)
{
    if ( m_mutation.isNull())
    {
        *ok = true;
        return obj;
    }

    Data data(ok);

    if ( m_mutation.isInsertMutation() )
    {
        if ( !m_mutation.isObject())
        {
            *ok = false;
            return obj;
        }

        check(m_mutation.toInsertMutation(), &data);

        if ( data.hasError() )
            return JSONObject();

        return m_mutation.clone().toObject();
    }
    else if ( m_mutation.isObjectMutation() )
    {
        // Examine the mutation and find out which objects are being lifted, because
        // during application the squeeze may be encountered before its corresponding lift.
        check(obj, m_mutation.toObjectMutation(), &data);

        if ( data.hasError() )
            return JSONObject();

        // Check that each lift has a corresponding squeeze
        foreach( QString name, data.lifts.keys() )
        {
            if ( !data.squeezes.contains(name))
            {
                *ok = false;
                return obj;
            }
        }
        if ( data.lifts.count() != data.squeezes.count() )
        {
            *ok = false;
            return obj;
        }

        return apply(obj, m_mutation.toObjectMutation(), &data).toObject();
    }
    else
    {
        *ok = false;
        return obj;
    }
}

void DocumentMutation::check(JSONAbstractObject dest, AbstractMutation mutation, Data* data)
{
    if ( mutation.isObjectMutation() )
        check(dest, mutation.toObjectMutation(), data);
    else if ( mutation.isArrayMutation() )
        check(dest, mutation.toArrayMutation(), data);
    else if ( mutation.isLiftMutation() )
    {
        data->lifted[ mutation.toLiftMutation().id() ] = dest;
        check(dest, mutation.toLiftMutation(), data );
    }
    else if ( mutation.isTextMutation() )
        check(dest, mutation.toTextMutation(), data);
    else if ( mutation.isRichTextMutation() )
        check(dest, mutation.toRichTextMutation(), data);
    else if ( mutation.isInsertMutation() )
        check( mutation.toInsertMutation(), data);
    else if ( mutation.isSqueezeMutation() )
        check( mutation.toSqueezeMutation(), data);
    else if ( mutation.isSkipMutation() )
    {
        // Do nothing by intention
    }
    else
    {
        data->setError();
    }
}

void DocumentMutation::check(JSONAbstractObject dest, ObjectMutation mutation, Data* data)
{
    if ( !dest.isObject() && !dest.isNull())
    {
        data->setError();
        return;
    }

    JSONObject destobj = dest.toObject();
//    if ( destobj.isNull() )
//    {
//        data->setError();
//        return;
//    }
    JSONObject m = mutation.toObject();
    foreach( QString name, m.attributeNames() )
    {
        if ( name[0] == '$' )
            continue;
        JSONAbstractObject value = destobj.attribute(name);
        AbstractMutation mut( m.attribute(name));
        if ( mut.isObjectMutation() )
            check(value, mut.toObjectMutation(), data);
        else if ( mut.isArrayMutation() )
            check(value, mut.toArrayMutation(), data);
        else if ( mut.isTextMutation() )
            check(value, mut.toTextMutation(), data);
        else if ( mut.isRichTextMutation() )
            check(value, mut.toRichTextMutation(), data);
        else if ( mut.isInsertMutation() )
            check( mut.toInsertMutation(), data);
        else if ( mut.isSkipMutation() )
        {
            // Do nothing by intention
        }
        else
        {
            data->setError();
            return;
        }
    }
}

void DocumentMutation::check(JSONAbstractObject dest, ArrayMutation mutation, Data* data)
{
    if ( !dest.isArray() && !dest.isNull())
    {
        data->setError();
        return;
    }

    JSONArray destarr = dest.toArray();
//    if ( destarr.isNull() )
//    {
//        data->setError();
//        return;
//    }
    int index = 0;
    JSONArray arr = mutation.content();
    for( int i = 0; i < arr.count(); ++i )
    {
        AbstractMutation m = arr[i];
        // Insertions are allowed at the end of the document.
        // TODO: recursion over the insert mutation
        if ( m.isInsertMutation() )
        {
            check( m.toInsertMutation(), data);
            continue;
        }
        if ( m.isSqueezeMutation() )
        {
            check( m.toSqueezeMutation(), data);
            continue;
        }

        // The mutation is larger than the document?
        if ( index >= destarr.count() )
        {
            qDebug("In %i %s\ndoc %s", i, qPrintable(arr.toJSON()), qPrintable(destarr.toJSON()));
            data->setError();
            return;
        }

        if ( m.isDeleteMutation() )
        {
            int count = m.toDeleteMutation().count();
            if ( count <= 0 )
            {
                data->setError();
                return;
            }
            index += count;
        }
        else if ( m.isSkipMutation() )
        {
            int count = m.toSkipMutation().count();
            if ( count <= 0 )
            {
                data->setError();
                return;
            }
            index += count;
        }
        else if ( m.isObjectMutation() )
        {
            check( destarr[index], m.toObjectMutation(), data );
            index++;
        }
        else if ( m.isArrayMutation() )
        {
            check( destarr[index], m.toArrayMutation(), data );
            index++;
        }
        else if ( m.isLiftMutation() )
        {
            check( destarr[index], m.toLiftMutation(), data);
            data->lifted[ m.toLiftMutation().id() ] = destarr[index];
            index++;
        }
        else if ( m.isTextMutation() )
        {
            check( destarr[index], m.toTextMutation(), data );
            index++;
        }
        else if ( m.isRichTextMutation() )
        {
            check( destarr[index], m.toRichTextMutation(), data );
            index++;
        }
        else
        {
            data->setError();
            return;
        }
    }

    // The mutations is smaller than the document?
    if ( index < destarr.count())
    {
        qDebug("destarr: %s", qPrintable(destarr.toJSON()));
        qDebug("mutation: %s", qPrintable(mutation.toJSON()));
        data->setError();
    }
}

void DocumentMutation::check(JSONAbstractObject dest, TextMutation mutation, Data* data)
{
    if ( !dest.isString() && !dest.isNull() )
    {
        data->setError();
        return;
    }

    QString text = dest.toString();
    int index = 0;
    JSONArray arr = mutation.content();
    for( int i = 0; i < arr.count(); ++i )
    {
        AbstractMutation m = arr[i];
        if ( m.isInsertMutation() )
        {
            if ( !m.isString() )
            {
                data->setError();
                return;
            }
            continue;
        }
        // The mutation is too large?
        if ( index >= text.length() )
        {
           data->setError();
           return;
        }
        if ( m.isDeleteMutation() )
        {
            int count = m.toDeleteMutation().count();
            if ( count <= 0 )
            {
                data->setError();
                return;
            }
            index += count;
        }
        else if ( m.isSkipMutation() )
        {
            int count = m.toSkipMutation().count();
            if ( count <= 0 )
            {
                data->setError();
                return;
            }
            index += count;
        }
        else
        {
            // This mutation is not allowed inside a text mutation
            data->setError();
            return;
        }
    }

    // The mutation is too small?
    if ( index < text.length() )
    {
       data->setError();
       return;
    }
}

void DocumentMutation::check(JSONAbstractObject dest, RichTextMutation mutation, Data* data)
{
    if ( !dest.isObject() && !dest.isNull() )
    {
        data->setError();
        return;
    }

    JSONArray destText = dest.toObject().attributeArray("$r");
    int index = 0;
    int inside = 0;

    JSONArray arr = mutation.content();
    for( int i = 0; i < arr.count(); ++i )
    {
        if ( data->hasError() )
            return;

        AbstractMutation m = arr[i];
        if ( m.isInsertMutation() )
            continue;

        if ( m.isDeleteMutation() || m.isSkipMutation() )
        {
            int count = 0;
            if ( m.isDeleteMutation() )
                count = m.toDeleteMutation().count();
            else
                count = m.toSkipMutation().count();
            while( count > 0 )
            {
                // Mutation is longer than the document?
                if ( index >= destText.count() )
                {
                    data->setError();
                    return;
                }
                if ( destText[index].isString() )
                {
                    int len = destText[index].toString().length();
                    int c = qMin(len, count);
                    count -= c;
                    inside += c;
                    if ( inside == len )
                    {
                        index++;
                        inside = 0;
                    }
                }
                else if ( destText[index].toObject().hasAttribute("$format"))
                {
                    // Intentionally do nothing here
                    index++;
                }
                else
                {
                    index++;
                    count--;
                }
            }
            continue;
        }

        // Mutation is longer than the document?
        if ( index >= destText.count() )
        {
            data->setError();
            return;
        }

        if ( destText[index].toObject().hasAttribute("$format"))
        {
            // Intentionally do nothing here
            index++;
            continue;
        }

        if ( m.isObjectMutation() )
        {
            check(destText[index], m.toObjectMutation(), data);
            index++;
        }
        else if ( m.isArrayMutation() )
        {
            check(destText[index], m.toArrayMutation(), data);
            index++;
        }
        else if ( m.isRichTextMutation() )
        {
            check(destText[index], m.toRichTextMutation(), data);
            index++;
        }
        else
        {
            // Operation not allowed here
            data->setError();
            return;
        }
    }

    // The mutation is too small?
    if ( index < destText.count() )
    {
       data->setError();
       return;
    }
}

void DocumentMutation::check(InsertMutation mutation, Data* data)
{
    if ( mutation.isConstant() )
        return;
    if ( mutation.isObject() )
    {
        JSONObject obj = mutation.toObject();
        foreach( QString name, obj.attributeNames() )
        {
            // TODO
            if ( name[0] == '$')
            {
                data->setError();
                return;
            }
            AbstractMutation m( obj.attribute(name));
            if ( m.isInsertMutation() )
            {
                check( m.toInsertMutation(), data );
            }
            else
            {
                data->setError();
                return;
            }
        }
        return;
    }
    if ( mutation.isArray() )
    {
        JSONArray arr = mutation.toArray();
        for( int i = 0; i < arr.count(); ++i )
        {
            AbstractMutation m( arr[i]);
            if ( m.isInsertMutation() )
            {
                check( m.toInsertMutation(), data );
            }
            else
            {
                data->setError();
                return;
            }
        }
        return;
    }

    data->setError();
}

void DocumentMutation::check(SqueezeMutation mutation, Data* data)
{
    if ( data->squeezes.contains(mutation.id()))
    {
        qDebug("ID:%s", qPrintable(mutation.id()));
        data->setError();
        return;
    }
    data->squeezes[mutation.id()] = mutation;
}

void DocumentMutation::check(JSONAbstractObject dest, LiftMutation mutation, Data* data)
{
    if ( dest.isNull() )
    {
        data->setError();
        return;
    }
    if ( data->lifts.contains(mutation.id()))
    {
        data->setError();
        return;
    }
    if ( mutation.hasMutation() )
    {
        AbstractMutation m(mutation.mutation());
        if ( !m.isArrayMutation() && !m.isObjectMutation() && !m.isTextMutation() && !m.isRichTextMutation() )
        {
            data->setError();
            return;
        }
        check(dest, mutation.mutation(), data);
    }
    data->lifts[mutation.id()] = mutation;
}

JSONAbstractObject DocumentMutation::apply(JSONAbstractObject dest, AbstractMutation mutation, Data* data)
{
    if ( mutation.isObjectMutation() )
        return apply(dest, mutation.toObjectMutation(), data);
    if ( mutation.isArrayMutation() )
        return apply(dest, mutation.toArrayMutation(), data);
    if ( mutation.isInsertMutation() )
        return mutation.clone();
//        return apply(mutation.toInsertMutation(), data);
    if ( mutation.isTextMutation() )
        return apply(dest, mutation.toTextMutation(), data);
    if ( mutation.isRichTextMutation() )
        return apply(dest, mutation.toRichTextMutation(), data);
    if ( mutation.isLiftMutation() )
        return JSONAbstractObject();
    if ( mutation.isSqueezeMutation() )
        return data->lifted[mutation.toSqueezeMutation().id()];
    if ( mutation.isDeleteMutation() )
        return JSONAbstractObject();
    if ( mutation.isSkipMutation() )
        return dest;

    return JSONAbstractObject();
}

JSONAbstractObject DocumentMutation::apply(JSONAbstractObject dest, ObjectMutation mutation, Data* data)
{
    JSONObject destobj = dest.toObject();
    JSONObject m = mutation.toObject();
    foreach( QString name, m.attributeNames() )
    {
        if ( name[0] == '$' )
            continue;
        JSONAbstractObject value = destobj.attribute(name);        
        if ( m.attribute(name).isNullValue() )
            destobj.removeAttribute(name);
        else
            destobj.setAttribute(name, apply( value, AbstractMutation(m.attribute(name)), data ) );
    }
    return destobj;
}

JSONAbstractObject DocumentMutation::apply(JSONAbstractObject dest, ArrayMutation mutation, Data* data)
{
    JSONArray destarr = dest.toArray();
    int index = 0;
    JSONArray arr = mutation.content();
    for( int i = 0; i < arr.count(); ++i )
    {
        AbstractMutation m = arr[i];
        if ( m.isDeleteMutation() )
        {
            int count = m.toDeleteMutation().count();
            for( int i = 0; i < count; ++i )
                destarr.removeAt(index);
        }
        else if ( m.isSkipMutation() )
            index += m.toSkipMutation().count();
        else if ( m.isObjectMutation() )
        {
            destarr.replace(index, apply( destarr[index], m.toObjectMutation(), data ) );
            index++;
        }
        else if ( m.isArrayMutation() )
        {
            destarr.replace(index, apply( destarr[index], m.toArrayMutation(), data ) );
            index++;
        }
        else if ( m.isTextMutation() )
        {
            destarr.replace(index, apply( destarr[index], m.toTextMutation(), data ) );
            index++;
        }
        else if ( m.isRichTextMutation() )
        {
            destarr.replace(index, apply( destarr[index], m.toRichTextMutation(), data ) );
            index++;
        }
        else if ( m.isLiftMutation() )
        {
            destarr.removeAt(index);
        }
        else if ( m.isSqueezeMutation() )
        {
            LiftMutation l = data->lifts[m.toSqueezeMutation().id()];
            Q_ASSERT(!l.isNull());
            if ( l.hasMutation() )
                destarr.insert(index++, apply( data->lifted[m.toSqueezeMutation().id()], l.mutation(), data) );
            else
                destarr.insert(index++, data->lifted[m.toSqueezeMutation().id()]);
        }
        else if ( m.isInsertMutation() )
        {
            // destarr.insert(index++, apply( m.toInsertMutation(), data));
            destarr.insert(index++, m.clone());
        }
    }
    return destarr;
}

JSONAbstractObject DocumentMutation::apply(JSONAbstractObject dest, TextMutation mutation, Data* data)
{
    Q_UNUSED(data);

    QString text = dest.toString();
    int index = 0;
    JSONArray arr = mutation.content();
    for( int i = 0; i < arr.count(); ++i )
    {
        AbstractMutation m = arr[i];
        if ( m.isDeleteMutation() )
        {
            int count = m.toDeleteMutation().count();
            text.remove(index, count);
        }
        else if ( m.isSkipMutation() )
            index += m.toSkipMutation().count();
        else if ( m.isInsertMutation() )
        {
            QString t = m.toInsertMutation().toString();
            text.insert(index, t);
            index += t.length();
        }
    }
    return InsertMutation(text);
}

JSONAbstractObject DocumentMutation::apply(JSONAbstractObject dest, RichTextMutation mutation, Data* data)
{
    JSONArray destText = dest.toObject().attributeArray("$r");

    int index = 0;
    int inside = 0;

    JSONArray arr = mutation.content();

    for( int i = 0; i < arr.count(); ++i )
    {        
        AbstractMutation m = arr[i];

        if ( m.isInsertMutation() )
        {
            // Insert characters?
            if ( m.isString() )
            {
                QString t = m.toInsertMutation().toString();
                if ( t.isEmpty())
                    continue;
                // The previous element is a text? (This case includes being at the end of the array)
                if ( inside == 0 && index > 0 && destText[index-1].isString())
                {
                    destText.replace(index-1, JSONConstant( destText[index-1].toString() + t) );
                }
                // The current element is a text?
                else if ( index < destText.count() && destText[index].isString())
                {
                    QString text = destText[index].toString();
                    text.insert(inside, t);
                    inside += t.length();
                    destText.replace(index, JSONConstant(text));
                }
                // The current element is not a text, there is a current element and the previous is not a text either
                else
                {
                    destText.insert(index++, JSONConstant(t));
                }
            }
            // Insert objects?
            else
            {
                // At the beginning of an object or text or at the end of the array?
                if ( inside == 0 )
                    destText.insert(index++, m.clone());
                // Somewhere in some text? -> split it
                else
                {
                    QString text = destText[index].toString();
                    destText.replace(index++, JSONConstant(text.left(inside)));
                    destText.insert(index++, m.clone());
                    text = text.mid(inside);
                    inside = 0;
                    destText.insert(index, JSONConstant(text));
                }
            }
            continue;
        }

        if ( m.isDeleteMutation() )
        {
            int count = m.toDeleteMutation().count();
            while( count > 0 )
            {
                Q_ASSERT(index < destText.count());
                if ( destText[index].isString() )
                {
                    QString text = destText[index].toString();
                    int c = qMin(text.length(), count);
                    count -= c;
                    // Deleted the entire string?
                    if ( text.length() == c )
                    {
                        inside = 0;
                        destText.removeAt(index);
                    }
                    else
                    {
                        text.remove(inside, c);
                        destText.replace(index, JSONConstant(text));
                        // Reached end of string?
                        if ( inside == text.length())
                        {
                            inside = 0;
                            index++;
                        }
                    }
                }
                else
                {
                    destText.removeAt(index);
                    count--;                    
                    // Did this join two text elements?
                    if ( index < destText.count() && index > 0 && destText[index-1].isString() && destText[index].isString() )
                    {
                        QString text = destText[index-1].toString();
                        destText.replace(index-1, JSONConstant( text + destText[index].toString()));
                        destText.removeAt(index);
                        index--;
                        inside = text.length();
                    }
                }
            }
        }
        else if ( m.isSkipMutation() )
        {
            int count = m.toSkipMutation().count();
            while( count > 0 )
            {
                Q_ASSERT(index < destText.count());
                if ( destText[index].isString() )
                {
                    QString text = destText[index].toString();
                    int c = qMin(text.length() - inside, count);
                    inside += c;
                    count -= c;
                    if ( inside == text.length())
                    {
                        inside = 0;
                        index++;
                    }
                }
                else
                {
                    index++;
                    count--;
                }
            }
        }
        else if ( m.isObjectMutation() )
        {
            Q_ASSERT(destText[index].isObject());
            destText.replace(index, apply( destText[index], m.toObjectMutation(), data ) );
            index++;
        }
        else if ( m.isArrayMutation() )
        {
            Q_ASSERT(destText[index].isArray());
            destText.replace(index, apply( destText[index], m.toArrayMutation(), data ) );
            index++;
        }
        else if ( m.isRichTextMutation() )
        {
            Q_ASSERT(destText[index].isObject());
            destText.replace(index, apply( destText[index], m.toRichTextMutation(), data ) );
            index++;
        }
    }

    return dest;
}

QString DocumentMutation::documentId() const
{
    return m_mutation.toObject().attributeString("_id");
}

void DocumentMutation::setDocumentId( const QString& docId )
{
    m_mutation.toObject().setAttribute("_id", docId);
}

QString DocumentMutation::revision() const
{
    return m_mutation.toObject().attributeString("_rev");
}

void DocumentMutation::setRevision( const QString& rev )
{
    m_mutation.toObject().setAttribute("_rev", rev);
}

QString DocumentMutation::author() const
{
    return m_mutation.toObject().attributeString("_author");
}

void DocumentMutation::setAuthor( const QString& author )
{
    m_mutation.toObject().setAttribute("_author", author);
}

int DocumentMutation::revisionNumber() const
{
    QString rev = m_mutation.toObject().attributeString("_rev");
    int index = rev.indexOf('-');
    if ( index == -1 )
        return 0;
    return rev.left(index).toInt();
}
