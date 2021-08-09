# gggit

This is a minimal Git reimplementation written in Go.

It does not aim for interoperability with Git - only a very limited subset of
features is implemented and large simplifications have been made.

The aims of this project are rather selfish, that is:

 - to reimplement interesting project from the ground up
 - to improve my Go skills
 - to gain a better understanding of Git plumbing

## quick start

You could really easily try the thing out!

```
go install github.com/antoniszczepanik/gggit@v0.1.0

mkdir project1 && cd project1
gggit init
```
(assuming GOPATH is in your PATH ;))
