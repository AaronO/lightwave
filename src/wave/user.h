#ifndef USER_H
#define USER_H

#include "usercontainer.h"

class User : public WaveContainer
{
public:
    User(UserContainer* parent, const QString& name);

    bool isRemote() const { return false; }
    bool buildsDigest() const { return false; }

    WaveContainer* createWaveContainer(const QString& name);
};

#endif // USER_H
