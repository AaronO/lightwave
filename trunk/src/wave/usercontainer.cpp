#include "usercontainer.h"
#include "rootcontainer.h"
#include "user.h"

UserContainer::UserContainer(RootContainer* parent, const QString& name)
    : WaveContainer(parent, name)
{
}

WaveContainer* UserContainer::createWaveContainer(const QString& name)
{
    Q_ASSERT(childContainer(name) == 0);
    User* u = new User(this, name);
    u->makePersistent();
    return u;
}
