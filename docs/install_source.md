# Installation from source

1. Verify that you have Go 1.26+ installed

   ```sh
   $ go version
   ```

   If `go` is not installed, follow instructions on [the Go website](https://golang.org/doc/install).

2. Clone this repository

   ```sh
   $ git clone https://github.com/cli/cli.git gh-cli
   $ cd gh-cli
   ```

3. Build and install

   **Unix-like systems**

   ```sh
   # installs to '/usr/local' by default; sudo may be required, or sudo -E for configured go environments
   $ make install

   # or, install to a different location
   $ make install prefix=/path/to/gh
   ```

   **Windows**

   ```pwsh
   # build the `bin\gh.exe` binary
   > go run script\build.go
   ```

   There is no install step available on Windows.

4. Run `gh version` to check if it worked.

   **Windows**

   Run `bin\gh version` to check if it worked.

## Building and installing ghx from source

This fork can also build a side-by-side `ghx` binary. The regular source install instructions above build upstream-style `gh`; use these targets when you want the fork binary without replacing `gh`.

**Unix-like systems**

```sh
$ make bin/ghx
$ make install-ghx prefix=$HOME/.local
$ ghx version
```

`install-ghx` installs only fork-owned paths:

- `${prefix}/bin/ghx`
- `${prefix}/share/man/man1/ghx*.1`
- `${prefix}/share/bash-completion/completions/ghx`
- `${prefix}/share/fish/vendor_completions.d/ghx.fish`
- `${prefix}/share/zsh/site-functions/_ghx`
- `${prefix}/share/zsh/vendor-completions/_ghx`

To remove those files without touching an upstream `gh` install:

```sh
$ make uninstall-ghx prefix=$HOME/.local
```

**Windows**

```pwsh
> go run script\build.go bin\ghx.exe
> bin\ghx version
```

The source install target builds ghx shell completions and ghx manpages locally. Package manager metadata and update metadata are still roadmap work.

## Cross-compiling binaries for different platforms

You can use any platform with Go installed to build a binary that is intended for another platform
or CPU architecture. This is achieved by setting environment variables such as GOOS and GOARCH.

For example, to compile the `gh` binary for the 32-bit Raspberry Pi OS:

```sh
# on a Unix-like system:
$ GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 make clean bin/gh
```

```pwsh
# on Windows, pass environment variables as arguments to the build script:
> go run script\build.go clean bin\gh GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0
```

Run `go tool dist list` to list all supported values of GOOS/GOARCH.

Tip: to reduce the size of the resulting binary, you can use `GO_LDFLAGS="-s -w"`. This omits
symbol tables used for debugging. See the list of [supported linker flags](https://golang.org/cmd/link/).
