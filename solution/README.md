# OOP 期末解答思路

## Complex

作業題，本題比較多人做錯的是浮點數精度的問題：

考慮到 a 以及 std::pow(std::sqrt(a), 2) 未必是相同的

您應該謹慎考慮運算的順序，另一種做法是使用 Approxy 的方式進行 `operator==` 實作

```cpp
if( doubleA - doubleB < 1e-6 ) {
  // ...
}
```

若兩數小於某個誤差，視作相等

> 作業題，同學是不應該出錯的

## Neural Network

閱讀完題目後，預期你應該實作 `Layer.h` 跟 `Optimizer.h` 的 virtual function

- OPTIMIZER 包含 `SGD` 與 `ADAM`
- LEARNING_RATE 分成 4 組，OPTIMIZER 各兩個
- EPOCHS 可能為 100, 200, 1000

所有的組合一共有 `2 * 2 * 3` = 12 種

一共要訓練五種DATASET，如果你運氣極差，每個都訓練12次才得到結果
總共要進行 60 次﹐使用窮舉法就可以完成作答。

> 本題重點是閱讀文件與分析需求，沒有其他難度
> 同學看到繼承應該要知道去實作 virtual function.

## School

無任何需要注意的事項，注意文件說明即可。

## URL Parser

本題由出題助教，以及助教長提供兩種不同的解答。

助教長的解法使用上課時提及的 `sstream` 分割做法。

仔細觀察每個區段的分隔符號，以及允許的字元集，設定對應的 delimter 即可。

> 考前若有認真聽助教長的複習，好好理解 scanf 與 streambuf 的運作，應該是可以輕鬆解決此題的。 若同學使用 `regex` 進行運算，此算法也行，但會提高一些計算複雜度。

## Slice

本題主要的難點包含

1. Functor 該如何推理
2. `Slice<T>::map`

iterator 的實作，參考 `hint.h` 跟 `iterator.h` 的部分實作，把 `--` 系列的實作出來就好。
其實就是把 `++` 系列的算法改成 `-` 的就好。

在文件中特別說明：

> The functors passed to `map`, `filter`, and `every` must have clearly defined parameter and return types, so that their usage can be type-checked safely.

這段話的意義是所有的 Functor 都可以合理地推理出 ReturnType，因此可以使用 `U (*functor)(T arg)` 這種函式指標匹配出型別。

另一種做法是使用 `decltype` 與 `auto`

透過 `decltype` 調用 `Functor` 取得 `U` 並使用 `auto` 讓編譯器推理回傳型別，這樣就不用於模板參數定義 `typename U`

```cpp
template <typename Functor>
auto map(Functor functor) const {
  using U = decltype(functor(T{}))
  Storage<U> result = {};
  for (auto& item : this->_data) {
    result.push_back(callback(item));
  }
  return Slice<U>(result);
}
```

以下描述：

> All template parameters T used in this problem are required to be default-initializable. That is, the expression T{} must be valid and produce a default value  of type T.

其意義是 `T` 包含了預設的建構，就是為了讓同學可以在 `map` 直接使用 `T{}` 初始化一個預設參數。

而 `filter` 與 `every` 可以跟 `map` 的推理相同，但更加簡單，因為 `U` 必然為 `bool`

一個潛在的困難是，若您嘗試在 `map` 直接操作 `Slice<U>::_data`

`Slice<U>` 與 `Slice<T>` 並不是相同型別，您嘗試透過`Slice<T>`操作`Slice<U>`的私有型別。因此助教的解答中，包含了

```cpp
  template <typename Iterable>
  Slice(const Iterable& container)
      : _data(std::begin(container), std::end(container)) {}
```

建構式，並在內部操作 `Storage<U>`進行複製。

您也可以透過 `friend`，允許所有基於 `Slice` 的模板互相操作

```cpp
template<typename T>
class Klass {
  template<typename U>
  friend class Klass;
}
```

clang 預設把基於 Klass 的模板都額外插入 friend 裝飾。但我覺得這不是一個可以依賴的行為，若考慮到模板的偏特化或是全特化，會有潛在性的問題

> 本題若您對 lambda , std::function , 函式指標夠熟悉，其實很接近 Easy 的等級
> 但考慮到同學對Template 不熟，所以定義為 Hard

## Glang

期末壓軸題，您要具備以下知識：

1. `Value<T>` 繼承 `IValue`
2. `Value<T>*` 可以轉換成 `IValue*`
3. 透過 `dispatch` 這個 `IValue*`，以及 virtual table，調用 `Value<T>` 的成員函式。

目前的題目應該已經預設了 `operator==` 的實作，其行為是嘗試優先轉換成Double 進行比較，若失敗；則進行 String 的轉換比較。

str + 100 與 1334 的比較，應該會進行 `var(1334.0) == 1334`，根據 C++ 的 implicit deduction，會嘗試把等號右側修改成 `var(1334)`，並嘗試進行 `var(1334.0) == var(1334)`

因此，您也需要理解建構式可能的行為。

對於轉型，應該始終使用 `sstream` 處理，考慮到 val 的 typename T 可能是 `bool` 或是 `std::string` 而 `std::to_string` 是不允許 `to_string(bool)` 或是 `to_string(std::string)` 的運算。
在文件上建議同學使用 stringstream 進行轉換，因為 to_string 根據實作，是有不低的機率擲出 Compile Failed。

在 `value.h` 中，定義了很多工具函式。若您對於這些都不了解，其實可以使用 `void*` 與 `static_cast` 進行大量的操作來處理測資。
否則，可以使用一些模板與 Marco 技術快速完成。

> 唯有本題可以真正視作 Hard，題型綜合了 Virtual Table , Template , Error Handler 等知識點
> 並進行合理的組合。本題是助教長的精巧之作。

## Note

若您嘗試在 005, 006 分離 .cpp 與 .h ， `template` 得要調用才能進行推導
這可能會導致您直接無法作答。這在分離式編譯與template的講義都有說明到
