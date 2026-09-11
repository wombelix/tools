# SPDX-FileCopyrightText: 2025 - 2026 Dominik Wombacher <dominik@wombacher.cc>
#
# SPDX-License-Identifier: Apache-2.0

"""
Tests for the REUSE Software Registration Tool.

These tests verify the functionality without hitting production systems.
"""

import argparse
import os
import sys
from unittest.mock import MagicMock, patch

import pytest
import requests
import responses

# Add parent directory to path to import script
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

from register_reuse import (
    fetch_csrf_token,
    parse_arguments,
    sanitize_project_url,
    submit_registration,
    validate_email,
    validate_project_url,
)

# Sample HTML for testing
SAMPLE_HTML = """
<form class="form-horizontal" action="" method="post">
  <input id="csrf_token" name="csrf_token" type="hidden" value="test_csrf_token">
  <!-- Other form fields -->
</form>
"""

SUCCESS_HTML = """
<!DOCTYPE html>
<html>
<head><title>Registration successful | REUSE API</title></head>
<body>
  <h1>Registration successful</h1>
  <p>
    Thank you for registering github.com/user/repo for the REUSE API. Please
    check your email for further steps.
  </p>
</body>
</html>
"""


def test_validate_email():
    """Test email validation logic"""
    # Valid cases
    assert validate_email("user@example.com") is True
    assert validate_email("user.name+tag@example.co.uk") is True

    # Invalid cases
    assert validate_email("invalid-email") is False
    assert validate_email("@example.com") is False
    assert validate_email("user@") is False
    assert validate_email("") is False


def test_validate_project_url():
    """Test project URL validation logic"""
    assert validate_project_url("github.com/user/repo") is True
    assert validate_project_url("gitlab.com/org/project") is True

    # Invalid cases
    assert validate_project_url("") is False
    assert validate_project_url("ab") is False  # Too short


def test_sanitize_project_url():
    """Test URL sanitization logic"""
    # Remove prefixes
    assert sanitize_project_url("git://github.com/user/repo") == "github.com/user/repo"
    assert (
        sanitize_project_url("https://github.com/user/repo") == "github.com/user/repo"
    )
    assert sanitize_project_url("http://github.com/user/repo") == "github.com/user/repo"

    # Remove .git suffix
    assert sanitize_project_url("github.com/user/repo.git") == "github.com/user/repo"

    # Combined cases
    assert (
        sanitize_project_url("git://github.com/user/repo.git") == "github.com/user/repo"
    )


@patch("sys.argv")
def test_parse_arguments(mock_argv):
    """Test argument parsing logic"""
    # Test with minimal args
    mock_argv.__getitem__.side_effect = lambda i: [
        "register_reuse.py",
        "github.com/user/repo",
    ][i]
    mock_argv.__len__.return_value = 2

    args = parse_arguments()
    assert args.repo == "github.com/user/repo"
    assert args.name == "Your Name"  # Default
    assert args.email == "your.email@example.com"  # Default
    assert args.updates is False

    # Test with URL sanitization
    mock_argv.__getitem__.side_effect = lambda i: [
        "register_reuse.py",
        "git://github.com/user/repo.git",
    ][i]
    mock_argv.__len__.return_value = 2

    args = parse_arguments()
    assert args.repo == "github.com/user/repo"


@responses.activate
def test_fetch_csrf_token():
    """Test CSRF token extraction without hitting the real server"""
    # Mock the GET request to return our sample HTML
    responses.add(
        responses.GET,
        "https://api.reuse.software/register",
        body=SAMPLE_HTML,
        status=200,
    )

    # Call the function that fetches the token
    token = fetch_csrf_token(requests.Session())

    # Verify correct token was extracted
    assert token == "test_csrf_token"
    assert len(responses.calls) == 1


@responses.activate
def test_submit_registration_success():
    """Test successful registration submission"""
    # Mock successful POST response
    responses.add(
        responses.POST,
        "https://api.reuse.software/register",
        body=SUCCESS_HTML,
        status=200,
    )

    # Create mock arguments
    args = MagicMock(spec=argparse.Namespace)
    args.name = "Test Name"
    args.email = "test@example.com"
    args.repo = "github.com/user/repo"
    args.updates = False

    # Test the submission
    result = submit_registration(requests.Session(), args, "test_csrf_token")

    # Verify success
    assert result is True
    assert len(responses.calls) == 1

    # Verify the correct form data was sent
    request_body = responses.calls[0].request.body
    assert "csrf_token=test_csrf_token" in request_body
    assert "name=Test+Name" in request_body
    assert "confirm=test%40example.com" in request_body
    assert "project=github.com%2Fuser%2Frepo" in request_body


@responses.activate
def test_submit_registration_failure():
    """Test failed registration submission"""
    # Mock error response
    responses.add(
        responses.POST,
        "https://api.reuse.software/register",
        body="<html>Error occurred</html>",
        status=400,
    )

    # Create mock arguments
    args = MagicMock(spec=argparse.Namespace)
    args.name = "Test Name"
    args.email = "test@example.com"
    args.repo = "github.com/user/repo"
    args.updates = False

    # Test the submission with exception handling
    with pytest.raises(SystemExit):
        submit_registration(requests.Session(), args, "test_csrf_token")


@patch("register_reuse.parse_arguments")
@patch("register_reuse.fetch_csrf_token")
@patch("register_reuse.submit_registration")
def test_main(mock_submit, mock_fetch_token, mock_parse_args):
    """Test the main function execution flow with mocked components"""
    # Configure mocks
    mock_args = MagicMock()
    mock_args.repo = "github.com/user/repo"
    mock_parse_args.return_value = mock_args

    mock_fetch_token.return_value = "test_csrf_token"
    mock_submit.return_value = True

    # Need to import main here to avoid circular import issues
    from register_reuse import main

    # Call main function
    main()

    # Verify all components were called with correct args
    mock_parse_args.assert_called_once()
    mock_fetch_token.assert_called_once()  # called with the session object
    # Session object is positional arg 0, then args and token
    mock_submit.assert_called_once()
    assert mock_submit.call_args[0][1] == mock_args
    assert mock_submit.call_args[0][2] == "test_csrf_token"


@responses.activate
def test_fetch_csrf_token_connection_error():
    """Test handling of connection errors during token fetch"""

    # Mock a connection error using the callback approach
    def request_callback(request):
        raise requests.exceptions.ConnectionError("Connection refused")

    responses.add_callback(
        responses.GET, "https://api.reuse.software/register", callback=request_callback
    )

    # Test that the function exits gracefully
    with pytest.raises(SystemExit):
        fetch_csrf_token(requests.Session())


@responses.activate
def test_fetch_csrf_token_missing():
    """Test handling when CSRF token is not found in the response"""
    # Mock HTML without a CSRF token
    html_without_token = "<form><input type='text' name='other'></form>"
    responses.add(
        responses.GET,
        "https://api.reuse.software/register",
        body=html_without_token,
        status=200,
    )

    # Test that the function exits gracefully
    with pytest.raises(SystemExit):
        fetch_csrf_token(requests.Session())
