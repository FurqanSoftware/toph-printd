# Printd

Allow participants to request prints right from within Toph. Process print jobs locally using Printd.

![](overview.png)

You will have to download Printd and run it on a local computer connected to a printer.

If the computer with the printer has Linux on it, you will need to ensure that CUPS is installed and the printer is configured correctly.

You will also need a [configuration file](#configuration) for Printd. This configuration file is contest-specific. You can download this configuration file from Toph once the printing feature is enabled for your contest.

## Dependencies

Linux:

- CUPS: Printd uses `lpr` to dispatch the print job to the printer.

## Usage

```
» ./printd -h
Usage of ./printd:
  -config string
    	path to configuration file (default "printd-config.toml")
```

## Configuration

``` toml
[printd]
fontSize = 13     # In px. All text uses this same font size.
lineHeight = 20   # In px. Must me larger than fontSize.
marginTop = 50    # Margin at the top edge of each page.
marginRight = 25  # ... on the right edge of each page.
marginBottom = 50 # ... on the bottom edge of each page.
marginLeft = 25   # ... on the left edge of each page.
keepPDF = true    # When true, does not delete generated PDF after print.

[printer]
name = "..."    # Name of the printer. Leave empty to use the system default.
pageSize = "A4" # Size of the page. Use one of "A4", "letter", or "legal".

[toph]
baseURL = "https://toph.co"
token = "..."               # Collect your printd token from Toph Support. The token is contest-specific.

[contest]
id = "..." # The 24 character hex ID of the contest goes here.
```

## To-dos

- [ ] Windows support
