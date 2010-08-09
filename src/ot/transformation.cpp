#include "transformation.h"
#include "documentmutation.h"
#include "objectmutation.h"
#include "textmutation.h"
#include "arraymutation.h"
#include "skipmutation.h"
#include "squeezemutation.h"
#include "textmutation.h"
#include "liftmutation.h"
#include "json/jsonobject.h"
#include "insertmutation.h"
#include "deletemutation.h"
#include "richtextmutation.h"
#include "json/jsonconstant.h"
#include <QtGlobal>

#define errorLog(x) { m_errorText = x; m_ok = false; }

Transformation::Transformation()
{
}

void Transformation::xform(ObjectMutation s, ObjectMutation c)
{
    m_ok = true;
    m_errorText = QString::null;

    // If at least one is null, then there is nothing to transform
    if ( s.isNull() || c.isNull() )
        return;
    xform_pass0(s,c);
    xform_pass1(s,c);
}

void Transformation::xform_pass0(ObjectMutation s, ObjectMutation c)
{
    JSONObject sobj = s.toObject();
    JSONObject cobj = c.toObject();

    foreach( QString name, sobj.attributeNames() )
    {
        if ( name[0] == '_')
            continue;

        AbstractMutation sm(sobj.attribute(name));
        AbstractMutation cm(cobj.attribute(name));

        if ( !cm.isObjectMutation() && !cm.isArrayMutation() && !cm.isRichTextMutation() && !cm.isTextMutation() && !cm.isInsertMutation() && !cm.isNull() )
        {
            errorLog("Client Mutation not allowed in an object");
            return;
        }
        if ( !sm.isObjectMutation() && !sm.isArrayMutation() && !sm.isRichTextMutation() && !sm.isTextMutation() && !sm.isInsertMutation() && !sm.isNull() )
        {
            errorLog("Server Mutation not allowed in an object");
            return;
        }

        // If at least one is null, then there is nothing to transform
        if ( sm.isNull() || cm.isNull() )
            continue;
        else if ( sm.isInsertMutation() || cm.isInsertMutation() )
            continue;
        else if ( sm.isObjectMutation() && cm.isObjectMutation() )
            xform_pass0(sm.toObjectMutation(), cm.toObjectMutation() );
        else if ( sm.isArrayMutation() && cm.isArrayMutation() )
            xform_pass0(sm.toArrayMutation(), cm.toArrayMutation() );
        else if ( sm.isTextMutation() && cm.isTextMutation() )
            continue;
        else if ( sm.isRichTextMutation() && cm.isRichTextMutation() )
            xform_pass0( sm.toRichTextMutation(), cm.toRichTextMutation() );
        else
        {
            errorLog("The two mutations of the object are not compatible");
            return;
        }

        if ( !m_ok)
            return;
    }
}

void Transformation::xform_pass0(ArrayMutation s, ArrayMutation c)
{
    JSONArray sarr = s.content();
    JSONArray carr = c.content();

    int sindex = 0;
    int cindex = 0;
    int sinside = 0;
    int cinside = 0;
    // Loop until end of one mutation is reached
    while( sindex < sarr.count() || cindex < carr.count() )
    {
        AbstractMutation smut;
        if ( sindex < sarr.count() )
            smut = sarr.at(sindex);
        AbstractMutation cmut;
        if ( cindex < carr.count() )
            cmut = carr.at(cindex);

        if ( !cmut.isObjectMutation() && !cmut.isArrayMutation() && !cmut.isRichTextMutation() && !cmut.isTextMutation() && !cmut.isInsertMutation() && !cmut.isNull() && !cmut.isSqueezeMutation() && !cmut.isLiftMutation() && !cmut.isDeleteMutation() && !cmut.isSkipMutation() )
        {
            errorLog("Client Mutation not allowed in an array");
            return;
        }
        if ( !smut.isObjectMutation() && !smut.isArrayMutation() && !smut.isRichTextMutation() && !smut.isTextMutation() && !smut.isInsertMutation() && !smut.isNull() && !smut.isSqueezeMutation() && !smut.isLiftMutation() && !smut.isDeleteMutation()  && !smut.isSkipMutation())
        {
            errorLog("Server Mutation not allowed in an array");
            return;
        }

        // Skip all server inserts
        while ( smut.isInsertMutation() || smut.isSqueezeMutation() )
        {
            sindex++;
            if ( sindex < sarr.count() )
            {
                smut = sarr.at(sindex);
                if ( !smut.isObjectMutation() && !smut.isArrayMutation() && !smut.isRichTextMutation() && !smut.isTextMutation() && !smut.isInsertMutation() && !smut.isNull() && !smut.isSqueezeMutation() && !smut.isLiftMutation() && !smut.isDeleteMutation() && !smut.isSkipMutation() )
                {
                    errorLog("Server Mutation not allowed in an array");
                    return;
                }
            }
            else
                smut = AbstractMutation();
        }
        // Skip all client inserts
        while ( cmut.isInsertMutation() || cmut.isSqueezeMutation() )
        {
            cindex++;
            if ( cindex < carr.count() )
            {
                cmut = carr.at(cindex);
                if ( !cmut.isObjectMutation() && !cmut.isArrayMutation() && !cmut.isRichTextMutation() && !cmut.isTextMutation() && !cmut.isInsertMutation() && !cmut.isNull() && !cmut.isSqueezeMutation() && !cmut.isLiftMutation() && !cmut.isDeleteMutation() && !cmut.isSkipMutation() )
                {
                    errorLog("Client Mutation not allowed in an array");
                    return;
                }
            }
            else
                cmut = AbstractMutation();
        }

        // End of mutation reached?
        if ( sindex == sarr.count() || cindex == carr.count() )
            break;

        if ( smut.isLiftMutation() )
        {
            qDebug("Set counterparts s:%s to %s", qPrintable(smut.toLiftMutation().id()), qPrintable(cmut.toJSON()));
            sLifts[smut.toLiftMutation().id()] = smut;
            cLiftCounterpart[smut.toLiftMutation().id()] = cmut;
        }
        if ( cmut.isLiftMutation() )
        {
            qDebug("Set counterparts c:%s to %s", qPrintable(cmut.toLiftMutation().id()), qPrintable(smut.toJSON()));
            cLifts[cmut.toLiftMutation().id()] = cmut;
            sLiftCounterpart[cmut.toLiftMutation().id()] = smut;
        }
        if ( smut.isLiftMutation() && cmut.isLiftMutation() && smut.toLiftMutation().hasMutation() && cmut.toLiftMutation().hasMutation() )
        {
            xform_pass0_lift( smut.toLiftMutation().mutation(), cmut.toLiftMutation().mutation() );
        }
        else if ( smut.isLiftMutation() && smut.toLiftMutation().hasMutation() && (cmut.isArrayMutation() || cmut.isRichTextMutation() || cmut.isTextMutation() || cmut.isObjectMutation() ) )
        {
            xform_pass0_lift( smut.toLiftMutation().mutation(), cmut );
        }
        else if ( cmut.isLiftMutation() && cmut.toLiftMutation().hasMutation() && (smut.isArrayMutation() || smut.isRichTextMutation() || smut.isTextMutation() || smut.isObjectMutation() ) )
        {
            xform_pass0_lift( smut, cmut.toLiftMutation().mutation() );
        }

        if ( (smut.isDeleteMutation() || smut.isSkipMutation() ) && ( cmut.isDeleteMutation() || cmut.isSkipMutation()))
        {
            int sdel, cdel;
            if ( smut.isDeleteMutation() )
                sdel = smut.toDeleteMutation().count();
            else
                sdel = smut.toSkipMutation().count();
            if ( cmut.isDeleteMutation() )
                cdel = cmut.toDeleteMutation().count();
            else
                cdel = cmut.toSkipMutation().count();
            int del = qMin(sdel - sinside, cdel - cinside);
            sinside += del;
            cinside += del;
            if ( sdel == del )
            {
                sinside = 0;
                sindex++;
            }
            if ( cdel == del )
            {
                cinside = 0;
                cindex++;
            }
        }
        else if ( smut.isSkipMutation() ) // ... and mutation at the client
        {
            Q_ASSERT(cmut.isArrayMutation() || cmut.isObjectMutation() || cmut.isRichTextMutation() || cmut.isTextMutation() || cmut.isLiftMutation() );
            cindex++;
            sinside++;
            if ( smut.toSkipMutation().count() == sinside )
            {
                sinside = 0;
                sindex++;
            }
        }
        else if ( cmut.isSkipMutation() ) // ... and mutation at the srver
        {
            Q_ASSERT(smut.isArrayMutation() || smut.isObjectMutation() || smut.isRichTextMutation() || smut.isTextMutation() || smut.isLiftMutation() );
            sindex++;
            cinside++;
            if ( cmut.toSkipMutation().count() == cinside )
            {
                cinside = 0;
                cindex++;
            }
        }
        else if ( smut.isDeleteMutation() ) // ... and mutation at the client
        {
            Q_ASSERT(cmut.isArrayMutation() || cmut.isObjectMutation() || cmut.isRichTextMutation() || cmut.isTextMutation() || cmut.isLiftMutation() );
            cindex++;
            sinside++;
            if ( smut.toDeleteMutation().count() == sinside )
            {
                sinside = 0;
                sindex++;
            }
        }
        else if ( cmut.isDeleteMutation() ) // ... and mutation at the server
        {
            Q_ASSERT(smut.isArrayMutation() || smut.isObjectMutation() || smut.isRichTextMutation() || smut.isTextMutation() || smut.isLiftMutation() );
            sindex++;
            cinside++;
            if ( cmut.toDeleteMutation().count() == cinside )
            {
                cinside = 0;
                cindex++;
            }
        }
        else if ( smut.isLiftMutation() )
        {
            Q_ASSERT(cmut.isArrayMutation() || cmut.isObjectMutation() || cmut.isRichTextMutation() || cmut.isTextMutation() || cmut.isLiftMutation() );
            sindex++;
            cindex++;
        }
        else if ( cmut.isLiftMutation() )
        {
            Q_ASSERT(smut.isArrayMutation() || smut.isObjectMutation() || smut.isRichTextMutation() || smut.isTextMutation() || smut.isLiftMutation() );
            sindex++;
            cindex++;
        }
        else if ( smut.isArrayMutation() && cmut.isArrayMutation() )
        {
            xform_pass0(smut.toArrayMutation(), cmut.toArrayMutation() );
            cindex++;
            sindex++;
        }
        else if ( smut.isObjectMutation() && cmut.isObjectMutation() )
        {
            xform_pass0(smut.toObjectMutation(), cmut.toObjectMutation() );
            cindex++;
            sindex++;
        }
        else if ( smut.isTextMutation() && cmut.isTextMutation() )
        {
            cindex++;
            sindex++;
        }
        else if ( smut.isRichTextMutation() && cmut.isRichTextMutation() )
        {
            xform_pass0(smut.toRichTextMutation(), cmut.toRichTextMutation() );
            cindex++;
            sindex++;
        }
        else
        {
            errorLog("The two mutations in the array do not match");
            qDebug("smut: %s",qPrintable(smut.toJSON()));
            qDebug("cmut: %s",qPrintable(cmut.toJSON()));
            return;
        }

        if ( !m_ok )
            return;
    }
}

void Transformation::xform_pass0(RichTextMutation s, RichTextMutation c)
{
    JSONArray sarr = s.content();
    JSONArray carr = c.content();

    int sindex = 0;
    int cindex = 0;
    int sinside = 0;
    int cinside = 0;
    // Loop until end of one mutation is reached
    while( sindex < sarr.count() || cindex < carr.count() )
    {
        AbstractMutation smut;
        if ( sindex < sarr.count() )
            smut = sarr.at(sindex);
        AbstractMutation cmut;
        if ( cindex < carr.count() )
            cmut = carr.at(cindex);

        if ( !cmut.isObjectMutation() && !cmut.isArrayMutation() && !cmut.isRichTextMutation() && !cmut.isInsertMutation() && !cmut.isNull() && !cmut.isDeleteMutation() && !cmut.isSkipMutation() )
        {
            errorLog("Client Mutation not allowed in an array");
            return;
        }
        if ( !smut.isObjectMutation() && !smut.isArrayMutation() && !smut.isRichTextMutation() && !smut.isInsertMutation() && !smut.isNull() && !smut.isDeleteMutation()  && !smut.isSkipMutation() )
        {
            errorLog("Server Mutation not allowed in an array");
            return;
        }

        // Skip all server inserts
        if ( smut.isInsertMutation() )
        {
            sindex++;
            continue;
        }
        // Skip all client inserts
        if ( cmut.isInsertMutation() )
        {
            cindex++;
            continue;
        }

        // End of mutation reached?
        if ( sindex == sarr.count() || cindex == carr.count() )
            break;

        if ( (smut.isDeleteMutation() || smut.isSkipMutation() ) && ( cmut.isDeleteMutation() || cmut.isSkipMutation()))
        {
            qDebug("Skip/Del 0");
            int sdel, cdel;
            if ( smut.isDeleteMutation() )
                sdel = smut.toDeleteMutation().count();
            else
                sdel = smut.toSkipMutation().count();
            if ( cmut.isDeleteMutation() )
                cdel = cmut.toDeleteMutation().count();
            else
                cdel = cmut.toSkipMutation().count();
            int del = qMin(sdel - sinside, cdel - cinside);
            sinside += del;
            cinside += del;
            if ( sdel == del )
            {
                sinside = 0;
                sindex++;
            }
            if ( cdel == del )
            {
                cinside = 0;
                cindex++;
            }
        }
        else if ( smut.isSkipMutation() ) // ... and mutation at the client
        {
            Q_ASSERT(cmut.isArrayMutation() || cmut.isObjectMutation() || cmut.isRichTextMutation() );
            cindex++;
            sinside++;
            if ( smut.toSkipMutation().count() == sinside )
            {
                sinside = 0;
                sindex++;
            }
        }
        else if ( cmut.isSkipMutation() ) // ... and mutation at the srver
        {
            Q_ASSERT(smut.isArrayMutation() || smut.isObjectMutation() || smut.isRichTextMutation() );
            sindex++;
            cinside++;
            if ( cmut.toSkipMutation().count() == cinside )
            {
                cinside = 0;
                cindex++;
            }
        }
        else if ( smut.isDeleteMutation() ) // ... and mutation at the client
        {
            Q_ASSERT(cmut.isArrayMutation() || cmut.isObjectMutation() || cmut.isRichTextMutation() );
            cindex++;
            sinside++;
            if ( smut.toDeleteMutation().count() == sinside )
            {
                sinside = 0;
                sindex++;
            }
        }
        else if ( cmut.isDeleteMutation() ) // ... and mutation at the server
        {
            Q_ASSERT(smut.isArrayMutation() || smut.isObjectMutation() || smut.isRichTextMutation() );
            sindex++;
            cinside++;
            if ( cmut.toDeleteMutation().count() == cinside )
            {
                cinside = 0;
                cindex++;
            }
        }
        else if ( smut.isArrayMutation() && cmut.isArrayMutation() )
        {
            xform_pass0(smut.toArrayMutation(), cmut.toArrayMutation() );
            cindex++;
            sindex++;
        }
        else if ( smut.isObjectMutation() && cmut.isObjectMutation() )
        {
            xform_pass0(smut.toObjectMutation(), cmut.toObjectMutation() );
            cindex++;
            sindex++;
        }
        else if ( smut.isRichTextMutation() && cmut.isRichTextMutation() )
        {
            xform_pass0(smut.toRichTextMutation(), cmut.toRichTextMutation() );
            cindex++;
            sindex++;
        }
        else
        {
            errorLog("The two mutations in the array do not match");
            qDebug("smut: %s",qPrintable(smut.toJSON()));
            qDebug("cmut: %s",qPrintable(cmut.toJSON()));
            return;
        }

        if ( !m_ok )
            return;
    }
}

void Transformation::xform_pass0_lift(AbstractMutation s, AbstractMutation c)
{
    if ( s.isObjectMutation() && c.isObjectMutation() )
    {
        xform_pass0(s.toObjectMutation(), c.toObjectMutation());
        xform_pass1(s.toObjectMutation(), c.toObjectMutation());
    }
    else if ( s.isArrayMutation() && c.isArrayMutation() )
    {
        xform_pass0(s.toArrayMutation(), c.toArrayMutation());
        xform_pass1(s.toArrayMutation(), c.toArrayMutation());
    }
    else if ( s.isTextMutation() && c.isTextMutation() )
    {
        xform_pass1(s.toTextMutation(), c.toTextMutation() );
    }
    else if ( s.isRichTextMutation() && c.isRichTextMutation() )
    {
        xform_pass0(s.toRichTextMutation(), c.toRichTextMutation() );
        xform_pass1(s.toRichTextMutation(), c.toRichTextMutation() );
    }
    else
        errorLog("The two mutations are either incompatible or they are not allowed inside a lift");
}

void Transformation::xform_pass1(ObjectMutation s, ObjectMutation c)
{
    JSONObject sobj = s.toObject();
    JSONObject cobj = c.toObject();

    foreach( QString name, sobj.attributeNames() )
    {        
        if ( name[0] == '_')
            continue;
        // Both mutations modify the same attribute? If not -> skip
        if ( !cobj.hasAttribute(name) )
            continue;

        AbstractMutation smut(sobj.attribute(name));
        AbstractMutation cmut(cobj.attribute(name));

        if ( smut.isInsertMutation() )
            cobj.removeAttribute(name);
        else if ( smut.isObjectMutation() )
        {
            if ( cmut.isObjectMutation() )
                xform_pass1(smut.toObjectMutation(), cmut.toObjectMutation() );
            else if ( cmut.isInsertMutation() )
                sobj.removeAttribute(name);
            else
                errorLog("The two mutations are not compatible and/or not allowed inside an object mutation. This should have been detected in pass 0");
        }
        else if ( smut.isArrayMutation() )
        {
            if ( cmut.isArrayMutation() )
                xform_pass1(smut.toArrayMutation(), cmut.toArrayMutation() );
            else if ( cmut.isInsertMutation() )
                sobj.removeAttribute(name);
            else
                errorLog("The two mutations are not compatible and/or not allowed inside an object mutation. This should have been detected in pass 0");
        }
        else if ( smut.isTextMutation() )
        {
            if ( cmut.isTextMutation() )
                xform_pass1(smut.toTextMutation(), cmut.toTextMutation() );
            else if ( cmut.isInsertMutation() )
                sobj.removeAttribute(name);
            else
                errorLog("The two mutations are not compatible and/or not allowed inside an object mutation. This should have been detected in pass 0");
        }
        else if ( smut.isRichTextMutation() )
        {
            if ( cmut.isRichTextMutation() )
                xform_pass1(smut.toRichTextMutation(), cmut.toRichTextMutation() );
            else if ( cmut.isInsertMutation() )
                sobj.removeAttribute(name);
            else
                errorLog("The two mutations are not compatible and/or not allowed inside an object mutation. This should have been detected in pass 0");
        }
        else
            errorLog("This mutation is not allowed inside an object mutation. This should have been detected in pass 0");

        if ( !m_ok )
            return;
    }
}

void Transformation::xform_pass1(ArrayMutation s, ArrayMutation c)
{
    JSONArray sarr = s.content();
    JSONArray carr = c.content();

    JSONArray sdebug( sarr.clone().toArray() );
    JSONArray cdebug( carr.clone().toArray() );

    int sindex = 0;
    int cindex = 0;
    int sinside = 0;
    int cinside = 0;

    // Loop until end of one mutation is reached
    while( sindex < sarr.count() || cindex < carr.count() )
    {
        AbstractMutation smut;
        if ( sindex < sarr.count() )
            smut = sarr.at(sindex);
        AbstractMutation cmut;
        if ( cindex < carr.count() )
            cmut = carr.at(cindex);

        //
        // Server insert/squeeze go first
        //

        while ( smut.isInsertMutation() || smut.isSqueezeMutation() )
        {
            //qDebug("server insert");
            if ( cinside > 0 )
            {
                if ( cmut.isDeleteMutation() )
                {
                    carr.insert(cindex+1, DeleteMutation( cmut.toDeleteMutation().count() - cinside));
                    cmut.toDeleteMutation().setCount(cinside);
                    cindex++;
                    cinside = 0;
                    cmut = carr.at(cindex);
                }
                else if ( cmut.isSkipMutation() )
                {
                    carr.insert(cindex+1, SkipMutation( cmut.toSkipMutation().count() - cinside));
                    cmut.toSkipMutation().setCount(cinside);
                    cindex++;
                    cinside = 0;
                    cmut = carr.at(cindex);
                }
            }

            if ( smut.isInsertMutation() )
            {
                // TODO: check the Insert mutation individually for correctness
                sindex++;
                carr.insert(cindex++, SkipMutation(1));
            }
            else if ( smut.isSqueezeMutation() )
            {
                SqueezeMutation sSqueeze = smut.toSqueezeMutation();                
                // Which operation does the client side have for this object?
                AbstractMutation c = cLiftCounterpart[sSqueeze.id()];
                if ( c.isSkipMutation())
                {
                    // Server lift remains, client skips it at the new position
                    carr.insert(cindex++, SkipMutation(1));
                    sindex++;
                }
                else if ( c.isDeleteMutation())
                {
                    // Client deletes the object at its the new position
                    carr.insert(cindex++, DeleteMutation(1));
                    // Server removes squeeze because it is already deleted by the client
                    sarr.removeAt(sindex);
                }
                else if ( c.isLiftMutation() )
                {
                    LiftMutation sLift = sLifts[sSqueeze.id()];
                    LiftMutation cLift = c.toLiftMutation();
                    if ( cLift.hasMutation() )
                    {
                        // AbstractMutation c2( cLift.mutation().clone() );
                        // xform_pass1_lift(sLift.mutation(), c2);
                        carr.insert(cindex++, cLift.mutation().clone());
                        sindex++;
                    }
                    else
                    {
                        // Client skips the squeezed object and does not lift it
                        carr.insert(cindex++, SkipMutation(1));
                        // Server keeps its squeeze at this position
                        sindex++;
                    }
                }
                else if ( c.isInsertMutation() )
                {
                    // Client deletes the object at its the new position
                    carr.insert(cindex++, DeleteMutation(1));
                    // The server does not lift the object and the client overwrites the JSONObject attribute
                    sarr.removeAt(sindex);
                }
                else if ( !c.isNull() )
                {
                    LiftMutation sLift = sLifts[sSqueeze.id()];
                    if ( sLift.hasMutation() )
                    {
                        // AbstractMutation c2(c.clone());
                        // xform_pass1_lift(sLift.mutation(), c2);
                        // The client applies its mutation at the object's new position
                        carr.insert(cindex++, c.clone());
                        sindex++;
                    }
                    else
                    {
                        // The client applies its mutation at the object's new position
                        carr.insert(cindex++, c);
                        sindex++;
                    }
                }
                else
                {
                    sindex++;
                    carr.insert(cindex++, SkipMutation(1));
                }
            }
            if ( sindex < sarr.count() )
                smut = sarr.at(sindex);
            else
            {
                smut = AbstractMutation();
                break;
            }
        }

        //
        // Client insert/squeeze go next
        //

        while ( cmut.isInsertMutation() || cmut.isSqueezeMutation() )
        {
            //qDebug("client insert");
            if ( sinside > 0 )
            {
                if ( smut.isDeleteMutation() )
                {
                    sarr.insert(sindex+1, DeleteMutation( smut.toDeleteMutation().count() - sinside));
                    smut.toDeleteMutation().setCount(sinside);
                    sindex++;
                    sinside = 0;
                    smut = sarr.at(sindex);
                }
                else if ( smut.isSkipMutation() )
                {
                    sarr.insert(sindex+1, SkipMutation( smut.toSkipMutation().count() - sinside));
                    smut.toSkipMutation().setCount(sinside);
                    sindex++;
                    sinside = 0;
                    smut = sarr.at(sindex);
                }
            }

            if ( cmut.isInsertMutation() )
            {
                // TODO: check the Insert mutation individually for correctness
                cindex++;
                sarr.insert(sindex++, SkipMutation(1));
            }
            else
            {
                SqueezeMutation cSqueeze = cmut.toSqueezeMutation();
                // Which operation does the server side have for this object?
                AbstractMutation s = sLiftCounterpart[cSqueeze.id()];
                if ( s.isSkipMutation())
                {
                    // Client lift remains, server skips it at the new position
                    sarr.insert(sindex++, SkipMutation(1));
                    cindex++;
                }
                else if ( s.isDeleteMutation())
                {
                    // Server deletes the object at its the new position
                    sarr.insert(sindex++, DeleteMutation(1));
                    // Client removes squeeze because it is already deleted by the client
                    carr.removeAt(cindex);
                }
                else if ( s.isLiftMutation() )
                {
                    LiftMutation cLift = cLifts[cSqueeze.id()];
                    LiftMutation sLift = s.toLiftMutation();
                    if ( sLift.hasMutation() && cLift.hasMutation() )
                    {
                        // AbstractMutation s2( sLift.mutation().clone() );
                        // xform_pass1_lift(s2, cLift.mutation());
                        // The server must lift the object here instead
                        sarr.insert(sindex++, s.clone());
                        // The server lifted this object as well -> the client cannot lift it ->
                        // then the client cannot squeeze it in here
                        carr.removeAt(cindex);
                    }
                    else
                    {
                        // The server lifted this object as well -> the client cannot lift it ->
                        // then the client cannot squeeze it in here
                        carr.removeAt(cindex);
                        // The server must lift the object here instead
                        sarr.insert(sindex++, s.clone());
                    }
                }
                else if ( s.isInsertMutation() )
                {
                    // This can only happen if the squeezed object is currently a JSONObject attribute.
                    // The client does not lift the object -> it does nots squeeze it here, and the server overwrites the JSONObject attribute
                    carr.removeAt(cindex);
                    // Server deletes the object at its the new position
                    sarr.insert(sindex++, DeleteMutation(1));
                }
                else if ( !s.isNull() )
                {
                    LiftMutation cLift = cLifts[cSqueeze.id()];
                    if ( cLift.hasMutation() )
                    {
                        // AbstractMutation s2( s.clone() );
                        // xform_pass1_lift(s2, cLift.mutation());
                        // The server mutates the object at its new position
                        sarr.insert(sindex++, s.clone());
                        // The client squeezes the object in here
                        cindex++;
                    }
                    else
                    {
                        // The server mutates the object at its new position
                        sarr.insert(sindex++, s);
                        // The client squeezes the object in here
                        cindex++;
                    }
                }
                else
                {
                    sarr.insert(sindex++, SkipMutation(1));
                    cindex++;
                }
            }
            if ( cindex < carr.count() )
                cmut = carr.at(cindex);
            else
            {
                cmut = AbstractMutation();
                break;
            }
        }

        if ( sindex == sarr.count() )
        {
            if ( cmut.isInsertMutation() || cmut.isSqueezeMutation() )
                continue;
            break;
        }
        if ( cindex == carr.count() )
        {
            //qDebug("smut = %s",qPrintable(smut.toJSON()));
            if ( smut.isInsertMutation() || smut.isSqueezeMutation() )
                continue;
            break;
        }

        //
        // Lift, Skip, Delete, mutations
        //

        if ( smut.isLiftMutation() )
        {
            // Both are lifting the same object?
            if ( cmut.isLiftMutation() )
            {
                // The client removes its lift. If it has a mutation then it will be moved to the corresponding squeeze
                carr.removeAt(cindex);
                // The server removes its lift. It will be places where the client moved it.
                sarr.removeAt(sindex);
            }
            else if ( cmut.isDeleteMutation() )
            {
                // The server does not lift the object because it is deleted
                sarr.removeAt(sindex);
                // The client removes its delete. The delete is put where the server squeezed it.
                DeleteMutation cdel = cmut.toDeleteMutation();
                cdel.setCount( cdel.count() - 1 );
                if ( cinside == cdel.count() )
                {
                    cinside = 0;
                    if ( cdel.count() == 0 )
                        carr.removeAt(cindex);
                    else
                        cindex++;
                }
            }
            else if ( cmut.isSkipMutation() )
            {
                // The client does not skip this element here. It is skipped where it is squeezed in
                SkipMutation cskip = cmut.toSkipMutation();
                cskip.setCount( cskip.count() - 1 );
                if ( cinside == cskip.count() )
                {
                    cinside = 0;
                    if ( cskip.count() == 0 )
                        carr.removeAt(cindex);
                    else
                        cindex++;
                }
                // Server remains with its lift
                sindex++;
            }
            else
            {
                Q_ASSERT(cmut.isArrayMutation() || cmut.isObjectMutation() || cmut.isTextMutation() || cmut.isRichTextMutation() );
                // The client removes its mutation here. It is shifted to the new position where the object is squeezed in.
                carr.removeAt(cindex);
                // Server remains with its lift
                sindex++;
            }
        }
        else if ( cmut.isLiftMutation() )
        {
            if ( smut.isDeleteMutation() )
            {
                // The client does not lift the object because it is deleted
                carr.removeAt(cindex);
                // The server removes its delete. The delete is put where the client squeezed it.
                DeleteMutation sdel = smut.toDeleteMutation();
                sdel.setCount( sdel.count() - 1 );
                if ( sinside == sdel.count() )
                {
                    sinside = 0;
                    if ( sdel.count() == 0 )
                        sarr.removeAt(sindex);
                    else
                        sindex++;
                }
            }
            else if ( smut.isSkipMutation() )
            {
                // The server does not skip this element here. It is skipped where it is squeezed in
                SkipMutation sskip = smut.toSkipMutation();
                sskip.setCount( sskip.count() - 1 );
                if ( sinside == sskip.count() )
                {
                    sinside = 0;
                    if ( sskip.count() == 0 )
                        sarr.removeAt(sindex);
                    else
                        sindex++;
                }
                // Client remains with its lift
                cindex++;
            }
            else
            {
                Q_ASSERT(smut.isArrayMutation() || smut.isObjectMutation() || smut.isTextMutation() || smut.isRichTextMutation() );
                // The server removes its mutation here. It is shifted to the new position where the object is squeezed in.
                sarr.removeAt(sindex);
                // Client remains with its lift
                cindex++;
            }
        }
        else if ( smut.isDeleteMutation() && cmut.isDeleteMutation() )
        {
            int sdel = smut.toDeleteMutation().count();
            int cdel = cmut.toDeleteMutation().count();
            int del = qMin(sdel - sinside, cdel - cinside);
            smut.toDeleteMutation().setCount(cdel - del);
            smut.toDeleteMutation().setCount(sdel - del);
            if ( sinside + del == sdel )
            {
                sinside = 0;
                if ( sdel == del )
                    sarr.removeAt(sindex);
                else
                    sindex++;
            }
            if ( cinside + del == cdel )
            {
                cinside = 0;
                if ( cdel == del )
                    carr.removeAt(cindex);
                else
                    cindex++;
            }
        }
        else if ( smut.isSkipMutation() && cmut.isSkipMutation() )
        {
            int sskip = smut.toSkipMutation().count();
            int cskip = cmut.toSkipMutation().count();
            int skip = qMin(sskip - sinside, cskip - cinside);
            sinside += skip;
            cinside += skip;
            if ( sinside == sskip )
            {
                sinside = 0;
                sindex++;
            }
            if ( cinside == cskip )
            {
                cinside = 0;
                cindex++;
            }
        }
        else if ( smut.isDeleteMutation() && cmut.isSkipMutation() )
        {
            int sdel = smut.toDeleteMutation().count();
            int cskip = cmut.toSkipMutation().count();
            int count = qMin(sdel - sinside, cskip - cinside);
            cmut.toSkipMutation().setCount( cskip - count );
            sinside += count;
            if ( sinside == sdel )
            {
                sinside = 0;
                sindex++;
            }
            if ( cinside + count == cskip )
            {
                cinside = 0;
                if ( cskip == count )
                    carr.removeAt(cindex);
                else
                    cindex++;
            }
        }
        else if ( smut.isSkipMutation() && cmut.isDeleteMutation() )
        {
            int sskip = smut.toSkipMutation().count();
            int cdel = cmut.toDeleteMutation().count();
            int count = qMin(sskip - sinside, cdel - cinside);
            smut.toSkipMutation().setCount( sskip - count );
            if ( sinside + count == sskip )
            {
                sinside = 0;
                if ( sskip == count )
                    sarr.removeAt(sindex);
                else
                    sindex++;
            }
            if ( cinside + count == cdel )
            {
                cinside = 0;
                cindex++;
            }
        }
        else if ( smut.isSkipMutation() ) // ... and mutation at the client
        {
            cindex++;
            sinside++;
            if ( smut.toSkipMutation().count() == sinside )
            {
                sinside = 0;
                sindex++;
            }
        }
        else if ( cmut.isSkipMutation() ) // ... and mutation at the srver
        {
            sindex++;
            cinside++;
            if ( cmut.toSkipMutation().count() == cinside )
            {
                cinside = 0;
                cindex++;
            }
        }
        else if ( smut.isDeleteMutation() ) // ... and mutation at the client
        {
            carr.removeAt(cindex);
            sinside++;
            if ( smut.toDeleteMutation().count() == sinside )
            {
                sinside = 0;
                sindex++;
            }
        }
        else if ( cmut.isDeleteMutation() ) // ... and mutation at the server
        {
            sarr.removeAt(sindex);
            cinside++;
            if ( cmut.toDeleteMutation().count() == cinside )
            {
                cinside = 0;
                cindex++;
            }
        }
        else
        {
            if ( smut.isObjectMutation() && cmut.isObjectMutation() )
                xform_pass1(smut.toObjectMutation(), cmut.toObjectMutation() );
            else if ( smut.isArrayMutation() && cmut.isArrayMutation() )
                xform_pass1(smut.toArrayMutation(), cmut.toArrayMutation() );
            else if ( smut.isTextMutation() && cmut.isTextMutation() )
                xform_pass1(smut.toTextMutation(), cmut.toTextMutation() );
            else if ( smut.isRichTextMutation() && cmut.isRichTextMutation() )
                xform_pass1(smut.toRichTextMutation(), cmut.toRichTextMutation() );
            else
                errorLog("The mutations are not compatible or not allowed inside an array mutation");
            sarr[sindex++] = smut;
            carr[cindex++] = cmut;
        }
        //qDebug("s = %s", qPrintable(s.toJSON()));
        //qDebug("c = %s", qPrintable(c.toJSON()));

        if ( !m_ok )
            return;
    }

    if( sindex < sarr.count() )
    {
        errorLog("Client array mutation is shorter than the server version");
        qDebug("smut: %s",qPrintable(sarr.toJSON()));
        qDebug("cmut: %s",qPrintable(carr.toJSON()));
        return;
    }
    if( cindex < carr.count() )
    {
        errorLog("Client array mutation is longer than the server version");
        qDebug("smut: %s",qPrintable(sarr.toJSON()));
        qDebug("cmut: %s",qPrintable(carr.toJSON()));
        return;
    }

    // DEBUG

    int scount = 0;
    for( int i = 0; i < sdebug.count(); ++i )
    {
        AbstractMutation m( sdebug[i] );
        if ( m.isInsertMutation() || m.isSqueezeMutation() )
            scount++;
        else if ( m.isDeleteMutation() )
            continue;
        else if ( m.isSkipMutation() )
            scount += m.toSkipMutation().count();
        else if ( m.isLiftMutation() )
            continue;
        else
            scount++;
    }

    int snewcount = 0;
    int slifts = 0;
    int ssqueezes = 0;
    for( int i = 0; i < sarr.count(); ++i )
    {
        AbstractMutation m( sarr[i] );
        if ( m.isInsertMutation() )
            continue;
        else if ( m.isSqueezeMutation() )
        {
            ssqueezes++;
            continue;
        }
        else if ( m.isDeleteMutation() )
            snewcount += m.toDeleteMutation().count();
        else if ( m.isSkipMutation() )
            snewcount += m.toSkipMutation().count();
        else if ( m.isLiftMutation() )
        {
            snewcount++;
            slifts++;
        }
        else
            snewcount++;
    }

    int ccount = 0;
    for( int i = 0; i < cdebug.count(); ++i )
    {
        AbstractMutation m( cdebug[i] );
        if ( m.isInsertMutation() )
            ccount++;
        else if ( m.isSqueezeMutation() )
        {
//            qDebug("Client id=%s, counterpart:%s", qPrintable(m.toSqueezeMutation().id()), qPrintable(sLiftCounterpart[m.toSqueezeMutation().id()].toJSON()));
            ccount++;
        }
        else if ( m.isDeleteMutation() )
            continue;
        else if ( m.isSkipMutation() )
            ccount += m.toSkipMutation().count();
        else if ( m.isLiftMutation() )
            continue;
        else
            ccount++;
    }

    int cnewcount = 0;
    int csqueezes = 0;
    int clifts = 0;
    for( int i = 0; i < carr.count(); ++i )
    {
        AbstractMutation m( carr[i] );
        if ( m.isInsertMutation() )
            continue;
        else if ( m.isSqueezeMutation() )
        {
            csqueezes++;
            continue;
        }
        else if ( m.isDeleteMutation() )
            cnewcount += m.toDeleteMutation().count();
        else if ( m.isSkipMutation() )
            cnewcount += m.toSkipMutation().count();
        else if ( m.isLiftMutation() )
        {
            clifts++;
            cnewcount++;
        }
        else
            cnewcount++;
    }

    if ( snewcount != ccount || cnewcount != scount )
    {
        qDebug("sarr: %s", qPrintable(sarr.toJSON()));
        qDebug("carr: %s", qPrintable(carr.toJSON()));
        qDebug("sdebug: %s", qPrintable(sdebug.toJSON()));
        qDebug("cdebug: %s", qPrintable(cdebug.toJSON()));
        Q_ASSERT(false);
    }

    if (ssqueezes != slifts )
    {
        qDebug("sarr: %s", qPrintable(sarr.toJSON()));
        qDebug("carr: %s", qPrintable(carr.toJSON()));
        qDebug("sdebug: %s", qPrintable(sdebug.toJSON()));
        qDebug("cdebug: %s", qPrintable(cdebug.toJSON()));
        Q_ASSERT(ssqueezes == slifts);
    }

    Q_ASSERT(csqueezes == clifts);

    // END DEBUG
}

void Transformation::xform_pass1(TextMutation s, TextMutation c)
{
    JSONArray sarr = s.content();
    JSONArray carr = c.content();

    int sindex = 0;
    int cindex = 0;
    int sinside = 0;
    int cinside = 0;
    // Loop until end of one mutation is reached
    while( sindex < sarr.count() || cindex < carr.count() )
    {
        AbstractMutation smut;
        if ( sindex < sarr.count() )
            smut = sarr.at(sindex);
        AbstractMutation cmut;
        if ( cindex < carr.count() )
            cmut = carr.at(cindex);

        // Server insertions go first
        if ( smut.isInsertMutation() )
        {
            // In the middle of a client skip/delete? Split it
            if ( cinside > 0 )
            {
                if ( cmut.isDeleteMutation() )
                {
                    Q_ASSERT(cmut.toDeleteMutation().count() - cinside > 0);
                    carr.insert(cindex+1, DeleteMutation( cmut.toDeleteMutation().count() - cinside));
                    cmut.toDeleteMutation().setCount(cinside);
                }
                else if ( cmut.isSkipMutation() )
                {
                    Q_ASSERT(cmut.toSkipMutation().count() - cinside > 0);
                    carr.insert(cindex+1, SkipMutation( cmut.toSkipMutation().count() - cinside));
                    cmut.toSkipMutation().setCount(cinside);
                }
                cindex++;
                cinside = 0;
                cmut = carr.at(cindex);
            }

            InsertMutation ins = smut.toInsertMutation();
            if ( !ins.isString() )
            {
                errorLog("Only strings allowed inside a text mutation");
                return;
            }
            // TODO: check the Insert mutation individually for correctness
            sindex++;
            if ( ins.toString().length() > 0 )
                carr.insert(cindex++, SkipMutation(ins.toString().length()));
            continue;
        }
        // Client insertions go next
        if ( cmut.isInsertMutation() )
        {
            // In the middle of a server skip/delete? Split it
            if ( sinside > 0 )
            {
                if ( smut.isDeleteMutation() )
                {
                    Q_ASSERT(smut.toDeleteMutation().count() - sinside > 0);
                    sarr.insert(sindex+1, DeleteMutation( smut.toDeleteMutation().count() - sinside));
                    smut.toDeleteMutation().setCount(sinside);
                }
                else if ( smut.isSkipMutation() )
                {
                    Q_ASSERT(smut.toSkipMutation().count() - sinside > 0);
                    sarr.insert(sindex+1, SkipMutation( smut.toSkipMutation().count() - sinside));
                    smut.toSkipMutation().setCount(sinside);
                }
                sindex++;
                sinside = 0;
                smut = sarr.at(sindex);
            }

            InsertMutation ins = cmut.toInsertMutation();
            if ( !ins.isString() )
            {
                errorLog("Only strings allowed inside a text mutation");
                return;
            }
            // TODO: check the Insert mutation individually for correctness
            cindex++;
            if ( ins.toString().length() > 0 )
                sarr.insert(sindex++, SkipMutation(ins.toString().length()));
            continue;
        }

        // End of mutation reached?
        if ( sindex == sarr.count() || cindex == carr.count() )
            break;

        if ( smut.isDeleteMutation() && cmut.isDeleteMutation() )
        {
            int sdel = smut.toDeleteMutation().count();
            int cdel = cmut.toDeleteMutation().count();
            int del = qMin(sdel - sinside, cdel - cinside);
            smut.toDeleteMutation().setCount(cdel - del);
            smut.toDeleteMutation().setCount(sdel - del);
            if ( sinside + del == sdel )
            {
                sinside = 0;
                if ( sdel == del )
                    sarr.removeAt(sindex);
                else
                    sindex++;
            }
            if ( cinside + del == cdel )
            {
                cinside = 0;
                if ( cdel == del )
                    carr.removeAt(cindex);
                else
                    cindex++;
            }
        }
        else if ( smut.isSkipMutation() && cmut.isSkipMutation() )
        {
            int sskip = smut.toSkipMutation().count();
            int cskip = cmut.toSkipMutation().count();
            int skip = qMin(sskip - sinside, cskip - cinside);
            sinside += skip;
            cinside += skip;
            if ( sinside == sskip )
            {
                sinside = 0;
                sindex++;
            }
            if ( cinside == cskip )
            {
                cinside = 0;
                cindex++;
            }
        }
        else if ( smut.isDeleteMutation() && cmut.isSkipMutation() )
        {
            int sdel = smut.toDeleteMutation().count();
            int cskip = cmut.toSkipMutation().count();
            int count = qMin(sdel - sinside, cskip - cinside);
            cmut.toSkipMutation().setCount( cskip - count );
            if ( sinside + count == sdel )
            {
                sinside = 0;
                sindex++;
            }
            if ( cinside + count == cskip )
            {
                cinside = 0;
                if ( cskip == count )
                    carr.removeAt(cindex);
                else
                    cindex++;
            }
        }
        else if ( smut.isSkipMutation() && cmut.isDeleteMutation() )
        {
            int sskip = smut.toSkipMutation().count();
            int cdel = cmut.toDeleteMutation().count();
            int count = qMin(sskip - sinside, cdel - cinside);
            smut.toSkipMutation().setCount( sskip - count );
            if ( sinside + count == sskip )
            {
                sinside = 0;
                if ( sskip == count )
                    sarr.removeAt(sindex);
                else
                    sindex++;
            }
            if ( cinside + count == cdel )
            {
                cinside = 0;
                cindex++;
            }
        }
    }

    if( sindex < sarr.count() )
    {
        errorLog("Client text mutation is shorter than the server version");
        return;
    }
    if( cindex < carr.count() )
    {
        errorLog("Client text mutation is longer than the server version");
        return;
    }
}

void Transformation::xform_pass1(RichTextMutation s, RichTextMutation c)
{
    JSONArray sarr = s.content();
    JSONArray carr = c.content();

    int sindex = 0;
    int cindex = 0;
    int sinside = 0;
    int cinside = 0;

    // Loop until end of one mutation is reached
    while( sindex < sarr.count() || cindex < carr.count() )
    {
        AbstractMutation smut;
        if ( sindex < sarr.count() )
            smut = sarr.at(sindex);
        AbstractMutation cmut;
        if ( cindex < carr.count() )
            cmut = carr.at(cindex);

        //
        // Server insertions go first
        //

        if ( smut.isInsertMutation() )
        {            
            // In the middle of a client skip/delete? Split it
            if ( cinside > 0 )
            {
                if ( cmut.isDeleteMutation() )
                {
                    Q_ASSERT(cmut.toDeleteMutation().count() - cinside > 0);
                    carr.insert(cindex+1, DeleteMutation( cmut.toDeleteMutation().count() - cinside));
                    cmut.toDeleteMutation().setCount(cinside);
                }
                else if ( cmut.isSkipMutation() )
                {
                    Q_ASSERT(cmut.toSkipMutation().count() - cinside > 0);
                    carr.insert(cindex+1, SkipMutation( cmut.toSkipMutation().count() - cinside));
                    cmut.toSkipMutation().setCount(cinside);
                }
                cindex++;
                cinside = 0;
                cmut = carr.at(cindex);
            }

            InsertMutation ins = smut.toInsertMutation();
            if ( !ins.isString() )
            {
                errorLog("Only strings allowed inside a text mutation");
                return;
            }
            sindex++;
            if ( !ins.isString() )
                carr.insert(cindex++, SkipMutation(1));
            // TODO: Empty inserts should not be allowed
            else if ( ins.toString().length() > 0 )
                carr.insert(cindex++, SkipMutation(ins.toString().length()));
            continue;
        }

        //
        // Client insertions go next
        //

        if ( cmut.isInsertMutation() )
        {
            // In the middle of a server skip/delete? Split it
            if ( sinside > 0 )
            {
                if ( smut.isDeleteMutation() )
                {
                    Q_ASSERT(smut.toDeleteMutation().count() - sinside > 0);
                    sarr.insert(sindex+1, DeleteMutation( smut.toDeleteMutation().count() - sinside));
                    smut.toDeleteMutation().setCount(sinside);
                }
                else if ( smut.isSkipMutation() )
                {
                    Q_ASSERT(smut.toSkipMutation().count() - sinside > 0);
                    sarr.insert(sindex+1, SkipMutation( smut.toSkipMutation().count() - sinside));
                    smut.toSkipMutation().setCount(sinside);
                }
                sindex++;
                sinside = 0;
                smut = sarr.at(sindex);
            }

            InsertMutation ins = cmut.toInsertMutation();
            if ( !ins.isString() )
                sarr.insert(sindex++, SkipMutation(1));
            // TODO: Empty inserts should not be allowed
            else if ( ins.toString().length() > 0 )
                sarr.insert(sindex++, SkipMutation(ins.toString().length()));
            cindex++;
            continue;
        }

        // End of at least mutation reached?
        // If we did not reach the end of both, then this is an error that is catched outside the loop.
        if ( sindex == sarr.count() || cindex == carr.count() )
            break;

        //
        // Mutations handled here: Delete, Skip, Insert, Array, Object, Text, RichText
        //

        if ( smut.isDeleteMutation() && cmut.isDeleteMutation() )
        {
            int sdel = smut.toDeleteMutation().count();
            int cdel = cmut.toDeleteMutation().count();
            int del = qMin(sdel - sinside, cdel - cinside);
            smut.toDeleteMutation().setCount(cdel - del);
            smut.toDeleteMutation().setCount(sdel - del);
            if ( sinside + del == sdel )
            {
                sinside = 0;
                if ( sdel == del )
                    sarr.removeAt(sindex);
                else
                    sindex++;
            }
            if ( cinside + del == cdel )
            {
                cinside = 0;
                if ( cdel == del )
                    carr.removeAt(cindex);
                else
                    cindex++;
            }
        }
        else if ( smut.isSkipMutation() && cmut.isSkipMutation() )
        {
            int sskip = smut.toSkipMutation().count();
            int cskip = cmut.toSkipMutation().count();
            int skip = qMin(sskip - sinside, cskip - cinside);
            sinside += skip;
            cinside += skip;
            if ( sinside == sskip )
            {
                sinside = 0;
                sindex++;
            }
            if ( cinside == cskip )
            {
                cinside = 0;
                cindex++;
            }
        }
        else if ( smut.isDeleteMutation() && cmut.isSkipMutation() )
        {
            int sdel = smut.toDeleteMutation().count();
            int cskip = cmut.toSkipMutation().count();
            int count = qMin(sdel - sinside, cskip - cinside);
            cmut.toSkipMutation().setCount( cskip - count );
            if ( sinside + count == sdel )
            {
                sinside = 0;
                sindex++;
            }
            if ( cinside + count == cskip )
            {
                cinside = 0;
                if ( cskip == count )
                    carr.removeAt(cindex);
                else
                    cindex++;
            }
        }
        else if ( smut.isSkipMutation() && cmut.isDeleteMutation() )
        {
            int sskip = smut.toSkipMutation().count();
            int cdel = cmut.toDeleteMutation().count();
            int count = qMin(sskip - sinside, cdel - cinside);
            smut.toSkipMutation().setCount( sskip - count );
            if ( sinside + count == sskip )
            {
                sinside = 0;
                if ( sskip == count )
                    sarr.removeAt(sindex);
                else
                    sindex++;
            }
            if ( cinside + count == cdel )
            {
                cinside = 0;
                cindex++;
            }            
        }
        else if ( smut.isSkipMutation() ) // ... and mutation at the client
        {
            cindex++;
            sinside++;
            if ( smut.toSkipMutation().count() == sinside )
            {
                sinside = 0;
                sindex++;
            }
        }
        else if ( cmut.isSkipMutation() ) // ... and mutation at the srver
        {
            sindex++;
            cinside++;
            if ( cmut.toSkipMutation().count() == cinside )
            {
                cinside = 0;
                cindex++;
            }
        }
        else if ( smut.isDeleteMutation() ) // ... and mutation at the client
        {
            carr.removeAt(cindex);
            sinside++;
            if ( smut.toDeleteMutation().count() == sinside )
            {
                sinside = 0;
                sindex++;
            }
        }
        else if ( cmut.isDeleteMutation() ) // ... and mutation at the server
        {
            sarr.removeAt(sindex);
            cinside++;
            if ( cmut.toDeleteMutation().count() == cinside )
            {
                cinside = 0;
                cindex++;
            }
        }
        else
        {
            if ( smut.isObjectMutation() && cmut.isObjectMutation() )
                xform_pass1(smut.toObjectMutation(), cmut.toObjectMutation() );
            else if ( smut.isArrayMutation() && cmut.isArrayMutation() )
                xform_pass1(smut.toArrayMutation(), cmut.toArrayMutation() );
            else if ( smut.isRichTextMutation() && cmut.isRichTextMutation() )
                xform_pass1(smut.toRichTextMutation(), cmut.toRichTextMutation() );
            else
                errorLog("The mutations are not compatible or not allowed inside an array mutation");
            sarr[sindex++] = smut;
            carr[cindex++] = cmut;
        }
    }

    if( sindex < sarr.count() )
    {
        errorLog("Client text mutation is shorter than the server version");
        return;
    }
    if( cindex < carr.count() )
    {
        errorLog("Client text mutation is longer than the server version");
        return;
    }
}
