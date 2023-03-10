# Printd

[![](https://img.shields.io/github/v/release/FurqanSoftware/toph-printd)](https://github.com/FurqanSoftware/toph-printd/releases) [![](https://img.shields.io/badge/support-community.toph.co-blue)](https://community.toph.co/c/support/printd/57)

Allow participants to request prints right from within Toph. Process print jobs locally using Printd.

![](overview.png)

You will have to download Printd and run it on a local computer connected to a printer.

If the computer with the printer has Linux on it, you will need to ensure that CUPS is installed and the printer is configured correctly.

You will also need a [configuration file](#configuration) for Printd. This configuration file is contest-specific. You can download this configuration file from Toph once the printing feature is enabled for your contest.

_If you have any questions or want to report issues about Printd, please [post about it on Toph Community](https://community.toph.co/new-topic?category=support/printd)._

## Example

[An example](example/example.pdf) is included in this repository.

![](example/header.png)

Every page printed is prepended with a header. Prints by participants show the participant number and name or account handle in the header. Test prints show "‹Test Print›" in the header instead.

The contest title, the timestamp of when the print was requested, and page numbers are also included in the header.

## Dependencies

Linux:

- CUPS: Printd uses `lpr` to dispatch the print job to the printer.

Windows:

- [PDFtoPrinter](http://www.columbia.edu/~em36/pdftoprinter.html): Printd uses `PDFtoPrinter` to dispatch the print job to the printer. Download PDFtoPrinter.exe and put it in the same directory as printd.exe.

## Usage

```
» ./printd -h
  ____       _       _      _ 
 |  _ \ _ __(_)_ __ | |_ __| |
 | |_) | '__| | '_ \| __/ _` |
 |  __/| |  | | | | | || (_| |
 |_|   |_|  |_|_| |_|\__\__,_|

For Toph, By Furqan Software (https://furqansoftware.com)

» Release: -

» Project: https://github.com/FurqanSoftware/toph-printd
» Support: https://community.toph.co/c/support/printd/57

Usage of /tmp/go-build71336074/b001/exe/toph-printd:
  -config string
      path to configuration file (default "printd-config.toml")
```

## Configuration

``` toml
[printd]
fontSize = 13            # In px. All text uses this same font size.
lineHeight = 20          # In px. It must be larger than the font size.
marginTop = 50           # Margin at the top edge of each page.
marginRight = 25         # ... at the right edge of each page.
marginBottom = 50        # ... at the bottom edge of each page.
marginLeft = 25          # ... at the left edge of each page.
tabSize = 4              # Replaces tabs with this many spaces.
headerExtra = ""         # Appends extra text to the page header.
reduceBlankLines = false # Replaces consecutive blank lines with one.
keepPDF = true           # Does not delete generated PDF after print.
delayAfter = "500ms"     # Forces a delay after each print.
delayError = "5s"        # Forces a delay after an error.
logColor = true          # Colors certain parts of the logs.
throbber = true          # Shows an activity throbber below logs.

[printer]
name = "..."    # Name of the printer. Leave empty to use the system default.
pageSize = "A4" # Size of the page. Use one of "A4", "letter", or "legal".

[toph]
baseURL = "https://toph.co"
token = "..."               # Collect your Printd token from Toph Support. The token is contest-specific.
contestID = "..."           # The 24-character hex ID of the contest goes here.
timeout = "10s"             # Timeout duration for HTTP client.
```

## Frequently Asked Questions

<details open>
<summary><b>Prints are missing a few lines of text near the bottom edge of the paper. They don't appear on the next page either. How can I increase the bottom margin of the prints?</b></summary>

This may happen when the printer cannot print content close to the bottom edge of the papers.

Open the configuration file and modify the `marginBottom` parameter under the `[printd]` section. Increase the value until enough margin is left so the printer doesn't lose content near the bottom edge of the papers.

If the `marginBottom` parameter is not present in the configuration file, you may add it under the `[printd]` section.

</details>

## To-dos

- [x] Windows support
- [ ] Improve tab-to-spaces behavior
