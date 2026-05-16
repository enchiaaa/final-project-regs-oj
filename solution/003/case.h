#ifndef _CASE_H_
#define _CASE_H_
#define DOCTEST_CONFIG_IMPLEMENT
#include "School.h"
#include "test.h"
#include <iostream>
#include <sstream>
#include <string>

inline std::string trimEnd(std::string s) {
  s.erase(s.find_last_not_of("\t\n\r\f\v ") + 1);
  return s;
}

TEST_CASE("Case1") {
  std::ostringstream oss;
  std::streambuf* cout_buf = std::cout.rdbuf(oss.rdbuf());
  School ntust("NTUST", 10000);
  School ntust2("NTUST2", 20000);

  std::cout << ntust << std::endl;
  std::cout << ntust2 << std::endl;

  ntust.admissions(1000);
  std::cout << ntust << std::endl;

  ntust.dropouts(1000);
  std::cout << ntust << std::endl;

  ntust.transfer(1000, ntust2);
  std::cout << ntust << std::endl;
  std::cout << ntust2 << std::endl;
  std::cout.rdbuf(cout_buf);
  auto output = "NTUST	10000	10000\n"
                "NTUST2	20000	20000\n"
                "NTUST	11000	10000\n"
                "NTUST	10000	10000\n"
                "NTUST	9000	10000\n"
                "NTUST2	21000	20000";
  CHECK_EQ(trimEnd(oss.str()) , trimEnd(output));
}

TEST_CASE("Case2") {
  std::ostringstream oss;
  std::streambuf* cout_buf = std::cout.rdbuf(oss.rdbuf());
  // PrivateSchool and PublicSchool Classes Test
  PrivateSchool fjcu("FJCU", 10000);
  PublicSchool ntut("NTUT", 20000);

  std::cout << fjcu << std::endl;
  std::cout << ntut << std::endl;

  fjcu.admissions(1000);
  ntut.admissions(1000);

  std::cout << fjcu << std::endl;
  std::cout << ntut << std::endl;

  fjcu.dropouts(1000);
  ntut.dropouts(1000);

  std::cout << fjcu << std::endl;
  std::cout << ntut << std::endl;

  fjcu.dropouts(1000);
  ntut.dropouts(1000);

  std::cout << fjcu << std::endl;
  std::cout << ntut << std::endl;

  ntut.transfer(500, fjcu);

  std::cout << fjcu << std::endl;
  std::cout << ntut << std::endl;

  fjcu.transfer(100, ntut);

  std::cout << fjcu << std::endl;
  std::cout << ntut << std::endl;
  std::cout.rdbuf(cout_buf);
  auto output = "FJCU	10000	10000\n"
                "NTUT	20000	20000\n"
                "FJCU	11000	10000\n"
                "NTUT	21000	20000\n"
                "FJCU	10000	10000\n"
                "NTUT	20000	19000\n"
                "FJCU	9000	9900\n"
                "NTUT	19000	18050\n"
                "FJCU	9500	9900\n"
                "NTUT	18500	17147\n"
                "FJCU	9400	9800\n"
                "NTUT	18600	17147";
  CHECK_EQ(trimEnd(oss.str()) , trimEnd(output));
}

TEST_CASE("Case3") {
  std::ostringstream oss;
  std::streambuf* cout_buf = std::cout.rdbuf(oss.rdbuf());
  PrivateSchool fjcu("FJCU", 10000);

  PublicSchool ntut("NTUT", 10000);

  std::cout << fjcu << std::endl;

  fjcu.dropouts(1);
  std::cout << fjcu << std::endl;

  fjcu.dropouts(99);
  std::cout << fjcu << std::endl;

  fjcu.dropouts(100);
  std::cout << fjcu << std::endl;

  std::cout << ntut << std::endl;

  ntut.dropouts(99);
  std::cout << ntut << std::endl;

  ntut.dropouts(101);
  std::cout << ntut << std::endl;

  ntut.dropouts(100);
  std::cout << ntut << std::endl;
  std::cout.rdbuf(cout_buf);
  auto output = "FJCU	10000	10000\n"
                "FJCU	9999	10000\n"
                "FJCU	9900	9900\n"
                "FJCU	9800	9800\n"
                "NTUT	10000	10000\n"
                "NTUT	9901	10000\n"
                "NTUT	9800	9500\n"
                "NTUT	9700	9500\n";
  CHECK_EQ(trimEnd(oss.str()) , trimEnd(output));
}

TEST_CASE("Case4") {
  std::ostringstream oss;
  std::streambuf* cout_buf = std::cout.rdbuf(oss.rdbuf());
  School ntust("NTUST", 12500);
  PublicSchool ntut("NTUT", 85000);
  PrivateSchool fjcu("FJCU", 25000);

  std::cout << ntust << std::endl;
  std::cout << ntut << std::endl;
  std::cout << fjcu << std::endl;

  ntust.admissions(200);
  std::cout << ntust << std::endl;

  ntust.dropouts(200);
  std::cout << ntust << std::endl;

  ntust.dropouts(100000);
  std::cout << ntust << std::endl;

  fjcu.admissions(1000);
  std::cout << fjcu << std::endl;

  fjcu.admissions(-1000);
  std::cout << fjcu << std::endl;

  fjcu.dropouts(50);
  std::cout << fjcu << std::endl;

  fjcu.dropouts(1000);
  std::cout << fjcu << std::endl;

  ntut.admissions(1000);
  std::cout << ntut << std::endl;

  ntut.apply_growth();
  std::cout << ntut << std::endl;

  ntut.dropouts(1000);
  std::cout << ntut << std::endl;

  std::cout << ntut << std::endl;
  ntut.transfer(1000, ntust);
  std::cout << ntut << std::endl;
  std::cout << ntust << std::endl;

  fjcu.transfer(30000, ntust);
  std::cout << fjcu << std::endl;
  std::cout << ntust << std::endl;
  std::cout.rdbuf(cout_buf);

  auto output = "NTUST	12500	12500\n"
                "NTUT	85000	85000\n"
                "FJCU	25000	25000\n"
                "NTUST	12700	12500\n"
                "NTUST	12500	12500\n"
                "Invalid student count\n"
                "NTUST	12500	12500\n"
                "FJCU	26000	25000\n"
                "Invalid amount\n"
                "FJCU	26000	25000\n"
                "FJCU	25950	25000\n"
                "FJCU	24950	24900\n"
                "NTUT	86000	85000\n"
                "NTUT	86000	89250\n"
                "NTUT	85000	84787\n"
                "NTUT	85000	84787\n"
                "NTUT	84000	80547\n"
                "NTUST	13500	12500\n"
                "Invalid student count\n"
                "FJCU	24950	24900\n"
                "NTUST	13500	12500";
  CHECK_EQ(trimEnd(oss.str()) , trimEnd(output));
}

TEST_CASE("Case5") {
  std::ostringstream oss;
  std::streambuf* cout_buf = std::cout.rdbuf(oss.rdbuf());
  PublicSchool test("TEST", 2222);
  std::cout << test << std::endl;

  test.admissions(2879);
  std::cout << test << std::endl;

  test.dropouts(101);
  std::cout << test << std::endl;

  PublicSchool test2("TEST2", 2879);
  std::cout << test2 << std::endl;

  test2.admissions(2222);
  std::cout << test2 << std::endl;

  test2.dropouts(101);
  std::cout << test2 << std::endl;

  PrivateSchool test3("TEST3", 1000);
  std::cout << test3 << std::endl;

  test3.admissions(0);
  std::cout << test3 << std::endl;

  test3.dropouts(-10000);
  std::cout << test3 << std::endl;

  test3.admissions(9000);
  std::cout << test3 << std::endl;

  test3.dropouts(100000);
  std::cout << test3 << std::endl;

  test3.transfer(0.5, test);
  std::cout << test3 << std::endl;
  std::cout << test << std::endl;

  test3.transfer(1, test);
  std::cout << test3 << std::endl;
  std::cout << test << std::endl;

  test3.transfer(0, test);
  std::cout << test3 << std::endl;
  std::cout << test << std::endl;

  test.transfer(100000, test3);
  std::cout << test3 << std::endl;
  std::cout << test << std::endl;
  std::cout.rdbuf(cout_buf);

  auto output = "TEST	2222	2222\n"
                "TEST	5101	2222\n"
                "TEST	5000	2110\n"
                "TEST2	2879	2879\n"
                "TEST2	5101	2879\n"
                "TEST2	5000	2735\n"
                "TEST3	1000	1000\n"
                "Invalid amount\n"
                "TEST3	1000	1000\n"
                "Invalid amount\n"
                "TEST3	1000	1000\n"
                "TEST3	10000	1000\n"
                "Invalid student count\n"
                "TEST3	10000	1000\n"
                "Invalid amount\n"
                "TEST3	10000	1000\n"
                "TEST	5000	2110\n"
                "TEST3	9999	1000\n"
                "TEST	5001	2110\n"
                "Invalid amount\n"
                "TEST3	9999	1000\n"
                "TEST	5001	2110\n"
                "Invalid student count\n"
                "TEST3	9999	1000\n"
                "TEST	5001	2110";
  CHECK_EQ(trimEnd(oss.str()) , trimEnd(output));
}

#endif
