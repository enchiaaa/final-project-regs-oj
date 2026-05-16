#ifndef _CASE_H_
#define _CASE_H_
#define DOCTEST_CONFIG_IMPLEMENT
#include "slice.h"
#include "test.h"
#include <algorithm>
#include <deque>
#include <list>
#include <type_traits>
#include <vector>

inline bool isEven(int x) {
  return x % 2 == 0;
}

struct Vec2 {
  int x;
  int y;
};

struct Vec3 {
  int x;
  int y;
  int z;

  inline bool operator==(const Vec3& other) const {
    return x == other.x && y == other.y && z == other.z;
  }
};

inline std::ostream& operator<<(std::ostream& os, const Vec3& v) {
  return os << "(" << v.x << "," << v.y << "," << v.z << ")";
}

TEST_CASE("Case1: Check type aliases and forbid undefined types") {
  using vector_iter = std::vector<int>::iterator;
  using list_iter = std::list<int>::iterator;
  using deque_iter = std::deque<int>::iterator;
  using Slice_iter = decltype(std::declval<Slice<int>>().begin());

  bool StorageIsUnknownType = std::is_same<Slice<int>::Storage_t, UnknownType>::value;
  CHECK_EQ(StorageIsUnknownType, false);

  auto PredicateIsUnknownType = std::is_same<Slice<int>::Predicate, UnknownType>::value;
  CHECK_EQ(PredicateIsUnknownType, false);

  auto isVectorIter = std::is_same<vector_iter, Slice_iter>::value;
  CHECK_EQ(isVectorIter, false);

  auto isDequeIter = std::is_same<deque_iter, Slice_iter>::value;
  CHECK_EQ(isDequeIter, false);

  auto isListIter = std::is_same<list_iter, Slice_iter>::value;
  CHECK_EQ(isListIter, false);
}

TEST_CASE("Case2: Basic Use") {
  Slice<int> original({1, 2, 3, 4, 5, 6});

  // map
  auto squared = original.map(+[](int x) { // [1, 4, 9, 16, 25, 36]
    return x * x;
  });
  CHECK_EQ(squared.size(), 6);
  CHECK_EQ(squared[0], 1);
  CHECK_EQ(squared[5], 36);

  // filter
  std::function<bool(int)> ge10 = [](int x) {
    return x > 10;
  };

  auto filtered = squared.filter(ge10); // [16, 25, 36]
  CHECK_EQ(filtered.size(), 3);
  CHECK_EQ(filtered[0], 16);
  CHECK_EQ(filtered[1], 25);
  CHECK_EQ(filtered[2], 36);

  // every: check if all remaining are even
  bool allIsEven = filtered.every(isEven);
  CHECK_EQ(allIsEven, false); // 25 is odd

  CHECK_EQ(
      original
          .map(+[](int x) {
            return x * x;
          })
          .filter(+[](int x) {
            return x > 10;
          })
          .every(+[](int x) {
            return x > 5;
          }),
      true);
}

TEST_CASE("Case3: suitable Based-range For-loop") {
  Slice<int> s({1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15});
  int sum = 0, count = 1;
  for (auto v : s) {
    CHECK_EQ(count, v);
    sum += v;
    ++count;
  }
  CHECK_EQ(sum, 120);
}

TEST_CASE("Case4: STL algorithm compatibility") {
  typedef struct Point {
    int x;
    int y;
  } Point;

  Slice<Point> points({
      {0,  0},
      {1,  2},
      {2,  4},
      {3,  6},
      {4,  8},
      {5, 10},
      {6, 12},
      {7, 14},
      {8, 16},
      {9, 18}
  });

  // 1. std::find_if: find point with x == 3
  auto it = std::find_if(points.begin(), points.end(), [](const Point& p) {
    return p.x == 3;
  });
  CHECK_EQ(it != points.end(), true);
  CHECK_EQ(it->y, 6);

  // 2. std::all_of: all y are even?
  bool all_even = std::all_of(points.begin(), points.end(), [](const Point& p) {
    return p.y % 2 == 0;
  });
  CHECK_EQ(all_even, true);

  // 3. std::any_of: is there any x > 8?
  bool any_large = std::any_of(points.begin(), points.end(), [](const Point& p) {
    return p.x > 8;
  });
  CHECK_EQ(any_large, true);

  // 4. std::count_if: how many x are odd?
  int odd_count = std::count_if(points.begin(), points.end(), [](const Point& p) {
    return p.x % 2 == 1;
  });
  CHECK_EQ(odd_count, 5);

  // 5. std::for_each: accumulate all y into a sum
  int total_y = 0;
  std::for_each(points.begin(), points.end(), [&](const Point& p) {
    total_y += p.y;
  });
  CHECK_EQ(total_y, 90);

  // 6. std::transform: create a new vector<int> containing only x
  std::vector<int> xs;
  std::transform(points.begin(), points.end(), std::back_inserter(xs), [](const Point& p) {
    return p.x;
  });
  CHECK_EQ(xs.size(), 10);
  CHECK_EQ(xs[0], 0);
  CHECK_EQ(xs[9], 9);
}

TEST_CASE("Case5: Complex testing") {
  Slice<Vec2> v2({
      {1, 2},
      {2, 2},
      {3, 3},
      {4, 1},
      {0, 6}
  });

  // map: Vec2 → Vec3, where z = x + y
  auto v3 = v2.map(+[](Vec2 v) {
    return Vec3{v.x, v.y, v.x + v.y};
  });

  CHECK_EQ(v3.size(), 5);
  CHECK_EQ(v3[0], Vec3{1, 2, 3});
  CHECK_EQ(v3[4], Vec3{0, 6, 6});

  // filter: only keep Vec3 where z > 5
  auto filtered = v3.filter(+[](Vec3 v) {
    return v.z > 5;
  });

  CHECK_EQ(filtered.size(), 2);
  CHECK_EQ(filtered[0].z, 6);
  CHECK_EQ(filtered[1].z, 6);

  // every: check that for all filtered elements, x + y == z
  bool valid = filtered.every(+[](Vec3 v) {
    return (v.x + v.y) == v.z;
  });
  CHECK_EQ(valid, true);
}

#endif
