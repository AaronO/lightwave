#ifndef USERCONTAINER_H
#define USERCONTAINER_H

#include "wavecontainer.h"

class RootContainer;

class UserContainer : public WaveContainer
{
public:
    UserContainer(RootContainer* parent, const QString& name);

    virtual bool isRemote() const { return false; }

protected:
    virtual WaveContainer* createWaveContainer(const QString& name);
};

#endif // USERCONTAINER_H
