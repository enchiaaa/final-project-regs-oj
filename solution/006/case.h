#ifndef _CASE_H_
#define _CASE_H_
#define DOCTEST_CONFIG_IMPLEMENT
#include "operator.h"
#include "test.h"
#include "variable.h"
#include <numeric>
#include <vector>

TEST_CASE("Case0") {
  CHECK_EQ(var(1) + 2.5, 3.5);
  CHECK_EQ(var(3) - 1.0f, 2.0);
  CHECK_EQ(var(4.0) * 2, 8.0);
  CHECK_EQ(var(10) / 2, 5.0);
  CHECK_EQ(var(5.0) == 5, true);
  CHECK_EQ(var(6) != 6.1, true);
  CHECK_EQ(var(7.0) == 7.0f, true);
  CHECK_EQ(var(2) + var(2), 4.0);
  CHECK_EQ(var(2.0) + var(2), 4.0);
  CHECK_EQ(var("3"s) + var(1), 4.0);
  CHECK_EQ(var("1.5"s) * 2, 3.0);
  CHECK_EQ(var("10"s) * var("2"s), 20.0);
  CHECK_EQ(var(1.0) + var("2.5"s), 3.5);
  CHECK_EQ(var("6.0"s), var(6.0));
}

TEST_CASE("Case1") {
  var a = 10;
  var b = 10.;
  var c = 10.f;
  var d = true;
  var str1 = "1234"s;
  var str2 = str1 + "5678"s;

  CHECK_EQ(str2 , "12345678"s);
  CHECK_EQ(str1 + 100 , 1334);
}

TEST_CASE("Case2: Operations") {
  var a = 10.;
  var b = 5;
  var x = a + b;
  var y = x - b;
  var z = y * "2"s;

  CHECK_EQ(x.as<float>() , doctest::Approx(15.0));
  CHECK_EQ(y.as<float>() , doctest::Approx(10.0));

  CHECK_EQ(x.as<double>() , doctest::Approx(15.0));
  CHECK_EQ(y.as<double>() , doctest::Approx(10.0));

  CHECK_EQ(z , 20);

  try {
    var f = true + var(true);
  } catch (std::invalid_argument e) {
    CHECK_EQ(e.what(), "incompatible types"s);
  }

  try {
    var z = a / 0;
  } catch (std::domain_error e) {
    CHECK_EQ(e.what(), "not allow devide by 0"s);
  }

  try {
    var str = "abc"s;
    str = str * 3;
  } catch (std::invalid_argument e) {
    CHECK_EQ(e.what(), "convert failed"s);
  }
}

TEST_CASE("Case3: Mixed coercion and operations") {
  var L1 = 30;
  var L2 = 10.0 / "30"s;
  var L3 = true;
  var L4 = "10"s + "20"s;

  CHECK_EQ(L1 , 30);
  CHECK_EQ(L2 , 10. / 30.);
  CHECK_EQ(L3 , 1);
  CHECK_EQ(L4.as<std::string>() , "1020");

  try {
    var("abc"s) + 10;
  } catch (std::invalid_argument) {
    auto invalid_argument = true;
    CHECK_EQ(invalid_argument, true);
  } catch (std::exception) {
    auto exception = true;
    CHECK_EQ(exception, false);
  }

  try {
    var(1 - "hello"s) + 10;
  } catch (std::invalid_argument) {
    auto invalid_argument = true;
    CHECK_EQ(invalid_argument, true);
  } catch (std::exception) {
    auto exception = true;
    CHECK_EQ(exception, false);
  }
}

TEST_CASE("Case4: computation and nested loop") {
  std::vector<var> symbols = {"x", "y", "z"};
  std::vector<var> coeffs = {1, 2.5, -3};

  std::string expression = "";

  for (std::size_t i = 0; i < symbols.size(); ++i) {
    std::string term = coeffs[i].as<std::string>() + symbols[i].as<std::string>();
    if (i != symbols.size() - 1)
      term += " +";
    expression += term;
  }

  CHECK_EQ(expression, "1x +2.5y +-3z");

  // Evaluate expression with x=1, y=2, z=3
  std::vector<var> values = {1, 2, 3};
  var result = 0;
  for (std::size_t i = 0; i < values.size(); ++i) {
    result = result + coeffs[i] * values[i];
  }
  CHECK_EQ(result.as<double>(), doctest::Approx(1 * 1 + 2.5 * 2 + -3 * 3)); // = 1 + 5 - 9 = -3
}

TEST_CASE("Case5: STL integration and mixed operations") {
  std::vector<var> data = {1, 2.0f, 3.5, 4};
  double result = std::accumulate(data.begin(), data.end(), var(0.0), [](var acc, var x) {
                    return acc + x;
                  }).as<double>();

  CHECK_EQ(result, doctest::Approx(10.5));
}

#endif