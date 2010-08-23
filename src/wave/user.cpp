#include "user.h"

User::User(UserContainer* parent, const QString& name)
    : WaveContainer(parent, name)
{
}

WaveContainer* User::createWaveContainer(const QString& name)
{
    Q_UNUSED(name);

    return 0;
}
