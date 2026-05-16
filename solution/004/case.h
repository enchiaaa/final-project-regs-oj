#ifndef _CASE_H_
#define _CASE_H_
#define DOCTEST_CONFIG_IMPLEMENT
#include "URL2.h"
#include "test.h"

TEST_CASE("Case1") {
  {
    // telnet://liquid-hydrocarbon.info:44829/crustulum/cavus/suppono?demergo=constans
    auto url = URL::from(
        "telnet%3A%2F%2Fliquid-hydrocarbon.info%3A44829%"
        "2Fcrustulum%2Fcavus%2Fsuppono%3Fdemergo%3Dconstans");
    CHECK_EQ(url->protocol, "telnet");
    CHECK_EQ(url->hostname, "liquid-hydrocarbon.info");
    CHECK_EQ(url->pathname.size(), 3);
    CHECK_EQ(url->pathname, URL::Segments{"crustulum", "cavus", "suppono"});
    CHECK_EQ(url->query.empty(), false);
    CHECK_EQ(url->query["demergo"], URL::Segments{"constans"});
    CHECK_EQ(url->fragment.empty(), true);

    CHECK_EQ(
        url->location(), "telnet://liquid-hydrocarbon.info:44829/crustulum/"
                         "cavus/suppono?demergo=constans");

    url->fragment = "123456";
    CHECK_EQ(
        url->location(), "telnet://liquid-hydrocarbon.info:44829/crustulum/"
                         "cavus/suppono?demergo=constans#123456");
  }
  {
    // svn://tiny-understanding.biz/solitudo
    auto url = URL::from("svn%3A%2F%2Ftiny-understanding.biz%2Fsolitudo");
    CHECK_EQ(url->protocol, "svn");
    CHECK_EQ(url->hostname, "tiny-understanding.biz");
    CHECK_EQ(url->pathname.size(), 1);
    CHECK_EQ(url->pathname, URL::Segments{"solitudo"});
    CHECK_EQ(url->query.empty(), true);

    CHECK_EQ(url->fragment.empty(), true);

    CHECK_EQ(url->location(), "svn://tiny-understanding.biz/solitudo");
    url->query["key"] = {"value"};
    CHECK_EQ(url->location(), "svn://tiny-understanding.biz/solitudo?key=value");
  }
  {
    // svn://bartoletti:4idJdR5b@true-handover.com:8717/recusandae?porro=stabilis
    auto url = URL::from(
        "svn%3A%2F%2Fbartoletti%3A4idJdR5b%40true-handover.com%"
        "3A8717%2Frecusandae%3Fporro%3Dstabilis");
    CHECK_EQ(url->protocol, "svn");
    CHECK_EQ(url->hostname, "true-handover.com");
    CHECK_EQ(url->pathname.size(), 1);
    CHECK_EQ(url->pathname, URL::Segments{"recusandae"});
    CHECK_EQ(url->query.empty(), false);
    CHECK_EQ(url->query["porro"], URL::Segments{"stabilis"});
    CHECK_EQ(url->fragment.empty(), true);

    CHECK_EQ(
        url->location(), "svn://bartoletti:4idJdR5b@true-handover.com:8717/"
                         "recusandae?porro=stabilis");

    url->username = "user";
    CHECK_EQ(
        url->location(), "svn://user:4idJdR5b@true-handover.com:8717/"
                         "recusandae?porro=stabilis");
  }
}

TEST_CASE("Case2") {

  {
    // ftp://similar-bookend.org/totus/caecus/tepidus#UL
    auto url = URL::from("ftp%3A%2F%2Fsimilar-bookend.org%2Ftotus%2Fcaecus%2Ftepidus%23UL");
    CHECK_EQ(url->protocol, "ftp");
    CHECK_EQ(url->hostname, "similar-bookend.org");
    CHECK_EQ(url->pathname.size(), 3);
    CHECK_EQ(url->pathname, URL::Segments{"totus", "caecus", "tepidus"});
    CHECK_EQ(url->query.empty(), true);

    CHECK_EQ(url->fragment.empty(), false);
    CHECK_EQ(url->fragment, "UL");
    CHECK_EQ(url->location(), "ftp://similar-bookend.org/totus/caecus/tepidus#UL");

    url->protocol = "file";
    CHECK_EQ(url->location(), "file://similar-bookend.org/totus/caecus/tepidus#UL");
  }
  {
    // ssh://wry-recovery.com/nihil/callide
    auto url = URL::from("ssh%3A%2F%2Fwry-recovery.com%2Fnihil%2Fcallide");
    CHECK_EQ(url->protocol, "ssh");
    CHECK_EQ(url->hostname, "wry-recovery.com");
    CHECK_EQ(url->pathname.size(), 2);
    CHECK_EQ(url->pathname, URL::Segments{"nihil", "callide"});
    CHECK_EQ(url->query.empty(), true);

    CHECK_EQ(url->fragment.empty(), true);
    CHECK_EQ(url->fragment, "");
    CHECK_EQ(url->location(), "ssh://wry-recovery.com/nihil/callide");

    url->hostname = "yhchen.org";
    CHECK_EQ(url->location(), "ssh://yhchen.org/nihil/callide");
  }
  {
    // ftp://lumpy-omelet.name:38382/possimus?agnitio=socius&deduco=aveho#@/'Ry4@ffrTwN
    auto url = URL::from(
        "ftp%3A%2F%2Flumpy-omelet.name%3A38382%2Fpossimus%3Fagnitio%"
        "3Dsocius%26deduco%3Daveho%23%40%2F'Ry4%40ffrTwN");
    CHECK_EQ(url->protocol, "ftp");
    CHECK_EQ(url->hostname, "lumpy-omelet.name");
    CHECK_EQ(url->pathname.size(), 1);
    CHECK_EQ(url->pathname, URL::Segments{"possimus"});
    CHECK_EQ(url->query.empty(), false);
    CHECK_EQ(url->query["agnitio"], URL::Segments{"socius"});
    CHECK_EQ(url->query["deduco"], URL::Segments{"aveho"});
    CHECK_EQ(url->fragment.empty(), false);
    CHECK_EQ(url->fragment, "@/'Ry4@ffrTwN");
    CHECK_EQ(
        url->location(), "ftp://lumpy-omelet.name:38382/"
                         "possimus?agnitio=socius&deduco=aveho#@/'Ry4@ffrTwN");
  }
}

TEST_CASE("Case3") {
  {
    // sftp://definite-alliance.info/comburo/summisse/uterque/nesciunt/iM~oyYEkl$/nhW,(67+/SXjhhc?UFAlq='tq&oNZ='9U&vIg=X1hU#[;knT2!vNSO
    auto url = URL::from(
        "sftp%3A%2F%2Fdefinite-alliance.info%2Fcomburo%2Fsummisse%"
        "2Futerque%2Fnesciunt%2FiM~oyYEkl%24%2FnhW%2C(67%2B%2FSXjhhc%"
        "3FUFAlq%3D'tq%26oNZ%3D'9U%26vIg%3DX1hU%23%5B%3BknT2!vNSO");
    CHECK_EQ(url->protocol, "sftp");
    CHECK_EQ(url->hostname, "definite-alliance.info");
    CHECK_EQ(url->pathname.size(), 7);
    CHECK_EQ(url->pathname, URL::Segments{"comburo", "summisse", "uterque", "nesciunt", "iM~oyYEkl$", "nhW,(67+", "SXjhhc"});
    CHECK_EQ(url->query.empty(), false);
    CHECK_EQ(url->query["UFAlq"], URL::Segments{"'tq"});
    CHECK_EQ(url->query["oNZ"], URL::Segments{"'9U"});
    CHECK_EQ(url->query["vIg"], URL::Segments{"X1hU"});
    CHECK_EQ(url->fragment.empty(), false);
    CHECK_EQ(url->fragment, "[;knT2!vNSO");
    CHECK_EQ(
        url->location(), "sftp://definite-alliance.info/comburo/summisse/uterque/nesciunt/"
                         "iM~oyYEkl$/nhW,(67+/SXjhhc?UFAlq='tq&oNZ='9U&vIg=X1hU#[;knT2!vNSO");

    url->fragment = "";
    url->query = {};
    CHECK_EQ(url->location(), "sftp://definite-alliance.info/comburo/summisse/uterque/nesciunt/iM~oyYEkl$/nhW,(67+/SXjhhc");
  }
  {
    // ftps://wiegand:vZEJus@amazing-molasses.com/quidem/cui/crustulum/usitas/titulus/stips#mEwxTgvtFC2I(
    auto url = URL::from(
        "ftps%3A%2F%2Fwiegand%3AvZEJus%40amazing-molasses.com%2Fquidem%"
        "2Fcui%2Fcrustulum%2Fusitas%2Ftitulus%2Fstips%23mEwxTgvtFC2I(");
    CHECK_EQ(url->protocol, "ftps");
    CHECK_EQ(url->hostname, "amazing-molasses.com");
    CHECK_EQ(url->pathname.size(), 6);
    CHECK_EQ(url->pathname, URL::Segments{"quidem", "cui", "crustulum", "usitas", "titulus", "stips"});
    CHECK_EQ(url->query.empty(), true);

    CHECK_EQ(url->fragment.empty(), false);
    CHECK_EQ(url->fragment, "mEwxTgvtFC2I(");
    CHECK_EQ(
        url->location(), "ftps://wiegand:vZEJus@amazing-molasses.com/quidem/"
                         "cui/crustulum/usitas/titulus/stips#mEwxTgvtFC2I(");

    url->pathname = {"path", "to", "new"};
    CHECK_EQ(
        url->location(), "ftps://wiegand:vZEJus@amazing-molasses.com/path/"
                         "to/new#mEwxTgvtFC2I(");
  }
  {
    // ssh://howe:zFGl@exotic-jacket.info/curiositas?2zVQTZ=vQ?ON&MNlVOR=5CMP-&vhAZA=!Y?YrJ&wdW=M,BNN&z1Y=wd+W#Qqe3.:j0g_5_0
    auto url = URL::from(
        "ssh%3A%2F%2Fhowe%3AzFGl%40exotic-jacket.info%2Fcuriositas%"
        "3F2zVQTZ%3DvQ%3FON%26MNlVOR%3D5CMP-%26vhAZA%3D!Y%3FYrJ%26wdW%"
        "3DM%2CBNN%26z1Y%3Dwd%2BW%23Qqe3.%3Aj0g_5_0");
    CHECK_EQ(url->protocol, "ssh");
    CHECK_EQ(url->hostname, "exotic-jacket.info");
    CHECK_EQ(url->pathname.size(), 1);
    CHECK_EQ(url->pathname, URL::Segments{"curiositas"});
    CHECK_EQ(url->query.empty(), false);
    CHECK_EQ(url->query["2zVQTZ"], URL::Segments{"vQ?ON"});
    CHECK_EQ(url->query["MNlVOR"], URL::Segments{"5CMP-"});
    CHECK_EQ(url->query["vhAZA"], URL::Segments{"!Y?YrJ"});
    CHECK_EQ(url->query["wdW"], URL::Segments{"M,BNN"});
    CHECK_EQ(url->query["z1Y"], URL::Segments{"wd+W"});
    CHECK_EQ(url->fragment.empty(), false);
    CHECK_EQ(url->fragment, "Qqe3.:j0g_5_0");
    CHECK_EQ(
        url->location(), "ssh://howe:zFGl@exotic-jacket.info/"
                         "curiositas?2zVQTZ=vQ?ON&MNlVOR=5CMP-&vhAZA=!Y?YrJ&"
                         "wdW=M,BNN&z1Y=wd+W#Qqe3.:j0g_5_0");

    url->fragment = "123";
    CHECK_EQ(
        url->location(),
        "ssh://howe:zFGl@exotic-jacket.info/curiositas?2zVQTZ=vQ?ON&MNlVOR=5CMP-&vhAZA=!Y?YrJ&wdW=M,BNN&z1Y=wd+W#123");
  }
}

TEST_CASE("Case4") {

  {
    // ftp://severe-humor.biz/acies/vester/constans/textor/varietas/ulciscor/sub/odio/C/fa/9K~&urNWzy/ID-W9m?03a5=S)G9
    auto url = URL::from(
        "ftp%3A%2F%2Fsevere-humor.biz%2Facies%2Fvester%"
        "2Fconstans%2Ftextor%2Fvarietas%2Fulciscor%2Fsub%2Fodio%"
        "2FC%2Ffa%2F9K~%26urNWzy%2FID-W9m%3F03a5%3DS)G9");
    CHECK_EQ(url->protocol, "ftp");
    CHECK_EQ(url->hostname, "severe-humor.biz");
    CHECK_EQ(url->pathname.size(), 12);
    CHECK_EQ(
        url->pathname,
        URL::Segments{
            "acies", "vester", "constans", "textor", "varietas", "ulciscor", "sub", "odio", "C", "fa", "9K~&urNWzy", "ID-W9m"});
    CHECK_EQ(url->query.empty(), false);
    CHECK_EQ(url->query["03a5"], URL::Segments{"S)G9"});
    CHECK_EQ(url->fragment.empty(), true);

    CHECK_EQ(
        url->location(), "ftp://severe-humor.biz/acies/vester/constans/textor/varietas/"
                         "ulciscor/sub/odio/C/fa/9K~&urNWzy/ID-W9m?03a5=S)G9");
    url->query = URL::Query{
        {"key", {"a", "b", "c"}}
    };
    url->hostname = "www.example.com";
    CHECK_EQ(
        url->location(), "ftp://www.example.com/acies/vester/constans/textor/varietas/"
                         "ulciscor/sub/odio/C/fa/9K~&urNWzy/ID-W9m?key=a&key=b&key=c");
  }
  {
    // ftps://unconscious-lady.net/comburo/illo/xiphias/aranea
    auto url = URL::from("ftps%3A%2F%2Funconscious-lady.net%2F%2F%2Fcomburo%2Fillo%2Fxiphias%2Faranea");
    CHECK_EQ(url->protocol, "ftps");
    CHECK_EQ(url->hostname, "unconscious-lady.net");
    CHECK_EQ(url->pathname.size(), 4);
    CHECK_EQ(url->pathname, URL::Segments{"comburo", "illo", "xiphias", "aranea"});
    CHECK_EQ(url->query.empty(), true);
    CHECK_EQ(url->fragment.empty(), true);
    CHECK_EQ(url->location(), "ftps://unconscious-lady.net/comburo/illo/xiphias/aranea");
  }
  {
    // http://wordy-cardboard.com:12750/clamo/angelus/esse/sit/modi/beatae/tego/appello?7nsH=!_7e&YLR=.ES)Q2ZV&nIVFQQ=KtZZ_zZ+&uaYg=3au/N&vd4=o;K
    auto url = URL::from(
        "http%3A%2F%2Fwordy-cardboard.com%3A12750%2Fclamo%2Fangelus%2Fesse%2Fsit%"
        "2Fmodi%2Fbeatae%2Ftego%2Fappello%3F7nsH%3D!_7e%26YLR%3D.ES)Q2ZV%"
        "26nIVFQQ%3DKtZZ_zZ%2B%26uaYg%3D3au%2FN%26vd4%3Do%3BK");
    CHECK_EQ(url->protocol, "http");
    CHECK_EQ(url->hostname, "wordy-cardboard.com");
    CHECK_EQ(url->pathname.size(), 8);
    CHECK_EQ(url->pathname, URL::Segments{"clamo", "angelus", "esse", "sit", "modi", "beatae", "tego", "appello"});
    CHECK_EQ(url->query.empty(), false);
    CHECK_EQ(url->query["7nsH"], URL::Segments{"!_7e"});
    CHECK_EQ(url->query["YLR"], URL::Segments{".ES)Q2ZV"});
    CHECK_EQ(url->query["nIVFQQ"], URL::Segments{"KtZZ_zZ+"});
    CHECK_EQ(url->query["uaYg"], URL::Segments{"3au/N"});
    CHECK_EQ(url->query["vd4"], URL::Segments{"o;K"});
    CHECK_EQ(url->fragment.empty(), true);

    CHECK_EQ(
        url->location(), "http://wordy-cardboard.com:12750/clamo/angelus/esse/sit/modi/beatae/"
                         "tego/appello?7nsH=!_7e&YLR=.ES)Q2ZV&nIVFQQ=KtZZ_zZ+&uaYg=3au/N&vd4=o;K");
  }
}

TEST_CASE("Case5") {
  {
    // ftp://blue-jury.net:22071/socius/coniecto/tumultus/nobis/civitas/usus/3pXDg6/6?7kX=5KBS&MmrZ=DExg3&fArZZY=@e-
    auto url = URL::from(
        "ftp%3A%2F%2Fblue-jury.net%3A22071%2Fsocius%2Fconiecto%"
        "2Ftumultus%2Fnobis%2Fcivitas%2Fusus%2F3pXDg6%2F6%3F7kX%"
        "3D5KBS%26MmrZ%3DDExg3%26fArZZY%3D%40e-");
    CHECK_EQ(url->protocol, "ftp");
    CHECK_EQ(url->hostname, "blue-jury.net");
    CHECK_EQ(url->pathname.size(), 8);
    CHECK_EQ(url->pathname, URL::Segments{"socius", "coniecto", "tumultus", "nobis", "civitas", "usus", "3pXDg6", "6"});
    CHECK_EQ(url->query.empty(), false);
    CHECK_EQ(url->query["7kX"], URL::Segments{"5KBS"});
    CHECK_EQ(url->query["MmrZ"], URL::Segments{"DExg3"});
    CHECK_EQ(url->query["fArZZY"], URL::Segments{"@e-"});
    CHECK_EQ(url->fragment.empty(), true);
    CHECK_EQ(url->fragment, "");
    CHECK_EQ(
        url->location(), "ftp://blue-jury.net:22071/socius/coniecto/tumultus/nobis/civitas/"
                         "usus/3pXDg6/6?7kX=5KBS&MmrZ=DExg3&fArZZY=@e-");

    url->port = "1111";
    url->pathname = {"a", "b", "c"};
    url->query = URL::Query{
        {"X", {"!@~$%"}}
    };
    url->fragment = "~!@#$^&*()_=%";
    CHECK_EQ(url->location(), "ftp://blue-jury.net:1111/a/b/c?X=!@~$%#~!@#$^&*()_=%");
  }
  {
    // ftps://small-scorn.net/celebrer/spoliatio/tolero#:
    auto url = URL::from("ftps%3A%2F%2Fsmall-scorn.net%2Fcelebrer%2Fspoliatio%2Ftolero%23%3A");
    CHECK_EQ(url->protocol, "ftps");
    CHECK_EQ(url->hostname, "small-scorn.net");
    CHECK_EQ(url->pathname.size(), 3);
    CHECK_EQ(url->pathname, URL::Segments{"celebrer", "spoliatio", "tolero"});
    CHECK_EQ(url->query.empty(), true);

    CHECK_EQ(url->fragment.empty(), false);
    CHECK_EQ(url->fragment, ":");
    CHECK_EQ(url->location(), "ftps://small-scorn.net/celebrer/spoliatio/tolero#:");
  }
  {
    // ldap://impassioned-disclosure.org:28935/umerus/perferendis/sui/a/triduana#yQLXEB5(GN
    auto url = URL::from(
        "ldap%3A%2F%2Fimpassioned-disclosure.org%3A28935%2Fumerus%"
        "2Fperferendis%2Fsui%2Fa%2Ftriduana%23yQLXEB5(GN");
    CHECK_EQ(url->protocol, "ldap");
    CHECK_EQ(url->hostname, "impassioned-disclosure.org");
    CHECK_EQ(url->pathname.size(), 5);
    CHECK_EQ(url->pathname, URL::Segments{"umerus", "perferendis", "sui", "a", "triduana"});
    CHECK_EQ(url->query.empty(), true);

    CHECK_EQ(url->fragment.empty(), false);
    CHECK_EQ(url->fragment, "yQLXEB5(GN");
    CHECK_EQ(
        url->location(), "ldap://impassioned-disclosure.org:28935/umerus/"
                         "perferendis/sui/a/triduana#yQLXEB5(GN");
    url->port = "0";

    CHECK_EQ(
        url->location(), "ldap://impassioned-disclosure.org/umerus/"
                         "perferendis/sui/a/triduana#yQLXEB5(GN");
  }
}

#endif // _CASE_H_
