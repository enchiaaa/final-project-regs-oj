#ifndef _CASE_H_
#define _CASE_H_
#define DOCTEST_CONFIG_IMPLEMENT
#include "operator.h"
#include "test.h"
#include "variable.h"

TEST_CASE("Case: Declare") {
  var a = 10;
  var b = 10.;
  var c = 10.f;
  var d = true;
  var str1 = "1234"s;
  var str2 = str1 + "5678"s;

  CHECK_EQ(str2, "12345678"s);
  CHECK_EQ(str1 + 100, 1334);
}

TEST_CASE("Case: Member Functions") {
  var a = 10.;
  var b = 5;
  var x = a + b;
  var y = x - b;
  var z = y * "2"s;

  CHECK_EQ(x.as<float>(), doctest::Approx(15.0));
  CHECK_EQ(y.as<float>(), doctest::Approx(10.0));

  CHECK_EQ(x.as<double>(), doctest::Approx(15.0));
  CHECK_EQ(y.as<double>(), doctest::Approx(10.0));

  CHECK_EQ(z, 20);
  CHECK_THROWS_WITH(var f = true + var(true), "incompatible types");
  CHECK_THROWS_WITH(var z = a / 0, "not allow devide by 0");
  CHECK_THROWS_WITH("abc"s * 3, "convert failed");
}

TEST_CASE("Case: computation and nested loop") {
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

#endif