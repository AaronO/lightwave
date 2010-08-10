#include "wavedocument.h"
#include "ot/transformation.h"

#include <QCryptographicHash>

QRegExp* WaveDocument::s_revRegExp = 0;

WaveDocument::WaveDocument(const QString& docId)
    : m_docId(docId)
{
    if ( !s_revRegExp )
        s_revRegExp = new QRegExp("([0-9]+)-([0-9a-f]+)");
}

bool WaveDocument::setContent(FCGI::FCGIRequest* req, JSONObject obj, bool suppressReply)
{    
    //
    // Compute the new version and hash
    //

    QByteArray json = obj.toJSON().toUtf8();
    QCryptographicHash hash(QCryptographicHash::Md5);
    hash.addData(json);
    QString checksum = QString(hash.result());

    if ( !m_rev.isEmpty() )
    {
        if ( !s_revRegExp->exactMatch(m_rev) )
        {
            if ( !suppressReply )
                req->errorReply("The revision is malformed");
            return false;
        }
        m_revNumber = s_revRegExp->cap(1).toInt() + 1;
    }
    else
        m_revNumber = 1;
    m_rev = QString::number(m_revNumber) + "-" + hash.result().toHex();
    obj.setAttribute("_rev", m_rev);

    // Store the new content
    m_json = obj;
    qDebug("DOC: %s", qPrintable(m_json.toJSON()));

    update();
    return true;
}

bool WaveDocument::processMutation(FCGI::FCGIRequest* req, DocumentMutation docop, bool suppressReply)
{
    // The mutation applies to the latest document version?
    if ( docop.revision() != m_rev )
    {
        if ( docop.mutation().isInsertMutation() )
        {
            if ( !suppressReply )
                req->errorReply("When replacing the entire document, you must replace the most recent version");
            return false;
        }

        // TODO: Transform the mutation
        QList<DocumentMutation> serverMutations = getMutations( docop.revision() );
        if ( serverMutations.isEmpty() )
        {
            if ( !suppressReply )
                req->errorReply("The revision " + docop.revision() + " is unknown");
            return false;
        }

        ObjectMutation c = docop.mutation().toObjectMutation();
        foreach( DocumentMutation sdocop, serverMutations )
        {
            Transformation t;
            ObjectMutation s( sdocop.mutation().clone() );
            // If somebody replaced the document in the meantime, then the mutation cannot be applied.
            // We simply build an empty mutation and that's it
            if ( s.isNull() ) // i.e. it is an InsertMutation
            {
                Q_ASSERT( sdocop.mutation().isInsertMutation());
                foreach( QString name, c.toObject().attributeNames())
                {
                    if ( name[0] == '_')
                        continue;
                    c.toObject().removeAttribute(name);
                }
                break;
            }
            qDebug("Transforming\n s: %s\n c: %s", qPrintable(s.toJSON()), qPrintable(c.toJSON()));
            t.xform(s, c);
            if ( t.hasError() )
            {
                if ( !suppressReply )
                    req->errorReply("Error transforming mutation: " + t.errorText());
                return false;
            }
            qDebug("result is\n s: %s\n c: %s", qPrintable(s.toJSON()), qPrintable(c.toJSON()));
        }
    }

    bool ok;
    qDebug("Apply\n %s\nto\n %s", qPrintable(docop.mutation().toJSON()), qPrintable(m_json.toJSON()));
    JSONObject result = docop.apply(m_json, &ok);
    if ( !ok )
    {
        if ( !suppressReply )
            req->errorReply("Could not apply mutation");
        return false;
    }
    qDebug("Result is %s", qPrintable(result.toJSON()));

    // Store the document and that's it
    if ( !setContent(req, result, suppressReply) )
    {
        qDebug("Could not store the content of the changed document");
        return false;
    }

    //
    // Add the mutation to the list
    //

    m_mutations.append(docop);

    // Set the revision at the delta to indicate which revision is yields.
    docop.mutation().toObject().setAttribute("_rev", m_rev);
    return true;
}

QList<DocumentMutation> WaveDocument::getMutations( const QString& sinceRevision )
{
    if ( sinceRevision.isEmpty() )
        return m_mutations;

    if ( !s_revRegExp->exactMatch(sinceRevision) )
    {
        qDebug("Malformed revision number");
        return QList<DocumentMutation>();
    }

    QList<DocumentMutation> result;
    int revNumber = s_revRegExp->cap(1).toInt();
    if ( revNumber >= m_mutations.count() )
    {
        qDebug("Revision number out of range");
        return result;
    }

    for( int i = revNumber; i < m_mutations.count(); ++i )
    {
        result.append(m_mutations.at(i));
    }
    return result;
}

void WaveDocument::update()
{
}
