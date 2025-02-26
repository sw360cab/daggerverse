# Gno Dagger

A Playground for Dagger CI in Gno

* `basic`: basic examples
* `gnoland`: clone Gnoland git repo adn run go tests
* `gnokey`: play with Gnokey basic operations
* `supernova`: full fledged Supernova call example

## TODO

* simulate PL
* call tx

```bash
gnokey maketx call -insecure-password-stdin \
-gas-fee 1000000ugnot -gas-wanted 3000000 \
-broadcast -chainid dev \
-pkgpath "gno.land/r/demo/tests_copy" -func "InitTestNodes" \
test1`
```
