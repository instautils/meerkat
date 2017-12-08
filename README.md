# Meerkat - Instagram watcher !
Watch your following activities on Instagram.

<img align="center" src="https://github.com/ahmdrz/meerkat/blob/master/resources/meerkat.jpg" alt="meerkat github">

## How to use ?

First of all , download meerkat from released binaries or using `go get` command.

```
    go get -u github.com/ahmdrz/meerkat
```

Now we have to make `meerkat.yaml` for meerkat configurations. with

```
    meerkat init
```

we can create default `meerkat` configuration file.

It's time to change default variable such as `username` and `password` of your Instagram account.

1. Open `meerkat.yaml` using `gedit`, `nano`, `vim` or any other editors.
2. Replace `username` and `password` in file.
3. Add your targets in `targetusers` array list.
4. Save, and run `meerkat`.
5. Enjoy !

You can set optional flags in `meerkat` command.

```
Usage of meerkat:
  -config string
    	Configuration file (YAML format)
  -output string
    	Log output file.
```

### TODOs 

1. Add more options for output of logs.
2. Add two sub commands for `encrypt` and `decrypt` yaml file.


Built with :heart: by Ahmadreza Zibaei
