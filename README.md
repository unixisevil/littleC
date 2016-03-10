# littleC
golang implement   herbert schildt's  little c  interpreter

it's just a  recursive descent parser exercise

little c 是一个功能相当有限的c语言子集:
-  支持局部变量的参数化函数，但是局部变量声明只能出现在函数的开头位置上
-  递归函数
-  if 语句， if body 需要用{}包围
-  do, do-while, for 循环， body需要用{}包围
-  整型，字符型局部及全局变量
-  整型，字符型函数参数
-  字符串常量
-  return 语句
-  五个内置函数
-  运算符 +, -, *, /, %, ^, < , > , <=, >=, ==, !=, 一元+， 一元-
-  返回整型的函数
-  c 的/*  .....   */ 注释

### compile 
```go build```

### run
```./littleC   demo/demo4.c ```
