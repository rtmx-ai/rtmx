"""RTMX config command.

Show and validate RTMX configuration.
"""

from __future__ import annotations

import json
import os
import sys
from pathlib import Path

import yaml

from rtmx.config import RTMXConfig
from rtmx.formatting import Colors


def run_config(
    config: RTMXConfig,
    validate_config: bool,
    output_format: str,
) -> None:
    """Run config command.

    Show or validate the effective RTMX configuration.

    Args:
        config: The loaded configuration
        validate_config: Whether to validate paths and settings
        output_format: Output format (terminal, yaml, json)
    """
    errors: list[str] = []
    warnings: list[str] = []

    if validate_config:
        # Validate database path
        db_path = Path(config.database)
        if db_path.exists():
            if not db_path.is_file():
                errors.append(f"Database path is not a file: {db_path}")
            elif not os.access(db_path, os.R_OK):
                errors.append(f"Database file is not readable: {db_path}")
        else:
            warnings.append(f"Database file does not exist: {db_path}")

        # Validate requirements directory
        req_dir = Path(config.requirements_dir)
        if req_dir.exists():
            if not req_dir.is_dir():
                errors.append(f"Requirements path is not a directory: {req_dir}")
        else:
            warnings.append(f"Requirements directory does not exist: {req_dir}")

        # Validate schema
        valid_schemas = ["core", "phoenix"]
        if config.schema not in valid_schemas:
            errors.append(f"Invalid schema '{config.schema}'. Must be one of: {valid_schemas}")

        # Check adapter environment variables
        if config.adapters.github.enabled:
            token_env = config.adapters.github.token_env
            if not os.environ.get(token_env):
                warnings.append(f"GitHub adapter enabled but {token_env} not set")

        if config.adapters.jira.enabled:
            token_env = config.adapters.jira.token_env
            email_env = config.adapters.jira.email_env
            if not os.environ.get(token_env):
                warnings.append(f"Jira adapter enabled but {token_env} not set")
            if not os.environ.get(email_env):
                warnings.append(f"Jira adapter enabled but {email_env} not set")

    if output_format == "json":
        output = {
            "config": config.to_dict(),
            "validation": {
                "errors": errors,
                "warnings": warnings,
            }
            if validate_config
            else None,
        }
        print(json.dumps(output, indent=2))

    elif output_format == "yaml":
        output_dict = config.to_dict()
        if validate_config:
            output_dict["_validation"] = {
                "errors": errors,
                "warnings": warnings,
            }
        print(yaml.dump({"rtmx": output_dict}, default_flow_style=False, sort_keys=False))

    else:  # terminal
        print(f"{Colors.BOLD}=== RTMX Configuration ==={Colors.RESET}")
        print()

        # Core settings
        print(f"{Colors.BOLD}Core:{Colors.RESET}")
        print(f"  Database:         {config.database}")
        print(f"  Requirements Dir: {config.requirements_dir}")
        print(f"  Schema:           {config.schema}")
        print()

        # Pytest settings
        print(f"{Colors.BOLD}Pytest:{Colors.RESET}")
        print(f"  Marker Prefix:    {config.pytest.marker_prefix}")
        print(f"  Register Markers: {config.pytest.register_markers}")
        print()

        # Agent settings
        print(f"{Colors.BOLD}Agents:{Colors.RESET}")
        agents = [
            ("Claude", config.agents.claude),
            ("Cursor", config.agents.cursor),
            ("Copilot", config.agents.copilot),
        ]
        for name, agent_config in agents:
            status = (
                f"{Colors.GREEN}enabled{Colors.RESET}"
                if agent_config.enabled
                else f"{Colors.DIM}disabled{Colors.RESET}"
            )
            print(f"  {name}: {status}")
            if agent_config.enabled:
                print(f"    Config Path: {agent_config.config_path}")
        print(f"  Template Dir: {config.agents.template_dir}")
        print()

        # Adapter settings
        print(f"{Colors.BOLD}Adapters:{Colors.RESET}")
        gh = config.adapters.github
        gh_status = (
            f"{Colors.GREEN}enabled{Colors.RESET}"
            if gh.enabled
            else f"{Colors.DIM}disabled{Colors.RESET}"
        )
        print(f"  GitHub: {gh_status}")
        if gh.enabled:
            print(f"    Repo:      {gh.repo}")
            print(f"    Token Env: {gh.token_env}")

        jira = config.adapters.jira
        jira_status = (
            f"{Colors.GREEN}enabled{Colors.RESET}"
            if jira.enabled
            else f"{Colors.DIM}disabled{Colors.RESET}"
        )
        print(f"  Jira: {jira_status}")
        if jira.enabled:
            print(f"    Server:    {jira.server}")
            print(f"    Project:   {jira.project}")
            print(f"    Token Env: {jira.token_env}")
        print()

        # MCP settings
        print(f"{Colors.BOLD}MCP Server:{Colors.RESET}")
        mcp_status = (
            f"{Colors.GREEN}enabled{Colors.RESET}"
            if config.mcp.enabled
            else f"{Colors.DIM}disabled{Colors.RESET}"
        )
        print(f"  Status: {mcp_status}")
        if config.mcp.enabled:
            print(f"  Host:   {config.mcp.host}")
            print(f"  Port:   {config.mcp.port}")
        print()

        # Sync settings
        print(f"{Colors.BOLD}Sync:{Colors.RESET}")
        print(f"  Conflict Resolution: {config.sync.conflict_resolution}")
        print()

        # Validation results
        if validate_config:
            print(f"{Colors.BOLD}Validation:{Colors.RESET}")
            if not errors and not warnings:
                print(f"  {Colors.GREEN}✓ Configuration is valid{Colors.RESET}")
            else:
                if errors:
                    print(f"  {Colors.RED}Errors:{Colors.RESET}")
                    for error in errors:
                        print(f"    {Colors.RED}✗{Colors.RESET} {error}")
                if warnings:
                    print(f"  {Colors.YELLOW}Warnings:{Colors.RESET}")
                    for warning in warnings:
                        print(f"    {Colors.YELLOW}⚠{Colors.RESET} {warning}")

    # Exit with error if validation failed
    if validate_config and errors:
        sys.exit(1)
