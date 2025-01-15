# Dagger

A Pipeline as a Code devkit.

> No more `Push & Pray` stategies for CI/CD pipelines

* Dagger CLI
* Dagger Functions (input/output/run in containers)
* Dagger Modules (packaged Dagger function, call from Dagger CLI or from other Dagger Functions)
  * called from Github repos

> Dagger is a container orchestrator and runtime. Functions are executed as containers, and can call the Dagger Api to orchestrate more containers

## Calling Function from Module using CLI

    dagger -m github.com/shykes/daggerverse/hello@v0.1.2 call hello

**NOTE:**
> When using dagger call, all names (functions, arguments, fields, etc) are converted into a shell-friendly "kebab-case" style. This is why the function named Hello() is invoked as dagger call hello.

To list available arguments for a function, use dagger call FUNC --help. The Dagger CLI will dynamically inspect the function's arguments, and print the corresponding flag information.

    dagger -m github.com/shykes/daggerverse/hello@v0.1.2 call hello --help

## Builtin Types

Dagger defines powerful core types, such as Container.

### Container Type

Container represents the state of an OCI container.
This can be passed as argument: the CLI will dynamically pull the image, and pass the resulting container object as argument to the function.

### Directory Type

Represents a directory. It can be on a local filesystem, or a remote Git repository.

## Building a Binary

A function that builds binaries could take a Directory as argument (the source code) and return another Directory (containing binaries).

    dagger -m github.com/kpenfound/dagger-modules/golang@v0.1.5 call build --project https://github.com/dagger/dagger --args ./cmd/dagger

**NOTE:**
> The generated binary resides within the container itself and is not exposed by default

### Function chaining

Dagger Functions can return either basic types or objects. When calling a function that returns an object, the Dagger API lets you follow up by calling one of that object's functions.
Dagger's core types (Container, Directory, Service, Secret) are all objects. They each define various functions for interacting with their respective objects. When a function returns a core type the caller typically continues the chain by calling a function from that directory.

**NOTE:**
> You only need the Dagger CLI and the ability to run containers, so there is no need for Go, Git or any other local dependencies installed on the Dagger host.

Examples:

* Build a container and publish on registry

    dagger -m github.com/shykes/daggerverse/wolfi@v0.1.4 call container publish --address=ttl.sh/my-wolfi

* Start container as a Servive on a specific port

    dagger -m github.com/kpenfound/dagger-modules/nginx@v0.1.0 call container as-service up --ports=8080:80

* Live debug Container by executing it

    dagger -m github.com/shykes/daggerverse/wolfi@v0.1.4 call container --packages=cowsay terminal

## Building a container

It is possible to chain and return Container types in functions.

* build a container with additional args

    dagger -m github.com/shykes/daggerverse/wolfi@v0.1.2 call container --packages=cowsay

* build a container and chain execution of commands

    dagger -m github.com/shykes/daggerverse/wolfi@v0.1.2 call container --packages=cowsay with-exec --args cowsay,dagger stdout

* build and start interactive session

    dagger -m github.com/shykes/daggerverse/wolfi@v0.1.2 call container --packages=cowsay terminal

## Module Help

### Functions exposed by module

    dagger -m github.com/kpenfound/dagger-modules/golang@v0.1.5 functions

or

    dagger -m github.com/kpenfound/dagger-modules/golang@v0.1.5 call --help

### Help on specific function

    dagger -m github.com/kpenfound/dagger-modules/golang@v0.1.5 call build --help

## Daggerize project

* Initialize

    dagger init

This generates the `dagger.json` file that can be commited in git.

* Add Dagger Go SDK

    dagger develop --sdk=go

* Add dependecies to the project

    dagger install github.com/shykes/daggerverse/hello@v0.1.2

This will automatically update the `dagger.json` file.

* Check existing modules in the `daggerverse`

> The Daggerverse is a free service run by Dagger, which indexes all publicly available Dagger modules, and lets you easily search and consume them.

* Regenerate module after finish develop

    dagger develop --sdk=go --mod=<path_to_git_module>

## Ref

* [GitLab CI | Dagger](https://docs.dagger.io/integrations/734201/gitlab)
* [CLI Reference | Dagger](https://docs.dagger.io/reference/979596/cli/)
* [Developing with Go | Dagger](https://docs.dagger.io/manuals/developer/go/)
* [Dagger Does AI - YouTube](https://www.youtube.com/watch?v=Kgqk5N2hlVU&t=1052s)
* [Cookbook | Dagger](https://docs.dagger.io/cookbook)
