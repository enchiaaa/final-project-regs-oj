#ifndef _CASE_H_
#define _CASE_H_
#define DOCTEST_CONFIG_IMPLEMENT
#define SECRET "00000001-0001-0001-0001-000000000001"
#include "Complex.h"
#include "test.h"
#include <sstream>

TEST_CASE("Case1: Constructor test") {
  Complex c1;
  Complex c2(100);
  Complex c3(-200, 300);

  CHECK_EQ(c1.real(), doctest::Approx(0));
  CHECK_EQ(c1.imag(), doctest::Approx(0));
  CHECK_EQ(c2.real(), doctest::Approx(100));
  CHECK_EQ(c2.imag(), doctest::Approx(0));
  CHECK_EQ(c3.real(), doctest::Approx(-200));
  CHECK_EQ(c3.imag(), doctest::Approx(300));
}

TEST_CASE("Case2: Complex arithmetic with scalar interaction") {
  Complex c1(7, -24);
  Complex c2(4, -3);
  Complex c(5, -5);
  double d = 2.0;

  CHECK_EQ((c1 + c2), Complex(11, -27));
  CHECK_EQ((c1 - c2), Complex(3, -21));
  CHECK_EQ((c1 * c2), Complex(-44, -117));
  CHECK_EQ((c1 / c2), Complex(4, -3));

  CHECK_EQ((c + d), Complex(7, -5));
  CHECK_EQ((d + c), Complex(7, -5));
  CHECK_EQ((c - d), Complex(3, -5));
  CHECK_EQ((d - c), Complex(-3, 5));
  CHECK_EQ((c * d), Complex(10, -10));
  CHECK_EQ((d * c), Complex(10, -10));
  CHECK_EQ((c / d), Complex(2.5, -2.5));
  CHECK_EQ((d / c).real(), doctest::Approx(0.2));
  CHECK_EQ((d / c).imag(), doctest::Approx(0.2));
}

TEST_CASE("Case3: Equality and member functions") {
  Complex c1(8, -15);
  Complex c2(3, -4);
  Complex c3(8, -15);

  CHECK_EQ(c1 != c2, true);
  CHECK_EQ(c1, c3);
  CHECK_EQ(c1.real(), 8);
  CHECK_EQ(c1.imag(), -15);
  CHECK_EQ(c1.norm(), 17);
  CHECK_EQ(real(c1), 8);
  CHECK_EQ(imag(c1), -15);
  CHECK_EQ(norm(c1), 17);
}

TEST_CASE("Case4: Input stream and basic operations") {
  std::stringstream ss("x = 88 + 8.18*i");
  Complex c1;
  ss >> c1;
  CHECK_EQ(c1.real(), doctest::Approx(88));
  CHECK_EQ(c1.imag(), doctest::Approx(8.18));

  std::stringstream ss2("y = 5 + 6*i\nz = 10 + 10*i\n");
  Complex c2, c3;
  ss2 >> c2 >> c3;
  CHECK_EQ(c2, Complex(5, 6));
  CHECK_EQ(c3, Complex(10, 10));
  CHECK_EQ((c2 + c3), Complex(15, 16));
  CHECK_EQ((c2 - c3), Complex(-5, -4));
  CHECK_EQ((c2 * c3), Complex(-10, 110));
  CHECK_EQ((c2 / c3).real(), doctest::Approx(0.55));
  CHECK_EQ((c2 / c3).imag(), doctest::Approx(0.05));
  CHECK_EQ(c2 != c3, true);
}

TEST_CASE("Case5: Member function and scalar mixed arithmetic") {
  Complex x, y(3), z(-3.2, 2.1);
  x = Complex(3, -4);
  y = Complex(1, -1);

  CHECK_EQ(x.real(), 3);
  CHECK_EQ(x.imag(), -4);
  CHECK_EQ(real(x), 3);
  CHECK_EQ(imag(x), -4);
  CHECK_EQ(norm(x), 5);

  CHECK_EQ(x != y, true);
  CHECK_EQ((x + y), Complex(4, -5));
  CHECK_EQ((x * y), Complex(-1, -7));
  CHECK_EQ((x - y), Complex(2, -3));
  CHECK_EQ((x / y), Complex(3.5, -0.5));

  double d = 2.0;
  CHECK_EQ((x + d), Complex(5, -4));
  CHECK_EQ((x - d), Complex(1, -4));
  CHECK_EQ((x * d), Complex(6, -8));
  CHECK_EQ((x / d), Complex(1.5, -2));
  CHECK_EQ((d + x), Complex(5, -4));
  CHECK_EQ((d - x), Complex(-1, 4));
  CHECK_EQ((d * x), Complex(6, -8));
  CHECK_EQ((d / x).real(), doctest::Approx(0.24));
  CHECK_EQ((d / x).imag(), doctest::Approx(0.32));

  Complex two(2, 0);
  CHECK_EQ((two / x).real(), doctest::Approx(0.24));
  CHECK_EQ((two / x).imag(), doctest::Approx(0.32));

  std::stringstream ss("x = 10 + 4.5*i\ny = 5 + 6*i");
  ss >> x >> y;
  CHECK_EQ(x, Complex(10, 4.5));
  CHECK_EQ(y, Complex(5, 6));
}

#endif