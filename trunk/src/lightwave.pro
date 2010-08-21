# -------------------------------------------------
# Project created by QtCreator 2010-07-22T01:41:45
# -------------------------------------------------
QT += network \
    script
QT -= gui
TARGET = lightwave
CONFIG += console
CONFIG -= app_bundle
TEMPLATE = app
SOURCES += main.cpp \
    json/jsonabstractobject.cpp \
    json/jsonobject.cpp \
    json/jsonarray.cpp \
    json/jsonconstant.cpp \
    json/jsonscanner.cpp \
    ot/documentmutation.cpp \
    ot/objectmutation.cpp \
    ot/abstractmutation.cpp \
    ot/arraymutation.cpp \
    ot/textmutation.cpp \
    ot/skipmutation.cpp \
    ot/deletemutation.cpp \
    ot/liftmutation.cpp \
    ot/squeezemutation.cpp \
    ot/transformation.cpp \
    ot/insertmutation.cpp \
    test/randomdocgenerator.cpp \
    test/randommutationgenerator.cpp \
    ot/richtextmutation.cpp \
    fcgi/fcgiserver.cpp \
    fcgi/fcgirequest.cpp \
    fcgi/fcgiprotocol.cpp \
    utils/settings.cpp \
    wave/waveprovider.cpp \
    wave/wavecontainer.cpp \
    wave/wavedocument.cpp \
    wave/waverootdocument.cpp \
    wave/session.cpp \
    utils/getopts.cpp \
    utils/jid.cpp \
    wave/waveid.cpp \
    wave/hostcontainer.cpp \
    wave/rootcontainer.cpp \
    wave/sessioncontainer.cpp \
    wave/viewcontainer.cpp \
    wave/view.cpp \
    wave/usercontainer.cpp \
    wave/user.cpp \
    auth/auth.cpp \
    js/jsengine.cpp
HEADERS += json/jsonabstractobject.h \
    json/jsonobject.h \
    json/jsonarray.h \
    json/jsonconstant.h \
    json/jsonscanner.h \
    ot/documentmutation.h \
    ot/objectmutation.h \
    ot/abstractmutation.h \
    ot/arraymutation.h \
    ot/textmutation.h \
    ot/skipmutation.h \
    ot/deletemutation.h \
    ot/liftmutation.h \
    ot/squeezemutation.h \
    ot/overwritemutation.h \
    ot/transformation.h \
    ot/insertmutation.h \
    test/randomdocgenerator.h \
    test/randommutationgenerator.h \
    ot/richtextmutation.h \
    fcgi/fcgiserver.h \
    fcgi/fcgirequest.h \
    fcgi/fcgiprotocol.h \
    fcgi/fcgi.h \
    utils/settings.h \
    wave/waveprovider.h \
    wave/wavecontainer.h \
    wave/wavedocument.h \
    wave/waverootdocument.h \
    utils/session.h \
    utils/getopts.h \
    utils/jid.h \
    wave/waveid.h \
    wave/hostcontainer.h \
    wave/rootcontainer.h \
    wave/session.h \
    wave/sessioncontainer.h \
    wave/viewcontainer.h \
    wave/view.h \
    wave/usercontainer.h \
    wave/user.h \
    auth/auth.h \
    js/jsengine.h
