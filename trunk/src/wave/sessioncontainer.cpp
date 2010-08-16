#include "sessioncontainer.h"
#include "rootcontainer.h"
#include "session.h"

SessionContainer::SessionContainer(RootContainer* parent, const QString& name)
    : WaveContainer(parent, name)
{
}

WaveContainer* SessionContainer::createWaveContainer(const QString& name)
{
    Q_ASSERT(childContainer(name) == 0);
    Session* s = new Session(this, name);
    return s;
}
