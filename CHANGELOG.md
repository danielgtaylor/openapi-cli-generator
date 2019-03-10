# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/).

## 2019-03-10
- Switch to Go modules.

## 2019-03-09
- Generate methods of the format `{{ API Name }}{{ Operation Name }}(...)` for each API operation. These can be used by custom code as if you were invoking the CLI, but it returns rather than formats the response.
- Decouple CLI commands from the command path used to register middleware. Each API operation now has one and only one command path regardless of which CLI command calls it.
- Add support for waiters through the `x-cli-waiters` extension.

## 2018-09-29
- Initial release
