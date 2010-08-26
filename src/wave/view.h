#ifndef VIEW_H
#define VIEW_H

#include "viewcontainer.h"
#include "json/jsonabstractobject.h"
#include "json/jsonarray.h"
#include <QScriptValue>
#include <QHash>
#include <QStringList>

class View : public WaveDocument
{
public:
    View(ViewContainer* parent, const QString& docId);
    ~View();

    void update();

    bool isMalformed() const { return m_malformed; }

    ViewContainer* viewContainer() const { return static_cast<ViewContainer*>(parent()); }

    void updateDigest(WaveContainer* c);

    class Index
    {
    public:
        Index(const QScriptValue& func) : m_mapFunction(func) { };

        QScriptValue mapFunction() const { return m_mapFunction; }

    private:
        QScriptValue m_mapFunction;
    };

    class Query
    {
    public:
        Query() { }
        Query( const QString& sessionId, const QString& userJID ) : m_sessionId(sessionId), m_user(userJID) { }
        Query( const Query& q ) : m_sessionId(q.m_sessionId), m_user( q.m_user ) { }

        QString userJID() const { return m_user; }
        QString sessionID() const { return m_sessionId; }

    private:
        QString m_sessionId;
        QString m_user;
    };

    class IndexItem
    {
    public:
        IndexItem() { }
        IndexItem( const IndexItem& item ) : m_key( item.m_key ), m_value( item.m_value ) { }
        IndexItem( const JSONArray& key, const JSONAbstractObject& value ) : m_key( key ), m_value( value ) { }

        JSONArray key() const { return m_key; }
        JSONAbstractObject value() const { return m_value; }

    private:
        JSONArray m_key;
        JSONAbstractObject m_value;
    };

    typedef QList<IndexItem> IndexItemList;

    Index* index(const QString& name) { return m_indices.value(name); }
    QStringList indexNames() const;

    /**
      * @return a queryID.
      */
    QString registerSessionQuery( const Query& query );
    void notifySessionQueries( QHash<QString,IndexItemList> newIndexItems );

private:
    void clearIndices();
    void computeDigestMap(WaveContainer* c);
    void computeDigestReduce(WaveContainer* c);
    void updateDigestReduce(WaveContainer* c);
    QScriptValue parseFunction(const QString& js);

    /**
      * This variable is true if the JSON object that defines this view is malformed.
      */
    bool m_malformed;
    QScriptValue m_digestMapFunction;
    QScriptValue m_digestReduceFunction;
    /**
      * The key is the name of the index.
      */
    QHash<QString,Index*> m_indices;
    /**
      * The key is the queryId.
      */
    QHash<QString,Query> m_sessionQueries;
    /**
      * The key is the dataBaseDocumentId which generated the item.
      */
    QHash<QString,IndexItemList> m_indexItems;
    /**
      * The key is the waveid of the container for which the digest has been created.
      */
    QHash<QString,QScriptValue> m_digestMap;
    /**
      * The key is the waveid of the container for which the digest has been created and reduced.
      */
    QHash<QString,QScriptValue> m_digestReduce;
};

#endif // VIEW_H
