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

### Prerequisits

Since Pistol depends on  [magicmime](https://github.com/rakyll/magicmime),
you'll need a `libmagic` package installed. Please refer to [this section in
magicmime's
README](https://github.com/rakyll/magicmime/tree/v0.1.0#prerequisites) for the
appropriete commands for every OS.

Assumming `libmagic` is installed, Use the following command to install Pistol to
`$GOPATH/.bin/pistol`:

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

## Footnotes

<b id="f1">1</b> Considering Pistol's indirect dependence on
[libmagic(3)](http://linux.die.net/man/3/libmagic), it will never attempt to
support Windows unless you'd be willing to take the trouble of compiling it for
Windows and teach [magicmime](https://github.com/rakyll/magicmime) to use it
your version of libmagic. If you'll succeed in this heroic task, please let us
know. [↩](#a1)
