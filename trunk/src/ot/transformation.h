#ifndef TRANSFORMATION_H
#define TRANSFORMATION_H

#include "liftmutation.h"
#include "objectmutation.h"
#include <QString>
#include <QHash>

class TextMutation;
class ArrayMutation;
class ObjectMutation;

class Transformation
{
public:
    Transformation();

    void xform(ObjectMutation s, ObjectMutation c);

    bool hasError() { return !m_ok; }
    QString errorText() const { return m_errorText; }

private:
    QHash<QString,LiftMutation> sLifts;
    QHash<QString,AbstractMutation> sLiftCounterpart;
    QHash<QString,LiftMutation> cLifts;
    QHash<QString,AbstractMutation> cLiftCounterpart;

    /**
      * Find all lifts and squeezes and determine if any of these cannot be transformed.
      * Mark the dead lifts and squeezes.
      */
    void xform_pass0(ObjectMutation m1, ObjectMutation m2);
    void xform_pass0(ArrayMutation m1, ArrayMutation m2);
    void xform_pass0(RichTextMutation m1, RichTextMutation m2);
    void xform_pass0_lift(AbstractMutation s, AbstractMutation c);

    void xform_pass1(ObjectMutation m1, ObjectMutation m2);
    void xform_pass1(ArrayMutation m1, ArrayMutation m2);
    void xform_pass1(TextMutation m1, TextMutation m2);
    void xform_pass1(RichTextMutation m1, RichTextMutation m2);

    bool m_ok;
    QString m_errorText;
};

#endif // TRANSFORMATION_H
