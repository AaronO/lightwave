#ifndef SQUEEZEMUTATION_H
#define SQUEEZEMUTATION_H

#include "abstractmutation.h"

class SqueezeMutation : public AbstractMutation
{
public:
    SqueezeMutation();
    SqueezeMutation(const QString& id);
    SqueezeMutation(const JSONAbstractObject& mutation);
    SqueezeMutation(const SqueezeMutation& mutation);

    QString id() const;
};

#endif // SQUEEZEMUTATION_H
