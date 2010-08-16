#include "sessioncontainer.h"
#include "rootcontainer.h"

SessionContainer::SessionContainer(RootContainer* parent, const QString& name)
    : WaveContainer(parent, name)
{
}
