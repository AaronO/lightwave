#include "hostcontainer.h"
#include "rootcontainer.h"
#include "viewcontainer.h"
#include "view.h"
#include "waveprovider.h"

HostContainer::HostContainer(RootContainer* parent, const QString& name, bool local)
    : WaveContainer(parent, name), m_local(local)
{
}

WaveContainer* HostContainer::createWaveContainer(const QString& name)
{
    WaveContainer* c = this->WaveContainer::createWaveContainer(name);
    if ( !c )
        return 0;
    c->makePersistent();
    foreach(View* v, WaveProvider::self()->viewContainer()->views())
    {
        c->addView(v->documentId(), v->revisionNumber());
    }
    return c;
}
