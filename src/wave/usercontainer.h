#ifndef USERCONTAINER_H
#define USERCONTAINER_H

#include "wavecontainer.h"

class RootContainer;
class WaveProvider;

class UserContainer : public WaveContainer
{
public:
    UserContainer(RootContainer* parent, const QString& name);

    virtual bool isRemote() const { return false; }

protected:
    friend class WaveProvider;

    virtual WaveContainer* createWaveContainer(const QString& name);
};

#endif // USERCONTAINER_H
