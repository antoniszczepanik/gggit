# gggit

This is a minimal Git reimplementation written in Go.

It does not aim for interoperability with Git. Only a limited subset of
features is implemented and some significant simplifications have been made.

The aims of this project are rather selfish, that is:

 - to reimplement interesting project from the ground up
 - to improve my Go skills
 - to gain a better understanding of Git plumbing

 Currently supported commands are:

```
gggit hash-object
gggit cat-file
gggit status
gggit commit
```

## quick start

```
go install github.com/antoniszczepanik/gggit@v0.1.0

mkdir project1 && cd project1
gggit init

```

### todo

- basic checkout command support (for now let's focus on checking out a branch)
    - add git branch support, which will only create a branch from a current one
    - add checkout to checkout given branch

- index (!!!), add, reset
- `.gitignore` support
- config file support
- commits serializitation/deserialization is not complete, cannot specify own message :(
- not all permission bits are set
- clean-up logging: use idiomatic go logging solution, levels etc
