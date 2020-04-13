# mpv-subserv

Inspiration taken (including the name, sorry) from
[github.com/kaervin/subserv-mpv-plugin](https://github.com/kaervin/subserv-mpv-plugin)

mpv-subserv is a plugin for the [mpv video player](https://mpv.io/) that sends
all subtitles to your browser, so you can copy-paste or use a mouseover
dictionary on them.

## Installation

Run `install.sh`

## Usage

The plugin is dormant by default and controlled through the following
environment variables:

- `MPV_SUBSERV` set to `1` to enable the plugin.
- `MPV_SUBSERV_LANG` will be put in the HTML bodys `lang` attribute. Important for
CJK languages.
- `MPV_SUBSERV_FILTER` path to the filter file to use (see [Filter files](#filter-files)).
- `MPV_SUBSERV_SUBFILE` full path the subtitle file (required, see [Design
considerations](#design-considerations)).
- `MPV_SUBSERV_BROWSER` your browser command, falls back to `BROWSER` or
`xdg-open`.

## Filter files

Sometimes, subtitle files contain unwanted noise that you do not want in your
subtitle list. By specifiying a filter file, you can filter those out.

A filter file is composed of line seperated regular expressions. If one of them
matches against a line, it will not be displayed on the web UI. Refer to the [Go
Documentation](https://golang.org/s/re2syntax) for information about the syntax.

## Design considerations

mpv-subserv has to do it's own subtitle parsing, as mpv's interface for getting
subtitle data is quite poor. Mpv does not expose how long subtitles are being
displayed, which will be needed for future functionality (see [TODO](#todo)).
Also, mpv's sub-text property is not reliable when retiming subtitles on the fly
using sub-delay.

## TODO

- Generate video/audio files or screenshots for the current subtitle.
- Do not trust mpv's sub-text. It's broken.
- Implement support for srt subtitles.
