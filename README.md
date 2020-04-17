# Pistol

## NOTE TO EXISTING USERS

If you've updated to [v0.1](https://github.com/doronbehar/pistol/releases) or
higher, you may experience errors with custom commands set in your config,
similar to:

```
[bat error]: '%pistol-filename%': No such file or directory (os error 2)
```

Basically, after hitting [issue #16](https://github.com/doronbehar/pistol/issues/16),
I realised that the old way Pistol substituted the file name in your config was
not scalable. So now, please replace every occurrence of `%s` with
`%pistol-filename%`. Or with:

```sh
sed -i 's/%s/%pistol-filename%/g ~/.config/pistol/pistol.conf
```

If you want to know more details, read
[this](https://github.com/doronbehar/pistol/issues/16#issuecomment-614471555).

## Introduction

Pistol is a file previewer for command line file managers such as
[Ranger](https://ranger.github.io/) and [Lf](https://github.com/gokcehan/lf)
intended to replace the file previewer
[`scope.sh`](https://github.com/ranger/ranger/blob/v1.9.2/ranger/data/scope.sh)
commonly used with them.

`scope.sh` is a Bash script that uses `case` switches and external programs to
decide how to preview every file it encounters. It knows how to handle every
file according to it's [MIME type](https://en.wikipedia.org/wiki/Media_type)
and/or file extension using `case` switches and external programs. This design
makes it hard to configure / maintain and it makes it slow for startup and
heavy when running.

Pistol is a Go program (with (almost) 0 dependencies) and it's MIME type detection is
internal. Moreover, it features native preview support for almost any archive
file and for text files along with syntax highlighting while `scope.sh` relies
on external programs to do these basic tasks.

The following table lists Pistol's native previewing support:

File/MIME Type  | Notes on implementation
---------- | -----------------------
`text/*`   | Prints text files with syntax highlighting thanks to [`chroma`](https://github.com/alecthomas/chroma).
Archives   | Prints the contents of archive files using [`archiver`](https://github.com/mholt/archiver).

In case Pistol encounters a MIME type it doesn't know how to handle natively
and you haven't configured a program to handle it, it'll print a general
description of the file type it encountered. For example, the preview for an
executable might be:

```
ELF 64-bit LSB executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, BuildID[sha1]=a34861a1ae5358dc1079bc239df9dfe4830a8403, for GNU/Linux 3.2.0, not stripped
```

This feature is available out of the box just like the previews for the common
mime types mentioned above.

### A Note on MIME type Detection

Some _pure_ Go libraries provide MIME type detection. Here are the top search
results I got using a common web search engine:

- [https://github.com/gabriel-vasile/mimetype](https://github.com/gabriel-vasile/mimetype)
- [https://github.com/h2non/filetype](https://github.com/h2non/filetype)
- [https://github.com/rakyll/magicmime](https://github.com/rakyll/magicmime)

Pistol uses the last one which leverages the well known C library
[libmagic(3)](http://linux.die.net/man/3/libmagic). I made this choice after
experimenting with the other candidates and due to their lack of an extensive
database such as [libmagic(3)](http://linux.die.net/man/3/libmagic) has,
I chose [magicmime](https://github.com/rakyll/magicmime).

Note that this choice also features compatibility with the standard command
`file` which is available by default on most GNU/Linux distributions <sup
id="a1">[1](#f1)</sup>.

### A Note on Archive Previews

Pistol previews all archive / compression formats supported by the Go library
[`archiver`](https://github.com/mholt/archiver). Some formats do nothing but
compression, meaning they operate on 1 file alone and some files are
a combination of archive, compressed in a certain algorithm.

For example, a `.gz` file is a _single_ file compressed with `gzip`. A `.tar`
file is an _uncompressed_ archive (collection) of files. A `.tar.gz` is
a `.tar` file compressed using `gzip`. 

When pistol encounters a single file compressed using a known compression
algorithm, it doesn't know how to handle it's content, it displays the type
of the archive. If a known compression algorithm has compressed a `.tar` file,
Pistol lists the files themselves.

[brotli](https://en.wikipedia.org/wiki/Brotli) compressed archives, (`.tar.br`) and
brotli compressed files (`.br`) are not detected by libmagic so Pistol doesn't know how to handle them.
<sup id="a2">[2](#f2)</sup>

## Install

If someone has packaged Pistol for your distribution, you might find a package
for of it linked [in the
WiKi](https://github.com/doronbehar/pistol/wiki/GNU-Linux-Distributions'-Packages).
If not, you can install it from source according to the following instructions:

### Prerequisites

Since Pistol depends on  [magicmime](https://github.com/rakyll/magicmime),
you'll need a `libmagic` package installed. Please refer to [this section in
magicmime's
README](https://github.com/rakyll/magicmime/tree/v0.1.0#prerequisites) for the
appropriate commands for every OS.

Assuming you've installed `libmagic` properly and you have [setup a Go
environment](https://golang.org/doc/install), Use the following command to
install Pistol to `$GOPATH/bin/pistol`:

```sh
env GO111MODULE=on go get -u github.com/doronbehar/pistol/cmd/pistol
```

<sup id="a3">[3](#f3)</sup>

## Usage

```
$ pistol --help
Usage: pistol OPTIONS <file>

OPTIONS

-c, --config <config>  configuration file to use (defaults to ~/.config/pistol/pistol.conf)
-h, --help             print help and exit
-v, --verbosity        increase verbosity

ARGUMENTS

file                   the file to preview
```

### Integrations

#### Ranger / Lf

You can use Pistol as a file previewer in [Ranger](https://ranger.github.io/)
and [Lf](https://github.com/gokcehan/lf). For Ranger, set your `preview_script`
in your `rc.conf` as follows:

```
set preview_script ~/.go/bin/pistol
```

The same goes for Lf, but in `lfrc`:

```
set previewer ~/.go/bin/pistol
```

#### fzf

If you use [fzf](https://github.com/junegunn/fzf) to search for files, you can
tell it to use `pistol` as the previewer. For example, the following command
edits with your `$EDITOR` selected python file(s) using `pistol` as
a previewer:

```sh
$EDITOR "$(find -name '*.py' | fzf --preview='pistol {}')"
```

## Configuration

Although Pistol previews files of certain MIME types by default, it doesn't
force you to use these internal previewers for these MIME types. You can change
this behaviour by writing a configuration file in
`$XDG_CONFIG_HOME/pistol/pistol.conf` (or `~/.config/pistol/pistol.conf`) with
the syntax as explained below.

### Syntax

You can configure preview commands according to file path or mime type regex.
The 1st word may is always interpreted first as a mime type regex such as:
`text/*`.

If a line is not matched but the 1st word is exactly `fpath`, then the 2nd
argument is interpreted as a file path regex, such as:
`/var/src/.*/README.md`.

On every line, whether you used `fpath` or not, the next arguments are the
command's arguments, where `%s` is substituted by `pistol` to the file at
question. You'll see more examples in the following sections.

Both regular expressions (for file paths and for mime types) are interpreted by
the [built-in library's `regexp.match`](https://golang.org/pkg/regexp/#Match).
Please refer to [this link](https://golang.org/pkg/regexp/syntax) for the full
reference regarding syntax.

#### Matching Mime Types

You can inspect the MIME type of any file on a GNU/Linux OS and on Mac OS with
the command `file --mime-type <file>`.

For example, say you wish to replace Pistol's internal text previewer with that
of [bat](https://github.com/sharkdp/bat)'s, you'd put the following in your
`pistol.conf`:

```
text/* bat --paging=never --color=always %s
```

Naturally, your configuration file overrides the internal previewers.

Here's another example which features [w3m](http://w3m.sourceforge.net/) as an
HTML previewer:

```
text/html w3m -T text/html -dump %s
```

And here's an example that leverages `ls` for printing directories' contents:

```
inode/directory ls -l --color %s
```

#### Matching File Path

For example, say you wish to preview all files that reside in a certain `./bin`
directory with [bat](https://github.com/sharkdp/bat)'s syntax highlighting for
bash. You could use:

```
fpath /var/src/my-bash-project/bin/[^/]+$ bat --map-syntax :bash --paging=never --color=always %s
```

#### A Note on RegEx matching

When Pistol parses your configuration file, as soon as it finds a match be it
a file path match or a mime type match, it stops parsing it and it invokes the
command written on the rest of the line. Therefor, if you wish to use the
examples from above which use `w3m` and `bat`, you'll need to put `w3m`'s line
**before** `bat`'s line. Since otherwise, `text/*` will be matched first and
`text/html` won't be checked at all.

Similarly, you'd probably want to put `fpath` lines at the top.

Of course that this is a mere example, the same may apply to any regular
expressions you'd choose to match against.

For a list of the internal regular expressions tested against when Pistol
reverts to it's native previewers, read the file
[`internal_writers/map.go`](https://github.com/doronbehar/pistol/blob/master/internal_writers/map.go#L8-L12).

More examples can be found in [this WiKi page](https://github.com/doronbehar/pistol/wiki/Config-examples).

#### Quoting and Shell Piping

Pistol is pretty dumb when it parses your config, it splits all line by spaces,
meaning that e.g:

```config
application/json jq '.' %s
```

This will result in an error by [`jq`](https://github.com/stedolan/jq):

```
jq: error: syntax error, unexpected INVALID_CHARACTER, expecting $end (Unix shell quoting issues?) at <top-level>, line 1:
'.'
jq: 1 compile error
```

Indicating that `jq` got a literal `'.'`. When you run in your shell `jq '.'
file.json` you don't get an error because your shell is stripping the quotes
around `.`. However, Pistol is not smarter then your shell because if you'd try
for example:

```config
application/json jq '.[] | .' %s
```

That would be equivalent to running in the typical shell:

```sh
jq "\'.[]" "|" ".'" file.json
```

That's because Pistol doesn't consider your quotes as interesting instructions,
it just splits words by spaces. Hence, to overcome this disability, you can
use:

```config
application/json sh: jq '.' %s
```

Thanks to the `sh:` keyword at the beginning of the command's definition, the
rest of the line goes straight as a single argument to `sh -c`.

You can worry not about quoting / escaping the rest of the line - it's not like when
you run e.g `sh -c 'command'` in your shell where you need to make sure single
quotes are escaped or not used at all inside `command`.

More over, with `sh:` you can use shell pipes:

```config
fpath .*.md$ sh: bat --paging=never --color=always %s | head -8
```

### Environmental Variables

Pistol's internal previewer for text files includes syntax highlighting thanks
to the Go library [chroma](https://github.com/alecthomas/chroma). You can customize
Pistol's syntax highlighting formatting and style through environmental
variables.

#### Chroma Formatters

The term _formatter_ refers to the way the given file is presented in the
terminal. These include:

- `terminal`: The default formatter that uses terminal control codes to change
  colors between every key word. This formatter has 8 colors and it's the
  default.
- `terminal256`: Same as `terminal` but with 256 colors available.
- `terminal16m`: Same as `terminal` but with 24 Bits colors i.e True-Color.

Other formatters include `json`, and `html` but I'd be surprised if you'll find
them useful for Pistol's purpose.

To tell Pistol to use a specific formatter, set `PISTOL_CHROMA_FORMATTER` in
your  environment, e.g:

```sh
export PISTOL_CHROMA_FORMATTER=terminal16m
```

Recent versions of [Lf](https://github.com/gokcehan/lf) support [256
colors](https://github.com/gokcehan/lf/pull/93) in it's preview window.
AFAIK<sup id="a4">[4](#f4)</sup>, [Ranger](https://ranger.github.io/) supports
8 colors and Lf's `color256` isn't enabled by default.

Therefor, I decided that it'll be best to keep this variable unset in your
general environment. If you do set `color256` in your `lfrc`, you may feel free
to set `PISTOL_CHROMA_FORMATTER` in your environment.

#### Chroma Styles

The term _style_ refers to the set of colors used to print a given file. the
chroma project documents all styles
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

<b id="f2">2</b> [`file` bug report](https://bugs.astron.com/view.php?id=111);
[`brotli` bug report](https://github.com/google/brotli/issues/727). [↩](#a2)

<b id="f3">3</b> `env GO111MODULE=on` is needed due to a recent bug / issue
[with Go](https://github.com/golang/go/issues/31529), see
[#6](https://github.com/doronbehar/pistol/issues/6) for more details. [↩](#a3)

<b id="f4">4</b> I don't use Ranger anymore, ever since I moved to Lf. If you
have evidence it does support 256 colors, let me know and I'll change the
default. [↩](#a4)
