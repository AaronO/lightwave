#ifndef RICHTEXTMUTATION_H
#define RICHTEXTMUTATION_H

#include "abstractmutation.h"

class RichTextMutation : public AbstractMutation
{
public:
    RichTextMutation();
    RichTextMutation(bool create_empty);
    RichTextMutation(const JSONAbstractObject& mutation);
    RichTextMutation(const RichTextMutation& mutation);

    JSONArray content() const;
};

#endif // TEXTMUTATION_H
