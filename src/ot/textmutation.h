#ifndef TEXTMUTATION_H
#define TEXTMUTATION_H

#include "abstractmutation.h"

class TextMutation : public AbstractMutation
{
public:
    TextMutation();
    TextMutation(bool create_empty);
    TextMutation(const JSONAbstractObject& mutation);
    TextMutation(const TextMutation& mutation);

    JSONArray content() const;
};

#endif // TEXTMUTATION_H
