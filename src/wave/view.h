#ifndef VIEW_H
#define VIEW_H

#include "wavecontainer.h"
#include <QScriptValue>

class ViewContainer;

class View : public WaveContainer
{
public:
    View(ViewContainer* parent, const QString& name);

    bool isRemote() const { return false; }
};

#endif // VIEW_H
