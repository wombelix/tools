<!--
SPDX-FileCopyrightText: 2025 - 2026 Dominik Wombacher <dominik@wombacher.cc>

SPDX-License-Identifier: CC0-1.0
-->

# REUSE Software Registration Tool

Automates project registration on
[REUSE Software](https://api.reuse.software/register).

Fetches the CSRF token from the registration page and submits the form.
Validates email and URL input. Name and email defaults are configurable.

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

## Running Tests

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
