# Pistol

## Introduction

Pistol is a file previewer for command line file managers such as
[Ranger](https://ranger.github.io/) and [Lf](https://github.com/gokcehan/lf)
intended to replace the file previewer
[`scope.sh`](https://github.com/ranger/ranger/blob/v1.9.2/ranger/data/scope.sh)
commonly used with them.

`scope.sh` is written in Bash and it uses several `case` switches to handle
every [MIME type](https://en.wikipedia.org/wiki/Media_type) or file extension.
Bash has a slow startup time and the `case` switches make it hard to configure
/ maintain. Additionally, `scope.sh` invokes the external program `file` to
determine the MIME type which makes it even slower.

Pistol is written in Go (with 0 dependencies) and it's MIME type detection is
internal. Moreover, it features native preview support for most types of
archive files and for text files along with syntax highlighting while
`scope.sh` relies on external programs to do these basic tasks.

The following table lists Pistol's native previewing support:

File/MIME Type  | Notes on implementation
---------- | -----------------------
`text/*`   | Prints text files with syntax highlighting thanks to [`chroma`](https://github.com/alecthomas/chroma).
Archives   | Prints the contents of archive files using [`archiver`](https://github.com/mholt/archiver).

In case Pistol encounters a MIME type which it doesn't know how to handle
natively and no configuration was defined for it, a general description of the
file type will be printed. E.g, the preview for an executable will be similar
to this:

```
ELF 64-bit LSB executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, BuildID[sha1]=a34861a1ae5358dc1079bc239df9dfe4830a8403, for GNU/Linux 3.2.0, not stripped
```

This feature is available out of the box as well.

### A Note on MIME type Detection

There are several Go libraries that provide MIME type detection. Here are the
top search results I got using a common web search engine:

- https://github.com/gabriel-vasile/mimetype
- https://github.com/h2non/filetype
- https://github.com/rakyll/magicmime

Pistol uses the last one which leverages the well known C library
[libmagic(3)](http://linux.die.net/man/3/libmagic). I made this choice after
experimenting with the other candidates and due to their lack of an extensive
database such as [libmagic(3)](http://linux.die.net/man/3/libmagic) has,
I chose [magicmime](https://github.com/rakyll/magicmime).

Note that this choice also features compatibility with the standard command
`file` which is available by default on most GNU/Linux distributions.
<sup id="a1">[1](#f1)</sup>.

## Install

### Prerequisites

Since Pistol depends on  [magicmime](https://github.com/rakyll/magicmime),
you'll need a `libmagic` package installed. Please refer to [this section in
magicmime's
README](https://github.com/rakyll/magicmime/tree/v0.1.0#prerequisites) for the
appropriate commands for every OS.

Assuming `libmagic` is installed and you have [setup a Go
environment](https://golang.org/doc/install), Use the following command to
install Pistol to `$GOPATH/.bin/pistol`:

```sh
go get -u github.com/doronbehar/pistol/cmd/pistol
```

## Usage

```
$ pistol --help
Usage: pistol OPTIONS <file>

OPTIONS

-c, --config <config>  configuration file to use (defaults to /home/doron/.config/pistol.conf)
-h, --help             print help and exit
-v, --verbosity        increase verbosity

ARGUMENTS

file                   the file to preview
```

### Integrations

#### Ranger / Lf

Pistol can be used as a file previewer in [Ranger](https://ranger.github.io/)
and [Lf](https://github.com/gokcehan/lf). For Ranger, set your `preview_script` in your
`rc.conf` as follows:

```
set preview_script ~/.go/bin/pistol
```

The same goes for Lf, but in `lfrc`:

```
set previewer ~/.go/bin/pistol
```

You need to put the full path of where `pistol` is installed. Use the command
`which pistol` to be sure of that.

#### fzf

If you use [fzf](https://github.com/junegunn/fzf) to search for files, you can
tell it to use `pistol` as the previewer. For example, the following command
opens python file(s) that are selected with fzf with pistol as a previewer:

```sh
vim "$(find -name '*.py' | fzf --preview='pistol {}')"
```

## Configuration

Although Pistol previews files of certain MIME types by default, it doesn't
force you to use these internal previewers for these MIME types. You can change
this behaviour by writing a configuration file in
`$XDG_CONFIG_HOME/pistol.conf` (or `~/.config/pistol.conf`) with a dumb simple
syntax as explained below.

### Syntax

The 1st word in every line is a regular expression, interpreted by the
[built-in go library](https://golang.org/pkg/regexp) which it's syntax is
documented [here](https://golang.org/pkg/regexp/syntax/). This regular
expression should match the MIME type of the file you may wish to preview. You
can inspect the MIME type of any file on a GNU/Linux OS and on Mac OS with the
command `file --mime-type <file>`.

The rest of the line, is interpreted as the command you wish to run on the file
when the given MIME type matches. `%s` is used in this part of the line as the
file argument.

For example, say you wish to replace Pistol's internal text previewer with that
of [bat](https://github.com/sharkdp/bat)'s, you'd put the following in your
`pistol.conf`:

```
text/* bat --paging=never --color=always %s
```

Naturally, Pistol reads the configuration file first in order to determine how
to preview a file. Only if such definition is not found, it'll attempt to use
it's own internal previewers.

Here's another example which features [w3m](http://w3m.sourceforge.net/) as an
HTML previewer:

```
text/html w3m -T text/html -dump %s
```

And here's an example that leverages `ls` for printing directories' contents:

```
inode/directory ls -l --color %s
```

#### A Note on RegEx matching

When Pistol parses your configuration file, as soon as it finds a match, it
stops parsing it and it invokes the command written on the rest of the line.
Therefor, if you wish to use the examples from above which use `w3m` and `bat`,
you'll need to put `w3m`'s line **before** `bat`'s line. Since otherwise,
`text/*` will be matched first and `text/html` won't be checked at all.

Of course that this is a mere example, the same may apply to any regular
expressions you'd choose to match against.

For a list of the internal regular expressions tested against when Pistol
reverts to it's native previewers, read the file
[`internal_writers/map.go`](https://github.com/doronbehar/pistol/blob/218310e5bf394d0d7edca4274145eef8f2f491df/internal_writers/map.go#L8-L12).


### Environmental Variables

Pistol's internal previewer for text files includes syntax highlighting thanks
to [chroma](https://github.com/alecthomas/chroma) Go library. You can customize
Pistol's syntax highlighting formatting and style through environmental
variables.

#### Chroma Formatters

The term _formatter_ refers to the way the given file is presented in the
terminal. These include:

- `terminal`: The default formatter that uses terminal control codes to change
  colors between every key word. This formatter has 8 colors and it is the
  default.
- `terminal256`: Same as `terminal` but with 256 colors available.
- `terminal16m`: Same as `terminal` but with 24 Bits colors i.e True-Color.

Other formatters include `json`, and `html` but they won't be useful to be used with Pistol.

To tell Pistol to use a specific formatter, set `PISTOL_CHROMA_FORMATTER` in
your  environment, e.g:

```sh
export PISTOL_CHROMA_FORMATTER=terminal16m
```

Recent versions of [Lf](https://github.com/gokcehan/lf) support [256
colors](https://github.com/gokcehan/lf/pull/93) in it's preview window.
AFAIK<sup id="a2">[2](#f2)</sup>, [Ranger](https://ranger.github.io/) supports
8 colors and Lf's `color256` isn't enabled by default.

Therefor, I decided that it'll be best to keep this variable unset in your
general environment. If you do set `color256` in your `lfrc`, you may feel free
to set `PISTOL_CHROMA_FORMATTER` in your environment.

#### Chroma Styles

The term _style_ refers to the set of colors used to print the given file.
All styles are documented and presented by chroma
[here](https://xyproto.github.io/splash/docs/all.html) and
[here](https://xyproto.github.io/splash/docs/longer/all.html).

The default style used by Pistol is `pygments`. To tell Pistol to use
a specific style set `PISTOL_CHROMA_STYLE` in your environment, e.g:

```sh
export PISTOL_CHROMA_STYLE=monokai
```

## Footnotes

<b id="f1">1</b> Considering Pistol's indirect dependence on
[libmagic(3)](http://linux.die.net/man/3/libmagic), I will never take the
trouble to personally try and make it work on Windows natively. If you'll
succeed in the heroic task of compiling libmagic for Windows and teach
[magicmime](https://github.com/rakyll/magicmime) to use it, please let me know.
[↩](#a1)

<b id="f2">2</b>I don't use Ranger anymore, ever since I moved to Lf. If you
have evidence it does support 256 colors, let me know and I'll change the
default. [↩](#a[↩](#a1)2)
