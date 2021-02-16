# monkey-go

## 语法

### features
```
let name = "monkey";
let age = 1;
let arr = ["Go","Lua","JavaScript"];
let person = {
    "name": "tom",
    "age" : 22,
};

let printNameAge = fn(book) {
    let name = book["name];
    let age = book["age"];
    puts(name + " - " + age);
};
printNameAge(person)

let numbers = [1,2,3,4];
map(numbers,fn(x) {
    return x * 2;
});
```

### types
- integers
- booleans
- strings
- arrays
- hashes
- prefix,infix and index operators
- conditionals
- global and local bindings
- first-class function
- return statements
- closures
