"""CLI commands for marker discovery and management.

Provides the `rtmx markers` command group.
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import TYPE_CHECKING

import click

if TYPE_CHECKING:
    from rtmx.markers.models import MarkerInfo


@click.group()
def markers() -> None:
    """Marker discovery and management commands."""
    pass


@markers.command("discover")
@click.argument("path", type=click.Path(exists=True, path_type=Path), default=".")
@click.option(
    "--format",
    "output_format",
    type=click.Choice(["text", "json", "csv"]),
    default="text",
    help="Output format",
)
@click.option(
    "--include-errors",
    is_flag=True,
    help="Include markers with validation errors",
)
@click.option(
    "--config",
    "config_path",
    type=click.Path(exists=True, path_type=Path),
    help="Path to rtmx.yaml config file",
)
@click.option(
    "-v",
    "--verbose",
    is_flag=True,
    help="Show detailed output",
)
def discover(
    path: Path,
    output_format: str,
    include_errors: bool,
    config_path: Path | None,
    verbose: bool,
) -> None:
    """Discover requirement markers in source files.

    Scans the specified PATH (file or directory) for requirement markers
    in all supported programming languages.

    Examples:

        rtmx markers discover                  # Scan current directory

        rtmx markers discover src/             # Scan specific directory

        rtmx markers discover --format json    # Output as JSON

        rtmx markers discover --include-errors # Include invalid markers
    """
    from rtmx.markers.discover import discover_markers

    markers_list = discover_markers(
        path=path,
        config_path=config_path,
        include_errors=include_errors,
    )

    if output_format == "json":
        _output_json(markers_list)
    elif output_format == "csv":
        _output_csv(markers_list)
    else:
        _output_text(markers_list, verbose=verbose)


def _output_text(markers_list: list[MarkerInfo], verbose: bool = False) -> None:
    """Output markers as formatted text.

    Args:
        markers_list: List of markers to output.
        verbose: Show detailed output.
    """
    if not markers_list:
        click.echo("No markers found.")
        return

    # Group by file
    by_file: dict[Path, list[MarkerInfo]] = {}
    for marker in markers_list:
        if marker.file_path not in by_file:
            by_file[marker.file_path] = []
        by_file[marker.file_path].append(marker)

    click.echo(f"Found {len(markers_list)} markers in {len(by_file)} files:\n")

    for file_path, file_markers in sorted(by_file.items()):
        click.echo(click.style(str(file_path), fg="cyan"))

        for marker in sorted(file_markers, key=lambda m: m.line_number):
            if marker.error:
                # Show error markers in red
                click.echo(
                    f"  {click.style('ERROR', fg='red')} " f"L{marker.line_number}: {marker.req_id}"
                )
                click.echo(f"         {marker.error}")
            else:
                # Normal marker
                status_parts = []
                if marker.scope:
                    status_parts.append(marker.scope)
                if marker.technique:
                    status_parts.append(marker.technique)
                if marker.env:
                    status_parts.append(marker.env)

                status_str = ", ".join(status_parts) if status_parts else ""

                click.echo(
                    f"  L{marker.line_number:4d}: "
                    f"{click.style(marker.req_id, fg='green')}"
                    f"{f' ({status_str})' if status_str else ''}"
                )

                if verbose and marker.function_name:
                    click.echo(f"         function: {marker.function_name}")

        click.echo()

    # Summary
    error_count = sum(1 for m in markers_list if m.error)
    valid_count = len(markers_list) - error_count

    click.echo(f"Summary: {valid_count} valid markers", nl=False)
    if error_count:
        click.echo(f", {click.style(f'{error_count} errors', fg='red')}")
    else:
        click.echo()


def _output_json(markers_list: list[MarkerInfo]) -> None:
    """Output markers as JSON.

    Args:
        markers_list: List of markers to output.
    """
    data = {
        "markers": [m.to_dict() for m in markers_list],
        "summary": {
            "total": len(markers_list),
            "valid": sum(1 for m in markers_list if not m.error),
            "errors": sum(1 for m in markers_list if m.error),
        },
    }
    click.echo(json.dumps(data, indent=2))


def _output_csv(markers_list: list[MarkerInfo]) -> None:
    """Output markers as CSV.

    Args:
        markers_list: List of markers to output.
    """
    import csv
    import sys

    writer = csv.writer(sys.stdout)
    writer.writerow(
        [
            "req_id",
            "file_path",
            "line_number",
            "language",
            "scope",
            "technique",
            "env",
            "function_name",
            "error",
        ]
    )

    for marker in markers_list:
        writer.writerow(
            [
                marker.req_id,
                str(marker.file_path),
                marker.line_number,
                marker.language,
                marker.scope or "",
                marker.technique or "",
                marker.env or "",
                marker.function_name or "",
                marker.error or "",
            ]
        )


@markers.command("languages")
def list_languages() -> None:
    """List supported programming languages.

    Shows all languages that RTMX can parse requirement markers from.
    """
    from rtmx.markers.detection import get_extensions_for_language, get_supported_languages

    click.echo("Supported languages:\n")

    languages = get_supported_languages()
    for lang in languages:
        extensions = get_extensions_for_language(lang)
        ext_str = ", ".join(sorted(extensions))
        click.echo(f"  {click.style(lang, fg='green'):15} {ext_str}")

    click.echo(f"\nTotal: {len(languages)} languages")


@markers.command("schema")
@click.option(
    "--format",
    "output_format",
    type=click.Choice(["json", "yaml"]),
    default="json",
    help="Output format",
)
def show_schema(output_format: str) -> None:
    """Display the marker JSON Schema.

    Shows the canonical schema for requirement marker annotations.
    """
    from rtmx.markers.schema import MARKER_SCHEMA

    if output_format == "json":
        click.echo(json.dumps(MARKER_SCHEMA, indent=2))
    else:
        import yaml

        click.echo(yaml.dump(MARKER_SCHEMA, default_flow_style=False))
