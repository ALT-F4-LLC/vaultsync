# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - 2020-03-22

### Added

- #1: Implements sync policies: Allows users to pre-define policies and sync them to target Vault instance
- #2: Implements sync auth methods: Allows users to pre-define auth methods and sync them to target Vault instance
- #3: Implements sync AppRoles: Allows users to pre-define approles to be auto-generated and optionally saved to the file system using `output:` option (only in logs by default)
- #4: Implements AppRole authention: Allows AppRoles to be configured and used as a source auth method

## [1.0.0] - 2020-03-18

### Added

- Initial release.
