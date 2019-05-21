# Pistol

## Background

Pistol is a file previewer that was designed to replace `scope.sh` which is
widely used in [Ranger](https://ranger.github.io/) configurations. 

Ranger is written in Python which means it has various disadvantages such as
slow startup time and lot's of dependencies including Python it self which need
to be installed on runtime. Never the less, it has proven useful and popular
among passionate command line users.

Ranger has given inspiration to an alternative written in Go called
[Lf](https://github.com/gokcehan/lf). Lf attempts to be a faster and a simpler
file manager which is intended to replace Ranger.

One of the main features used by both file managers is the preview window which
displays a preview of the highlighted file in it. Ranger uses a shell
script called `scope.sh` which is the common way to configure what program is
used for what kind of file.

Inspired by Lf, Pistol aims to provide a faster and simpler alternative to
`scope.sh`. Instead of spawning external programs like `file` and running
multiple `case` switches to handle every mimetype or file extension, Pistol
attempts to enable the same functionality but in a much simpler manner and
using only through mimetype detection.

## Design

Pistol was designed to be not only a faster replacement for `scope.sh`, but
also provide a simpler interface - both in regard to configuration and first
time usage. Therefor, out of the box, Pistol provides preview using internal
previewers for certain common mimetypes:

File Type  | Notes on implementation
---------- | -----------------------
`text/*`   | Prints text files with syntax highlighting thanks to [`chroma`](https://github.com/alecthomas/chroma).
Archives   | Prints the contents of archive files using [`archiver`](https://github.com/mholt/archiver).

Besides these, if a previewer wasn't configured for a given mimetype and
a file of this mimetype is encountered, a general description of the file type will
be printed. So for example, the preview for an executable will be similar to this:

```
ELF 64-bit LSB executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, BuildID[sha1]=a34861a1ae5358dc1079bc239df9dfe4830a8403, for GNU/Linux 3.2.0, not stripped
```

This feature is available out of the box as well.

### MimeType Detection

There are several Go libraries that provide mimetype detection. Here are the
top search results (using a common web search engine) I got when I started
develop Pistol and I realised I need such a library for the project.

- https://github.com/gabriel-vasile/mimetype/
- https://github.com/h2non/filetype
- https://github.com/rakyll/magicmime

Pistol uses the last one which leverages the well known C library
[libmagic(3)](http://linux.die.net/man/3/libmagic). This choice was made after
experimenting with the other candidates and due to their lack of extensive
database as [libmagic(3)](http://linux.die.net/man/3/libmagic) has, I chose
[magicmime](https://github.com/rakyll/magicmime).

Note that this choice also features compatibility with the standard command
`file` which is available by default on most GNU/Linux platforms
<sup id="a1">[1](#f1)</sup>.

## Install

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

Although Pistol previews files of certain mimetypes by default, it doesn't
force you into using these internal previewers for these mimetypes. You can
change this behaviour by writing a configuration file in
`$XDG_CONFIG_HOME/pistol.conf` (or `~/.config/pistol.conf`) which features
a dumb simple syntax as explained below.

### Syntax

The 1st word is a regular expression, interpreted by the [built-in go
library](https://golang.org/pkg/regexp) which it's syntax is documented
[here](https://golang.org/pkg/regexp/syntax/). This regular expression should
match the mimetype of the file you may wish to preview. You can inspect the
mimetype of any file on GNU/Linux with the command `file --mime-type <file>`.

The rest of the line, is interpreted as the command you wish to run on the
file. `%s` is used in this part of the line as the file argument.

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

And here's an example that leverages `ls` for printing a directories content,
just in case this is needed with whatever usage you have in mind for Pistol:

```
inode/directory ls -l --color %s
```

#### A Note on RegEx matching

When Pistol parses your configuration file, as soon as it finds a match, it
stops the parsing process and it invokes the command written on the rest of the
line. Therefor, if you wish to use the examples from above which use `w3m` and
`bat`, you'll need to put `w3m`'s line **before** `bat`'s line. Since
otherwise, `text/*` will be matched first and `text/html` won't be checked at
all.

Of course that this is a mere example, the same may apply to any regular
expressions used internally or externally.

For a list of the internal regular expressions which are tested against in
order to find a suitable internal previewer, see [this
variable](https://github.com/doronbehar/pistol/blob/218310e5bf394d0d7edca4274145eef8f2f491df/internal_writers/map.go#L8-L12).

## Footnotes

<b id="f1">1</b> Considering Pistol's dependence on
[libmagic(3)](http://linux.die.net/man/3/libmagic), it will never attempt to
have Windows support unless you are willing to compile it for Windows and
attempt to teach [magicmime](https://github.com/rakyll/magicmime) to use it.
[↩](#a1)

