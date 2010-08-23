#ifndef VIEW_H
#define VIEW_H

#include "viewcontainer.h"
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

    QScriptValue digestMapFunction() const { return m_digestMapFunction; }
    QScriptValue digestReduceFunction() const { return m_digestReduceFunction; }

    QScriptValue computeDigestMap(WaveContainer* c);
    QScriptValue computeDigestReduce(WaveContainer* c);

    class Index
    {
    public:
        Index(const QScriptValue& func) : m_mapFunction(func) { };

        QScriptValue mapFunction() const { return m_mapFunction; }

    private:
        QScriptValue m_mapFunction;
    };

    Index* index(const QString& name) { return m_indices.value(name); }
    QStringList indexNames() const;

private:
    void clearIndices();

    QScriptValue parseFunction(const QString& js);

    bool m_malformed;
    QScriptValue m_digestMapFunction;
    QScriptValue m_digestReduceFunction;
    QHash<QString,Index*> m_indices;
};

#endif // VIEW_H
