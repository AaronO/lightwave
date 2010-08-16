#ifndef HOSTCONTAINER_H
#define HOSTCONTAINER_H

#include "wavecontainer.h"

class RootContainer;

class HostContainer : public WaveContainer
{
public:
    HostContainer(RootContainer* parent, const QString& name, bool local);

    bool isLocal() const { return m_local; }
    bool isRemote() const { return !m_local; }

private:
    bool m_local;
};

#endif // HOSTCONTAINER_H
