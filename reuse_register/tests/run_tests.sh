#!/bin/bash

# SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
#
# SPDX-License-Identifier: Apache-2.0

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PARENT_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"
VENV_DIR="$PARENT_DIR/.venv"

# Check if virtual environment exists
if [ ! -d "$VENV_DIR" ]; then
    echo "Virtual environment not found. Running setup script..."
    # Run the setup script first
    "$PARENT_DIR/setup.sh"
fi

# Activate virtual environment
# shellcheck source=/dev/null
source "$VENV_DIR/bin/activate"

# Run tests with pytest
echo "Running tests..."
cd "$PARENT_DIR" || exit
python -m pytest -v "$SCRIPT_DIR/test_register_reuse.py"

# Capture exit code
exit_code=$?

# Return the exit code from pytest
exit $exit_code
