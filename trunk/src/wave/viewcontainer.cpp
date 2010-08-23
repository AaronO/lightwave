#include "viewcontainer.h"
#include "rootcontainer.h"
// #include "sessioncontainer.h"
#include "waveprovider.h"
#include "view.h"

ViewContainer::ViewContainer(RootContainer* parent, const QString& name)
    : WaveContainer(parent, name)
{
}

WaveDocument* ViewContainer::createDocument(const QString& docId)
{
    return new View(this, docId);
}

void ViewContainer::onDocumentUpdate(WaveDocument* wdoc)
{
    View* v = dynamic_cast<View*>(wdoc);
    if ( !v )
        return;

    v->update();
}

QList<View*> ViewContainer::views() const
{
    QList<View*> result;
    foreach( WaveDocument* doc, documents() )
    {
        if ( doc == metaDocument() )
            continue;
        Q_ASSERT(dynamic_cast<View*>(doc) != 0);
        result.append(static_cast<View*>(doc));
    }
    return result;
}

WaveContainer* ViewContainer::createWaveContainer(const QString& name)
{
    Q_UNUSED(name);

    return 0;
}
