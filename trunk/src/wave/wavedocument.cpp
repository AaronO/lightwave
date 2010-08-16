#include "wavedocument.h"
#include "wavecontainer.h"
#include "hostcontainer.h"
#include "rootcontainer.h"
#include "sessioncontainer.h"
#include "ot/transformation.h"
#include "session.h"
#include <QCryptographicHash>

QRegExp* WaveDocument::s_revRegExp = 0;

WaveDocument::WaveDocument(WaveContainer* parent, const QString& docId)
    : QObject(parent), m_docId(docId), m_revNumber(0), m_revNumberOffset(0)
{
    if ( !s_revRegExp )
        s_revRegExp = new QRegExp("([0-9]+)-([0-9a-f]+)");
}

bool WaveDocument::setSnapshotFromHost( JSONObject doc )
{
    Q_ASSERT(container() && container()->isRemote());
    Q_ASSERT(!doc.attribute("_snapshot").isNull() && doc.attribute("_snapshot").toBool() == true);

    doc.removeAttribute("_snapshot");
    QString rev = doc.attributeString("_rev");
    if ( !s_revRegExp->exactMatch(rev) )
    {
        qDebug("Error: The revision is malformed");
        return false;
    }
    if ( s_revRegExp->cap(1).toInt() < 1 )
    {
        qDebug("Snapshot must have a version numer >= 1");
        return false;
    }
    m_revNumber = s_revRegExp->cap(1).toInt();
    m_revNumberOffset = m_revNumber - 1;
    m_rev = rev;

    // Store the snapshot as the first mutation
    AbstractMutation m(doc);
    DocumentMutation docop(m);
    m_mutations.clear();
    m_mutations.append(docop);

    // Store the new content
    m_json = doc;

    return true;
}

bool WaveDocument::setContent(JSONObject obj)
{    
    Q_ASSERT(session() || !container()->isRemote());

    // Compute the new version and hash
    QByteArray json = obj.toJSON().toUtf8();
    QCryptographicHash hash(QCryptographicHash::Md5);
    hash.addData(json);
    QString checksum = QString(hash.result());

    if ( !m_rev.isEmpty() )
    {
        if ( !s_revRegExp->exactMatch(m_rev) )
        {
            qDebug("Error: The revision is malformed");
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

    return true;
}

bool WaveDocument::processMutation(DocumentMutation docop)
{
    Q_ASSERT(session() || !container()->isRemote());

    // The mutation applies to the latest document version?
    if ( docop.revision() != m_rev )
    {
        if ( docop.mutation().isInsertMutation() )
        {
            qDebug("Error: When replacing the entire document, you must replace the most recent version");
            return false;
        }

        // TODO: Transform the mutation
        QList<DocumentMutation> serverMutations = getMutations( docop.revision() );
        if ( serverMutations.isEmpty() )
        {
            qDebug("The revision %s is unknown", qPrintable(docop.revision()) );
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
                qDebug("Error: transforming mutation: %s", qPrintable(t.errorText()));
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
        qDebug("Error: Could not apply mutation");
        return false;
    }
    qDebug("Result is %s", qPrintable(result.toJSON()));

    // Store the document and that's it
    if ( !setContent(result) )
    {
        qDebug("Could not store the content of the changed document");
        return false;
    }

    // Set the revision at the delta to indicate which revision is yields, i.e. the revision AFTER the mutation has been applied.
    docop.mutation().toObject().setAttribute("_rev", m_rev);

    // Add the mutation to the list
    m_mutations.append(docop);

    return true;
}

bool WaveDocument::processMutationFromHost(DocumentMutation docop)
{
    Q_ASSERT(container() && container()->isRemote());
    Q_ASSERT(docop.mutation().toObject().attribute("_snapshot").isNull());

    if ( !s_revRegExp->exactMatch(docop.revision()) )
    {
        qDebug("Error: The revision is malformed");
        return false;
    }
    if ( m_revNumber + 1 != s_revRegExp->cap(1).toInt() )
    {
        qDebug("Error: The mutation must advance the current revision number by 1.");
        return false;
    }
    int num = s_revRegExp->cap(1).toInt();

    bool ok;
    qDebug("Apply\n %s\nto\n %s", qPrintable(docop.mutation().toJSON()), qPrintable(m_json.toJSON()));
    JSONObject result = docop.apply(m_json, &ok);
    if ( !ok )
    {
        qDebug("Error: Could not apply mutation");
        return false;
    }
    qDebug("Result is %s", qPrintable(result.toJSON()));

    // Store the document and that's it
    m_revNumber = num;
    m_rev = docop.revision();
    result.setAttribute("_rev", m_rev);
    m_json = result;

    // Add the mutation to the list
    m_mutations.append(docop);

    return true;
}

QList<DocumentMutation> WaveDocument::getMutations( const QString& sinceRevision )
{
    // All mutations?
    if ( sinceRevision.isEmpty() )
        return m_mutations;

    if ( !s_revRegExp->exactMatch(sinceRevision) )
    {
        qDebug("Malformed revision number");
        return QList<DocumentMutation>();
    }

    QList<DocumentMutation> result;
    int revNumber = s_revRegExp->cap(1).toInt();
    if ( revNumber >= m_revNumberOffset + m_mutations.count() || revNumber < m_revNumberOffset )
    {
        qDebug("Revision number out of range");
        return result;
    }

    for( int i = revNumber; i < m_mutations.count(); ++i )
    {
        result.append(m_mutations.at(i - m_revNumberOffset));
    }
    return result;
}

WaveId WaveDocument::waveId() const
{
    if ( dynamic_cast<const HostContainer*>(container()))
        return WaveId( container()->name(), QStringList(), m_docId);
    if ( dynamic_cast<const SessionContainer*>(container()))
        return WaveId( container()->name(), QStringList(), m_docId);
    if ( dynamic_cast<const RootContainer*>(container()))
        return WaveId( QString::null, QStringList(), m_docId);
    
    QStringList pathItems;
    WaveContainer* c = container();
    QString host;
    while(c)
    {
        if ( dynamic_cast<HostContainer*>(c) )
        {
            host = c->name();
            break;
        }
        if ( dynamic_cast<SessionContainer*>(c) )
        {
            host = "_session";
            break;
        }
        pathItems.prepend(c->name());
        c = c->parentContainer();
    }
    Q_ASSERT(c);

    return WaveId( host, pathItems, m_docId);
}
