<!--
SPDX-FileCopyrightText: 2025 - 2026 Dominik Wombacher <dominik@wombacher.cc>

SPDX-License-Identifier: CC0-1.0
-->

# REUSE Software Registration Tool

A lightweight command-line utility for automating project registrations on
[REUSE Software](https://api.reuse.software/register).

## Overview

This tool automates the registration process for Git repositories with the REUSE
API service by:

1. Extracting CSRF tokens from the registration form
1. Submitting project registrations with proper validation
1. Handling the entire flow without manual intervention

The implementation uses minimal dependencies while ensuring robust error handling
and validation.

## Features

- Automatic CSRF token handling
- Input validation for email and repository URLs
- Configurable defaults for name and email
- Optional opt-in for REUSE/FSFE updates
- Comprehensive test suite for development

## Installation

Set up the tool in an isolated virtual environment:

```bash
# Run the setup script
./setup.sh

# Activate the virtual environment
source .venv/bin/activate
```

## Usage

Basic usage with default values:

```bash
./register_reuse.py github.com/user/repo
```

Override name and email:

```bash
./register_reuse.py \
  --name "Jane Doe" \
  --email "jane@example.com" \
  github.com/user/repo
```

Include opt-in for updates:

```bash
./register_reuse.py --updates github.com/user/repo
```

Enable debug logging:

```bash
./register_reuse.py --debug github.com/user/repo
```

## Development

The project follows these principles:

- **Single Responsibility**: Each function has a clear, focused purpose
- **Error Handling**: Robust error handling with meaningful messages
- **Testing**: Comprehensive test coverage without hitting production systems
- **Documentation**: Clear code documentation explaining the "why" not just
  "what"

### Running Tests

```bash
# Run tests with the test script (automatically sets up the environment)
./tests/run_tests.sh
```

Or manually:

```bash
# Activate the virtual environment
source .venv/bin/activate

# Run tests
pytest tests/
```

## License

Apache-2.0
