#include "viewcontainer.h"
#include "rootcontainer.h"
#include "view.h"

ViewContainer::ViewContainer(RootContainer* parent, const QString& name)
    : WaveContainer(parent, name)
{
}

WaveContainer* ViewContainer::createWaveContainer(const QString& name)
{
    Q_ASSERT(childContainer(name) == 0);
    View* v = new View(this, name);
    v->makePersistent();
    return v;
}
