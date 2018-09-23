# Building from the source

You will need the latest version of the Go programming language.

If you already have the go1.11 installed, skip to [Downloading The Source](#downloading-the-source)

If there is a [newer version of go](https://golang.org/dl/), use that instead.

### Installing Go on Linux

```
wget https://dl.google.com/go/go1.11.linux-amd64.tar.gz
sudo tar xvaf go1.11.linux-amd64.tar.gz -C /usr/local/
```
Add go to path (consider adding this to $HOME/.bashrc)

```
export PATH=$PATH:/usr/local/go/bin
```

Check go is installed

```
which go
go version
```

### Installing Go on Mac OS X

```
wget https://dl.google.com/go/go1.1.darwin-amd64.pkg
sudo tar xvaf go1.11.darwin-amd64.pkg -C /usr/local/
```
Add go to path (consider adding this to $HOME/.bashrc)

```
export PATH=$PATH:/usr/local/go/bin
```

Check go is installed

```
which go
go version
```



## Installing Aquachain RPC node on openbsd

Install git, go, build aquachain and start RPC server on 127.0.0.1:8543

```
su
pkg_add git
pkg_add go
exit
go get -d -v gitlab.com/aquachain/aquachain
cd go/src/gitlab.com/aquachain/aquachain/
make
./build/bin/aquachain -rpc
```

## Downloading The Source

```
go get -v -d gitlab.com/aquachain/aquachain
cd $(go env GOPATH)/src/gitlab.com/aquachain/aquachain
```

If you are contributing, you will want to 'fork' the main repo on github, and add your fork like so, changing `your-name` and `patch-1` to whatever you need:

```
git remote add fork git@github.com:yourname/aquachain.git
git checkout -b patch-1
```

When done making commits, use `git push fork patch-1` and either open a pull request or ask to merge.

During development on your branch there may be many commits on the `master` branch. You can re-synchronize by using `git pull -r origin master` or `git rebase -i` to avoid needing _Merge commits_.

## Compiling

In the base directory of the repository, you can run a variety of 'make' targets.

When finished compiling, they are in the `./build/bin` directory.

Aquachain command

```
make
```

Aquaminer command

```
make aquaminer
```

Available make targets on linux

```
all              aquachain-nocgo  cross            install          static
all-musl         aquaminer        devtools         lint             test
all-static       aquastrat        docker-run       musl             test-musl
aquachain        clean            generate         race             usb
   
```

## Cross Compilation

This is how releases are made, with Docker and xgo. Here we build only amd64 targets, Windows, Linux, OSX (darwin).


```
xgo -out aquachain-$VERS --remote https://github.com/aquanetwork/aquachain  --branch master --pkg cmd/aquachain --targets='*/amd64' github.com/aquanetwork/aquachain`

Then zip:

```
for i in aquachain-$VERS-*; do zip $i.zip $i; done
```

Then hash:

```
sha256sum * | grep -v zip
```

