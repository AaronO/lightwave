#include "hostcontainer.h"
#include "rootcontainer.h"

HostContainer::HostContainer(RootContainer* parent, const QString& name, bool local)
    : WaveContainer(parent, name), m_local(local)
{
}
