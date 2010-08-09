#include <QtCore/QCoreApplication>
//#include "jsonobject.h"
//#include "jsonarray.h"
//#include "jsonscanner.h"
//#include "documentmutation.h"
//#include "abstractmutation.h"
//#include "transformation.h"
//#include "randomdocgenerator.h"
//#include "randommutationgenerator.h"
//#include "objectmutation.h"
#include "utils/settings.h"
#include "fcgi/fcgiserver.h"
#include "utils/getopts.h"
#include <stdio.h>
#include <QTextStream>

enum
{
    Option_Port = 1,
    Option_Domain = 2,
    Option_Help = 3
};

int main(int argc, char *argv[])
{
    QCoreApplication a(argc, argv);

    QTextStream out(stdout);

    QOptions options(argc, (const char**)argv);
    options.setProgramName("lightwave");
    options.setMaxNonOptionalValues(0);
    options.setMinNonOptionalValues(0);
    QOption o1( Option_Port, "port" );
    o1.setShortCode('p');
    o1.setNeedsValue( true, "9871", "port_number" );
    o1.setDescription("The FCGI port used by the web server to talk to lightwave");
    options.addOption( o1 );
    QOption o2( Option_Domain, "domain" );
    o2.setShortCode('d');
    o2.setNeedsValue( true, "localhost", "domain" );
    o2.setDescription("The domain served by the lightwave server");
    options.addOption( o2 );
    QOption o3( Option_Help, "help" );
    o3.setShortCode('h');
    o3.setDescription("Print help page");
    options.addOption( o3 );

    if ( !options.parse() )
    {
        options.printError(out);
        return 1;
    }
    if ( options.option(Option_Help).occurrence() )
    {
        options.printHelp(out);
        return 1;
    }

    // Get the settings
    QString profile = "./waveserver.conf";
    if ( argc == 2 )
        profile = QString(argv[1]);
    Settings settings( profile );

    if ( options.option(Option_Port).occurrence() )
        settings.setFcgiPort(options.option(Option_Port).value().toInt() );
    if ( options.option(Option_Domain).occurrence() )
        settings.setDomain(options.option(Option_Domain).value());

    FCGI::FCGIServer server;

    return a.exec();

//    JSONObject obj;
//    obj.setAttribute("name", "Torben");
//    obj.setAttribute("age", 35);
//    JSONArray arr;
//    arr.append("Heinz");
//    arr.append("Franz");
//    obj.setAttribute("friends", arr);
//    JSONObject o2;
//    o2.setAttribute("white", true);
//    o2.setAttribute("height", 1.73);
//    obj.setAttribute("props", o2);
//
//    out << obj.toJSON() << endl;
//
//    const char* str = "{\"friends\":[\"Heinz\",\"Franz\",\"Fritz\"],\"age\":35,\"props\":{\"white\":true,\"height\":1.73},\"name\":\"Torben\"}";
//    JSONScanner scanner(str, strlen(str));
//    bool ok;
//    JSONObject res = scanner.scan(&ok);
//    if ( !ok )
//        out << "Scanner error" << endl;
//    else
//        out << res.toJSON() << endl;

//    const char* mut = "{\"actions\":[ {\"object\":\"props\", \"actions\":[ { \"key\":\"white\", \"value\":\"false\" }, { \"key\":\"male\", \"value\":\"yes\" } ] }, { \"key\":\"age\", \"value\":36 }, {\"array\":\"friends\", \"actions\":[ {\"skip\":1 }, {\"delete\":1}, {\"insert\":\"Hannes\" } ] } ] }";
//    JSONScanner scanner2(mut, strlen(mut));
//    DocumentMutation mutation( scanner2.scan(&ok) );
//    if ( !ok )
//        out << "Scanner error" << endl;
//    else
//        out << mutation.toJSON() << endl;


    // {"actions":[ {"object":"props", "actions":[ { "key":"white", "value":"false" }, { "key":"male", "value":"yes" } ] }, { "key":"age", "value":36 },
    //              {"array":"friends", "actions":[ {"skip":1 }, {"delete":1}, {"insert":"Hannes" } ] } ] }

    // {"actions":[ {"text":"name", "actions":[ {"delete":1}, {"insert":"t" } ] } ] }

    // {"actions":[ {"remove":"props"} ] }
    // {"actions":[ {"remove":"age"} ] }
    // {"actions":[ {"object":"props", "replace":true, "actions":[ { "value":"white", "value":"false" } ] } ] }

    // {"chat": [ "Hello", {"tag":"image", "url":"http..."}, "how are you?" ] }
    // {"actions":[ {"array":"chat", "actions":[ {"text":[ {"skip":5}, {"insert":"!"} ] } ] } ] }

    // {"list": [ 1,2,3,4,5 ] }
    // {"actions":[ {"array":"list", actions:[ {"lift":"id1234"}, {"skip":3}, {"squeeze":"id1234"} ] } ] }

//    const char* m1 = "{\"_object\":true, \"foo\":100, \"bar\":300, \"obj\":{\"_object\":true, \"name\":\"torben\", \"age\":37, \"props\":{\"doof\":\"server\"}, \"list\":[1,2,3,4] }, \"txt\":{\"_text\":[ \"Hallo\" ] } }";
//    const char* m2 = "{\"_object\":true, \"foo\":200, \"snoop\":400, \"obj\":{\"_object\":true, \"name\":\"Torben\", \"props\":{\"irre\":\"Wahnsinn\"}, \"list\":[5,6,7,8] }, \"txt\":{\"_text\":[\"Welt\"] } }";

//    const char* m1 = "{\"_object\":true, \"foo\":{\"_array\":[1,2,3,{\"_delete\":5},{\"_object\":true, \"x\":5,\"y\":6}]}}";
//    const char* m2 = "{\"_object\":true, \"foo\":{\"_array\":[{\"_skip\":3},41,51,{\"_delete\":1},{\"_object\":true, \"bar\":100},{\"_object\":true, \"x\":50,\"z\":60}]}}";

//    const char* m2 = "{\"_object\":true, \"foo\":{\"_array\":[{\"_object\":true, \"x\":10},{\"_skip\":2}]}}";
//    const char* m1 = "{\"_object\":true, \"foo\":{\"_array\":[{\"_lift\":\"L1\"}, {\"_skip\":2}, {\"_squeeze\":\"L1\"}]}}";

//    const char* m2 = "{\"_object\":true, \"foo\":{\"_array\":[{\"_object\":true, \"x\":10, \"y\":20},{\"_skip\":3}]}}";
//    const char* m1 = "{\"_object\":true, \"foo\":{\"_array\":[{\"_lift\":\"L1\", \"_mutation\":{ \"_object\":true, \"y\":200, \"z\":300 }}, {\"_skip\":2}, {\"_squeeze\":\"L1\"}, {\"_skip\":1}]}}";

//    const char* m2 = "{\"_object\":true, \"foo\":{\"_array\":[{\"_delete\":1},{\"_skip\":2}]}}";
//    const char* m1 = "{\"_object\":true, \"foo\":{\"_array\":[{\"_lift\":\"L1\"}, {\"_skip\":2}, {\"_squeeze\":\"L1\"}]}}";

//    const char* m2 = "{\"_object\":true, \"foo\":{\"_array\":[{\"_skip\":3}]}}";
//    const char* m1 = "{\"_object\":true, \"foo\":{\"_array\":[{\"_lift\":\"L1\"}, {\"_skip\":2}, {\"_squeeze\":\"L1\"}]}}";

//    const char* m2 = "{\"_object\":true, \"foo\":{\"_array\":[{\"_lift\":\"L2\"}, {\"_skip\":1}, {\"_squeeze\":\"L2\"}, {\"_skip\":1}]}}";
//    const char* m1 = "{\"_object\":true, \"foo\":{\"_array\":[{\"_lift\":\"L1\"}, {\"_skip\":2}, {\"_squeeze\":\"L1\"}]}}";

//    const char* m1 = "{\"_object\":true, \"foo\":{\"_array\":[{\"_lift\":\"L1\"}, {\"_skip\":2}]}, \"gonzo\":{\"_array\":[{\"_squeeze\":\"L1\"}]}}";
//    const char* m2 = "{\"_object\":true, \"foo\":{\"_array\":[{\"_lift\":\"L2\", \"_mutation\":{\"_object\":true, \"fuzzy\":34}}, {\"_skip\":2}]}, \"bar\":{\"_array\":[1,2,{\"_squeeze\":\"L2\"},{\"_object\":true, \"holla\":123}]}}";

//    const char* m1 = "{\"_object\":true, \"foo\":{\"_array\":[{\"_lift\":\"L1\", \"_mutation\":{\"_array\":[{\"_squeeze\":\"L2\"}]}}, {\"_lift\":\"L2\"}]}, \"gonzo\":{\"_array\":[{\"_squeeze\":\"L1\"}]}}";
//    const char* m2 = "{\"_object\":true, \"foo\":{\"_array\":[{\"_lift\":\"L1\"}, {\"_lift\":\"L2\", \"_mutation\":{\"_object\":true, \"hudel\":\"dudel\"}}, 12, {\"_squeeze\":\"L2\"}, {\"_squeeze\":\"L1\"}]}}";

//    const char* m1 = "{\"_object\":true, \"foo\":{\"_lift\":\"L1\"}, \"bar\":{\"_squeeze\":\"L1\"}}";
//    const char* m2 = "{\"_object\":true, \"foo\":{\"_array\":[1,2,{\"_skip\":2},5,6]}}";
//
//    JSONScanner scanner3(m1, strlen(m1));
//    AbstractMutation mutation1( scanner3.scan(&ok) );
//    if ( !ok )
//        out << "Scanner error" << endl;
//    else
//        out << "Server: " << mutation1.toJSON() << endl;
//
//    JSONScanner scanner4(m2, strlen(m2));
//    AbstractMutation mutation2( scanner4.scan(&ok) );
//    if ( !ok )
//        out << "Scanner error" << endl;
//    else
//        out << "Client: " << mutation2.toJSON() << endl;
//
//    Transformation t;
//    t.xform(mutation1, mutation2);
//    out << "Server: " << mutation1.toJSON() << endl;
//    out << "Client: " << mutation2.toJSON() << endl;
//
//    out << "===========================================" << endl;
//
//    const char* doc = "{\"string\":\"xxxWelt\", \"men\":[\"Einstein\", \"Heisenberg\", \"Wirth\", \"Galileo\"], \"friends\":[\"Heinz\",\"Franz\",\"Fritz\"],\"age\":35,\"props\":{\"white\":true,\"height\":1.73},\"name\":\"Torben\"}";
//    JSONScanner scanner1(doc, strlen(doc));
//    JSONObject obj( scanner1.scan(&ok) );
//    if ( !ok )
//        out << "Scanner error" << endl;
//    else
//        out << "Doc: " << obj.toJSON() << endl;
//
//    const char* mut = "{\"_object\":true, \"string\":{\"_text\":[{\"_delete\":3}, \"Hallo \", {\"_skip\":4}, \"!\"]}, \"name\":{\"_lift\":\"N\"}, \"newname\":{\"_squeeze\":\"N\"}, \"men\":{\"_array\":[{\"_squeeze\":\"G\"},{\"_skip\":3},{\"_lift\":\"G\"}]}, \"age\":40, \"friends\":{\"_array\":[{\"_skip\":1}, {\"_delete\":1}, {\"_skip\":1}, \"Georg\"]}, \"foo\":[1,2,3,4], \"props\":{\"_object\":true, \"white\":false}}";
//    JSONScanner scanner2(mut, strlen(mut));
//    DocumentMutation mutation( ObjectMutation( scanner2.scan(&ok) ) );
//    if ( !ok )
//        out << "Scanner error" << endl;
//    else
//        out << "Mutation: " << mutation.mutation().toJSON() << endl;
//
//    obj = mutation.apply(obj, &ok);
//    if ( !ok )
//        out << "Application error" << endl;
//    else
//        out << "Doc2: " << obj.toJSON() << endl;
//
//    out << "===========================================" << endl;

//    // const char* doc = "{\"a\":{\"text\":[\"Hallo Welt\"], \"style\":[{\"count\":6}, {\"weight\":\"bold\", \"count\":4}]}}";
//    const char* doc = "{\"a\":{\"_r\":[\"Hallo\", {\"newline\":true}, {\"_format\":{\"weight\":\"bold\"}}, \"Welt\"]}}";
//    JSONScanner scanner1(doc, strlen(doc));
//    JSONObject obj( scanner1.scan(&ok) );
//    if ( !ok )
//        out << "Scanner error" << endl;
//    else
//        out << "Doc: " << obj.toJSON() << endl;
//
//    const char* mut = "{\"_object\":true, \"a\":{\"_richtext\":[{\"_format\":{\"fontsize\":21}}, {\"_skip\":10}]}}";
//    JSONScanner scanner2(mut, strlen(mut));
//    DocumentMutation mutation( ObjectMutation( scanner2.scan(&ok) ) );
//    if ( !ok )
//        out << "Scanner error" << endl;
//    else
//        out << "Mutation: " << mutation.mutation().toJSON() << endl;
//
//    obj = mutation.apply(obj, &ok);
//    if ( !ok )
//        out << "Application error" << endl;
//    else
//        out << "Doc2: " << obj.toJSON() << endl;
//
//    out << "===========================================" << endl;
//
//    return 0;

//    int lifts = 5;
//    qsrand(9);
//    RandomDocGenerator rand(1);
//    RandomMutationGenerator rmut(lifts);
//    for( int i = 0; i < 300000; ++i )
//    {
//        JSONObject r = rand.createObject(0);
//        out << "Doc " << i << ": " << r.toJSON() << endl;
//
//        DocumentMutation d1 = rmut.createDocumentMutation(r);
//        out << "M1: " << d1.mutation().toJSON() << endl;
//
//        DocumentMutation d2 = rmut.createDocumentMutation(r);
//        out << "M2: " << d2.mutation().toJSON() << endl;
//
//        Transformation t;
//        ObjectMutation m1b( d1.mutation().clone() );
//        ObjectMutation m2b( d2.mutation().clone() );
//        t.xform(m1b, m2b);
//        if ( t.hasError() )
//        {
//            out << t.errorText() << endl;
//            qFatal("Error in transformation");
//        }
//        out << "M1': " << m1b.toJSON() << endl;
//        out << "M2': " << m2b.toJSON() << endl;
//
//        JSONObject r2( r.clone().toObject() );
//        DocumentMutation d1b( m1b );
//        DocumentMutation d2b( m2b );
//        d1.apply(r, &ok);
//        if ( !ok )
//            qFatal("d1: Application error");
//        else
//            out << "d1: " << r.toJSON() << endl;
//        d2b.apply(r, &ok);
//        if ( !ok )
//            qFatal("d1b: Application error");
//        else
//            out << "d1b: " << r.toJSON() << endl;
//
//        d2.apply(r2, &ok);
//        if ( !ok )
//            qFatal("d2: Application error");
//        else
//            out << "d2: " << r2.toJSON() << endl;
//        d1b.apply(r2, &ok);
//        if ( !ok )
//            qFatal("d2b: Application error");
//        else
//            out << "d2b: " << r2.toJSON() << endl;
//
//        if ( !r.equals(r2))
//            qFatal("The two resulting documents differ");
//        out << "===========================================" << endl;
//    }
//    out << "Success" << endl;

//    const char* m1 = "{\"_object\":true, \"a\":{\"_richtext\":[ \"A\", {\"_skip\":2},{\"_format\":{\"a\":1, \"b\":2}}, \"abc\" ]}}";
//    // const char* m1 = "{\"_object\":true, \"a\":{\"_richtext\":[ {\"_skip\":2}, \"abc\" ]}}";
//    const char* m2 = "{\"_object\":true, \"a\":{\"_richtext\":[ \"B\", {\"_format\":{\"c\":3, \"b\":4}}, \"C\", {\"_skip\":2}, \"xyz\" ]}}";

//    const char* m1 = "{\"_object\":true,\"a0\":{\"_richtext\":[\"n\",{\"_skip\":1},\"gv\",{\"_skip\":1},{\"_skip\":1},{\"_format\":{\"sb\":\"up\",\"sc\":\"f\"}},{\"_delete\":1},{\"_delete\":1},{\"_format\":{}},{\"_skip\":1},{\"_format\":{\"sb\":\"\",\"sa\":\"v\"}},{\"_skip\":1},{\"_skip\":1},{\"_delete\":1},\"\",{\"_skip\":1},\"m\",{\"_format\":{\"sa\":\"gl\"}},{\"_delete\":1},{\"_format\":{\"sb\":\"\",\"sc\":\"m\"}},{\"_delete\":1}]}}";
//    const char* m2 = "{\"_object\":true,\"a0\":{\"_richtext\":[{\"_format\":{}},{\"_skip\":1},{\"_format\":{\"sb\":\"yt\",\"sa\":\"\"}},{\"_delete\":1},{\"_skip\":1},{\"_format\":{\"sa\":\"\"}},{\"_skip\":1},{\"_format\":{\"sb\":\"irt\",\"sa\":null}},\"ba\",\"wfq\",{\"_format\":{\"sa\":null}},\"ftl\",\"elq\",{\"_skip\":1},{\"_skip\":1},{\"_skip\":1},{\"_skip\":1},{\"_skip\":1},{\"_delete\":1},{\"_skip\":1},\"g\",{\"_skip\":1}]}}";
//
//    const char* m1 = "{\"_object\":true, \"a\":{\"_array\":[{\"_squeeze\":\"AL0\"},{\"_delete\":1},{\"_text\":[{\"_skip\":1},{\"_skip\":1}]},\"nqh\",{\"_delete\":1},{\"_lift\":\"AL0\"},{\"_delete\":1},{\"_delete\":1}]}}";
//    const char* m2 = "{\"_object\":true, \"a\":{\"_array\":[{\"_lift\":\"AL2\",\"_mutation\":{\"_text\":[\"\",{\"_skip\":1},{\"_skip\":1},{\"_delete\":1}]}},{\"_text\":[{\"_skip\":1},\"\",{\"_skip\":1}]},{\"_lift\":\"AL1\",\"_mutation\":{\"_text\":[{\"_skip\":1}]}},\"na\",{\"_lift\":\"AL0\",\"_mutation\":{\"_text\":[{\"_skip\":1},\"jpp\",\"\",\"\",\"if\",\"una\",\"y\",\"c\",{\"_delete\":1},\"jz\",\"g\",{\"_skip\":1}]}},{\"_squeeze\":\"AL2\"},{\"_squeeze\":\"AL1\"},{\"_squeeze\":\"AL0\"},{\"_delete\":1},{\"_text\":[\"i\",{\"_delete\":1},{\"_delete\":1}]}]}}";
//
//        JSONScanner scanner3(m1, strlen(m1));
//        AbstractMutation mutation1( scanner3.scan(&ok) );
//        if ( !ok )
//            out << "Scanner error" << endl;
//        else
//            out << "Server: " << mutation1.toJSON() << endl;
//
//        JSONScanner scanner4(m2, strlen(m2));
//        AbstractMutation mutation2( scanner4.scan(&ok) );
//        if ( !ok )
//            out << "Scanner error" << endl;
//        else
//            out << "Client: " << mutation2.toJSON() << endl;
//
//        DocumentMutation d1(mutation1.clone());
//        DocumentMutation d2(mutation2.clone());
//
//        Transformation t;
//        t.xform(mutation1, mutation2);
//        if ( t.hasError())
//            out << "Transformation error: " << t.errorText() << endl;
//        out << "Server': " << mutation1.toJSON() << endl;
//        out << "Client': " << mutation2.toJSON() << endl;
//
//        DocumentMutation d1b(mutation1);
//        DocumentMutation d2b(mutation2);

//
//        const char* doc = "{\"a\":{\"_r\":[\"XY\"]}}";
//        JSONScanner scanner1(doc, strlen(doc));
//        JSONObject obj( scanner1.scan(&ok) );
//        if ( !ok )
//            out << "Scanner error" << endl;
//        else
//            out << "Doc: " << obj.toJSON() << endl;
//        JSONObject obj2 = obj.clone().toObject();
//
//        obj = d1.apply(obj, &ok);
//        if ( !ok )
//            out << "Application error" << endl;
//        else
//            out << "Doc1a: " << obj.toJSON() << endl;
//        obj = d2b.apply(obj, &ok);
//        if ( !ok )
//            out << "Application error" << endl;
//        else
//            out << "Doc1b: " << obj.toJSON() << endl;
//
//        obj2 = d2.apply(obj2, &ok);
//        if ( !ok )
//            out << "Application error" << endl;
//        else
//            out << "Doc2a: " << obj2.toJSON() << endl;
//        obj2 = d1b.apply(obj2, &ok);
//        if ( !ok )
//            out << "Application error" << endl;
//        else
//            out << "Doc2b: " << obj2.toJSON() << endl;
//
//        out << "===========================================" << endl;

}
