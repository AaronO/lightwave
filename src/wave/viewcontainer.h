#ifndef VIEWCONTAINER_H
#define VIEWCONTAINER_H

#include "wavecontainer.h"
#include <QString>

class RootContainer;

class ViewContainer : public WaveContainer
{
public:
    ViewContainer(RootContainer* parent, const QString& name);

    bool isRemote() const { return false; }

protected:
    virtual WaveContainer* createWaveContainer(const QString& name);
};

#endif // VIEWCONTAINER_H
