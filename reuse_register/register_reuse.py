#!/usr/bin/env python3

# SPDX-FileCopyrightText: 2025 - 2026 Dominik Wombacher <dominik@wombacher.cc>
#
# SPDX-License-Identifier: Apache-2.0

"""
REUSE Software Registration Tool

A lightweight command-line utility for automating project registrations on REUSE Software.
This tool handles CSRF token extraction and form submission to register Git repositories
with the REUSE API service.
"""

import argparse
import logging
import re
import sys

import requests
from bs4 import BeautifulSoup

# Configuration
REGISTER_URL = "https://api.reuse.software/register"
DEFAULT_NAME = "Your Name"
DEFAULT_EMAIL = "your.email@example.com"

# Set up logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger("reuse-register")


def validate_email(email: str) -> bool:
    """
    Validate email format according to basic rules.

    Args:
        email: Email address to validate

    Returns:
        bool: True if email format is valid
    """
    pattern = r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"
    return bool(re.match(pattern, email))


def validate_project_url(url: str) -> bool:
    """
    Validate a project URL format.

    Args:
        url: Project URL to validate

    Returns:
        bool: True if URL format is valid
    """
    # Allow URLs with or without protocol, ensure valid characters, and enforce minimum length
    if len(url.strip()) < 3:
        return False
    pattern = r"^(https?|git)://[\w.-]+(/[\w.-]+)*(/)?$|^[\w.-]+(/[\w.-]+)*(/)?$"
    return bool(re.match(pattern, url))


def sanitize_project_url(url: str) -> str:
    """
    Clean and normalize a project URL by removing protocol prefixes and .git suffix.

    Args:
        url: URL to sanitize

    Returns:
        str: Sanitized URL
    """
    # Remove prefixes
    url = url.replace("git://", "").replace("http://", "").replace("https://", "")

    # Remove .git suffix
    url = url.removesuffix(".git")

    return url


def parse_arguments() -> argparse.Namespace:
    """
    Parse command line arguments for the registration tool.

    Returns:
        argparse.Namespace: Parsed arguments
    """
    parser = argparse.ArgumentParser(
        description="Register a project on reuse.software",
        formatter_class=argparse.ArgumentDefaultsHelpFormatter,
    )

    parser.add_argument("--name", default=DEFAULT_NAME, help="Your name")

    parser.add_argument("--email", default=DEFAULT_EMAIL, help="Your email address")

    parser.add_argument(
        "repo", help="Git repository without protocol (e.g., github.com/user/repo)"
    )

    parser.add_argument(
        "--updates",
        action="store_true",
        help="Receive occasional information about REUSE and FSFE activities",
    )

    parser.add_argument("--debug", action="store_true", help="Enable debug logging")

    args = parser.parse_args()

    # Set debug level if requested
    if args.debug:
        logger.setLevel(logging.DEBUG)

    # Validate inputs
    if not validate_project_url(args.repo):
        parser.error("Repository URL is required and must be valid")

    if not validate_email(args.email):
        parser.error("Invalid email format")

    # Sanitize repository URL
    args.repo = sanitize_project_url(args.repo)

    return args


def fetch_csrf_token(session: requests.Session) -> str:
    """
    Fetch the CSRF token from the registration page.

    Returns:
        str: CSRF token from the form

    Raises:
        SystemExit: If token cannot be retrieved
    """
    try:
        logger.info(f"Fetching registration page from {REGISTER_URL}")
        response = session.get(REGISTER_URL)
        response.raise_for_status()

        soup = BeautifulSoup(response.text, "html.parser")
        csrf_token = soup.find("input", {"id": "csrf_token"})

        if not csrf_token or "value" not in csrf_token.attrs:
            logger.error("CSRF token missing in the registration page response")
            sys.exit("Error: CSRF token not found. Please try again later.")

        token_value = csrf_token["value"]
        logger.debug(f"Found CSRF token: {token_value[:10]}...")
        return token_value

    except requests.exceptions.RequestException as e:
        logger.error(f"Error fetching the registration page: {e}")
        sys.exit(
            "Error: Unable to fetch the registration page. Please check your connection."
        )


def submit_registration(
    session: requests.Session, args: argparse.Namespace, csrf_token: str
) -> bool:
    """
    Submit the registration form with the provided details.

    Args:
        args: Command line arguments with form data
        csrf_token: CSRF token for the form

    Returns:
        bool: True if submission was successful

    Raises:
        SystemExit: If request fails
    """
    form_data = {
        "csrf_token": csrf_token,
        "name": args.name,
        "confirm": args.email,  # 'confirm' is the field name for email
        "project": args.repo,
    }

    if args.updates:
        form_data["wantupdates"] = "y"

    logger.info("Submitting registration form...")

    try:
        response = session.post(REGISTER_URL, data=form_data)
        response.raise_for_status()

        # Check for success indicators in the response
        if "Thank you for registering" in response.text:
            logger.info(f"Successfully registered {args.repo}")
            return True
        else:
            logger.warning(
                "Registration response did not contain expected success message"
            )
            return False

    except requests.exceptions.RequestException as e:
        logger.error(f"Error submitting the registration: {e}")
        sys.exit("Error: Registration submission failed. Please try again later.")


def main() -> None:
    """
    Main function to execute the registration process.
    """
    args = parse_arguments()
    logger.info(f"Registering repository: {args.repo}")

    # One session for both requests so the cookie from GET is sent with POST
    session = requests.Session()
    csrf_token = fetch_csrf_token(session)
    logger.info("CSRF token retrieved successfully")

    success = submit_registration(session, args, csrf_token)

    if success:
        print(f"✅ Repository {args.repo} has been successfully registered!")
        print("Please check your email for further steps.")
    else:
        print("⚠️ Registration might have been processed but could not be confirmed")


if __name__ == "__main__":
    main()
