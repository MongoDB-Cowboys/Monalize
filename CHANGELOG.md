# Changelog

## 1.2.3

Fix dependabot alerts:

- golang.org/x/crypto Vulnerable to Denial of Service (DoS) via Slow or Incomplete Key Exchange

- Misuse of ServerConfig.PublicKeyCallback may cause authorization bypass in golang.org/x/crypto 

- golang.org/x/net vulnerable to Cross-site Scripting 

- HTTP Proxy bypass using IPv6 Zone IDs in golang.org/x/net 

## 1.2.2

- Update golang docker module

## 1.2.1

### Important Notes

- **Important**: Changed the project structure to improve readability and convenience.
- **Important**: Updated all Go modules.
- **Important**: Added new `podman` flag.

### Features

- Added Podman support

## 1.2.0

### Important Notes

- **Important**: Refactored codebase.
- **Important**: Removed "Clean history" feature.
- **Important**: Updated all Go modules.
- **Important**: Added new `container` flag.

### Features

- :tada: Added Docker support
    - Now "Monalize" can scan logs from the default Docker log streaming or custom log files in the container.
- Disabled shell for running the currentOp query in MongoDB. Now it's functioning with the MongoDB module.


## v0.0.1 (Example)

### Important Notes

- **Important**: Removed something
- **Important**: Updated something
- App now requires something

### Maintenance

- Removed a redundant feature
- Added a new functionality
- Improved overall performance
    - Enhanced user interface responsiveness

### Features

- :tada: Implemented a new feature
    - This feature allows users to...
- Added a user profile customization option

### Bug Fixes

- Fixed a critical issue that caused...