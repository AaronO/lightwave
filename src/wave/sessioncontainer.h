#ifndef SESSIONCONTAINER_H
#define SESSIONCONTAINER_H

#include "wavecontainer.h"
#include <QString>

class RootContainer;

class SessionContainer : public WaveContainer
{
public:
    SessionContainer(RootContainer* parent, const QString& name);

    bool isRemote() const { return false; }
    bool buildsDigest() const { return false; }

protected:
    virtual WaveContainer* createWaveContainer(const QString& name);
};

#endif // SESSIONCONTAINER_H
