#pragma once
namespace lib {
template <typename T> class Box {
public:
  Box() = default;
  explicit Box(const T &v) : v_(v) {}
  T get() const { return v_; }
  void set(const T &v) { v_ = v; }
  bool eq(const T &o) const { return v_ == o; }
  static Box<T> make(T v) { return Box<T>(v); }
private:
  T v_{};
};
struct NoEq {
  int x = 0;
};
class Holder {
public:
  Holder() : b_(7) {}
  Box<int> box() const { return b_; }
  void setBox(const Box<int> &b) { b_ = b; }
private:
  Box<int> b_;
};
} // namespace lib
