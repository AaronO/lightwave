#ifndef RANDOMMUTATIONGENERATOR_H
#define RANDOMMUTATIONGENERATOR_H

#include <QSet>
#include "ot/abstractmutation.h"

class DocumentMutation;

class RandomMutationGenerator
{
public:
    RandomMutationGenerator(int lifts);

    DocumentMutation createDocumentMutation(JSONObject obj);

private:
    LiftMutation createLift();
    SqueezeMutation createSqueeze();
    AbstractMutation createMutation(JSONAbstractObject obj);
    ObjectMutation createMutation(JSONObject obj);
    ArrayMutation createMutation(JSONArray arr);
    TextMutation createMutation(const QString& str);
    RichTextMutation createRichTextMutation(JSONObject object);
    QString createString();

    int m_lifts;
    int m_liftCount;
};

#endif // RANDOMMUTATIONGENERATOR_H
