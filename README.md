# Tiny Core Linux Extension Downloader

This command-line software allows you to recursively download Tiny Core Linux
extensions on a system that is not running Tiny Core Linux. It was created to
allow extensions to be downloaded on an Internet-connected system to be
installed on a Tiny Core Linux system with no Internet connection.

Usage:
`TceDownload [options] <extension> [extension [...]]`

Options:
- `-arch string` The architecture for which to get extensions. (default "x86")
- `-help` Shows a help message which will look very familiar after viewing this
  readme.
- `-out string` The directory to which to output files. (default "tce/%v/%a")
- `-version string` The Tiny Core Linux version for which to get extensions.
  (default "7.x")

This software is licensed under the MIT license. See `LICENSE` for the wording
of this license.
