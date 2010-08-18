#include "viewcontainer.h"
#include "rootcontainer.h"
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
