# Examples

## Hello World Example

* Relies on [Daggerize an example application](https://docs.dagger.io/quickstart/daggerize/#initialize-a-dagger-module)

```bash
cd basic
dagger init --sdk=go --source=./hello-dagger
```

* Echo function

```bash
dagger -m basic/hello-dagger call container-echo --string-arg=hello
```

* Grep Function

```bash
dagger -m basic/hello-dagger call grep-dir --directory-arg=. --pattern=Examples
```

## Gnoland Git

Download the Gnoland source code from Git and run project's go tests

```bash
cd gnoland
dagger init --sdk=go --source=./git-test
```

* Git source code tests

```bash
dagger -m gnoland/git-test call git-code-test
```

## Gnokey

Play with `gnokey` image

```bash
cd gnokey
dagger init --sdk=go --source=./gnokey-tx
```

* Generate keys

```bash
dagger call generate-key-out --home-dir-key=. --password-string=helloworld
```

* Add a package (after generating keys)

```bash
dagger call make-txt --home-dir-key=. --password-string=helloworld
```
