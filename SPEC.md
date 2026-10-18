# Tokype Specification

**Version**: 0.0.3

## Comments

```tokype
//comment
```

## Creating variables

```tokype
a = 10                           // Int
b = "Hello"                      // String
c = true                         // Boolean
d = [1, 2, 3]                    // List
e = {"Hello" :: "world", 1 :: 2} // Map
f = null                         // Null
g = 2.5                          // Float
h = .5                           // Float
i = -2                           // Int
```

## Built-in functions

```tokype
print(10) // 10

input("You can enter text: ")

a = [10, 20, 30]
b = len(a)

print(b) // 3

print(negate(5)) // -5

push(a, 40)
print(a) // [10, 20, 30, 40]

print(pop(a))          // [10, 20, 30]
print(first(a))        // 10
print(rest(a))         // [20, 30, 40]
print(contains(a, 10)) // true

c = get_time()
print(c) // time since the interpreter started running

a = reserve([], 1000000) // memory allocation for the list
big_list = repeat(5, 10) // creates a list (for small numbers of elements) or a repeat (for large numbers)

print("List: " + big_list.[1]) // 5
```

## Functions

```tokype
funct main():
   result = 10
   print(result)
end

main() // 10
```

## Conditions

```tokype
if 5 > 7:
   print("5 > 7")
elif 5 > 6:
   print("5 > 6")
else:
   print("5 < 7 and 5 < 6")
end
```

## Loops

```tokype
sum = 0
for i = 1, 10:
   sum = sum + i // 55
end

x = 10

while x > 0:
   print(x) // 10, 9, 8, 7, 6, 5, 4, 3, 2, 1
   x = x - 1
end

for i = 1, 10, 2:
   print(i) // 1, 3, 5, 7, 9
end
```

## Lists and dictionaries

```tokype
a = [10, 20, 30]
print(a.[1]) // 20

b = {10 :: "10", "20" :: "20"}
print(b.["20"]) // 20
```

## Operators

```tokype
a = 10 + 5   // 15
b = 10 - 5   // 5
c = 10 * 5   // 50
d = 10 / 5   // 2
e = 10 % 3   // 1

print(5 == 5)  // equality (true)
print(5 ?= 5)  // inequality (false)
print(5 < 10)  // less than (true)
print(5 > 10)  // greater than (false)
print(5 <= 5)  // less than or equal to (true)
print(5 >= 5)  // greater than or equal to (true)

print("Hello, " + "world!") // Hello, world!
```

## Null

```tokype
x = null

if x == null:
    print("x is null")
end

print(x == 5) // false
print(x ?= 5) // true
```

## Float

```tokype
a = .5
b = 90.12
c = 1.0
d = 5 / 2

print(a) // 0.5
print(d) // 2.5
```

## Unary minus

```tokype
a = -5
b = -0.1
c = -a

print(c) // 5
```

## Include

```tokype
//math.tp

funct add(x, y):
   return x + y
end
```

```tokype
//main.tp

include "math.tp"

funct main():
   print(add(1, 5)) // 6
end

main()
```

## Include with namespace

```tokype
// math.tp

funct add(x, y):
    return x + y
end

pi = 3.14159265359
```

```tokype
// main.tp

include "math.tp" as math

funct main():
    print(math.pi)        // 3.14159265359
    print(math.add(5, 3)) // 8
end

main()
```

## String and list operations

```tokype
print("abc" * 3) // abcabcabc

print([1, 2] * 3) // [1, 2, 1, 2, 1, 2]

print("hello" - "l")  // heo
print("hello" - "lo") // he

print([1, 2, 3] - 2)          // [1, 3]
print([1, 2, 2, 3] - [2, 3])  // [1]
```

## Documentation

- [**README.md**](README.md) — overview and quick start
- [**TitanJerboa.md**](TitanJerboa.md) — optimizer description