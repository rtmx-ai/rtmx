"""Migration command to help users transition from Python CLI to Go CLI.

REQ-DIST-002: rtmx migrate command helps users transition workflows.
"""

from __future__ import annotations

import os
import platform
import shutil
import stat
import subprocess
import sys
import tempfile
import urllib.request
from pathlib import Path
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    pass

from rtmx.formatting import Colors

# GitHub release URLs
RELEASE_BASE = "https://github.com/rtmx-ai/rtmx-go/releases/latest/download"

# Platform mappings
PLATFORM_MAP = {
    ("Darwin", "x86_64"): "rtmx_darwin_amd64.tar.gz",
    ("Darwin", "arm64"): "rtmx_darwin_arm64.tar.gz",
    ("Linux", "x86_64"): "rtmx_linux_amd64.tar.gz",
    ("Linux", "aarch64"): "rtmx_linux_arm64.tar.gz",
    ("Windows", "AMD64"): "rtmx_windows_amd64.zip",
}


def detect_platform() -> tuple[str, str]:
    """Detect current platform and architecture."""
    system = platform.system()
    machine = platform.machine()
    return system, machine


def get_download_url() -> str | None:
    """Get the download URL for current platform."""
    system, machine = detect_platform()
    filename = PLATFORM_MAP.get((system, machine))
    if filename:
        return f"{RELEASE_BASE}/{filename}"
    return None


def get_install_dir() -> Path:
    """Get the installation directory."""
    # Try common locations
    if sys.platform == "win32":
        local_bin = Path(os.environ.get("LOCALAPPDATA", "")) / "rtmx" / "bin"
    else:
        local_bin = Path.home() / ".local" / "bin"

    return local_bin


def download_file(url: str, dest: Path) -> bool:
    """Download a file from URL to destination."""
    try:
        print(f"  Downloading from {url}...")
        urllib.request.urlretrieve(url, dest)
        return True
    except Exception as e:
        print(f"  {Colors.RED}Download failed: {e}{Colors.RESET}")
        return False


def extract_archive(archive: Path, dest_dir: Path) -> Path | None:
    """Extract archive and return path to binary."""
    import tarfile
    import zipfile

    try:
        if str(archive).endswith(".zip"):
            with zipfile.ZipFile(archive, "r") as zf:
                zf.extractall(dest_dir)
            binary_name = "rtmx.exe"
        else:
            with tarfile.open(archive, "r:gz") as tf:
                tf.extractall(dest_dir)
            binary_name = "rtmx"

        binary_path = dest_dir / binary_name
        if binary_path.exists():
            return binary_path
        return None
    except Exception as e:
        print(f"  {Colors.RED}Extraction failed: {e}{Colors.RESET}")
        return None


def install_binary(src: Path, dest_dir: Path) -> Path | None:
    """Install binary to destination directory."""
    try:
        dest_dir.mkdir(parents=True, exist_ok=True)

        if sys.platform == "win32":
            dest = dest_dir / "rtmx-go.exe"
        else:
            dest = dest_dir / "rtmx-go"

        shutil.copy2(src, dest)

        # Make executable on Unix
        if sys.platform != "win32":
            dest.chmod(dest.stat().st_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)

        return dest
    except Exception as e:
        print(f"  {Colors.RED}Installation failed: {e}{Colors.RESET}")
        return None


def verify_installation(binary: Path) -> tuple[bool, str]:
    """Verify the installed binary works."""
    try:
        result = subprocess.run(
            [str(binary), "version"],
            capture_output=True,
            text=True,
            timeout=10,
        )
        if result.returncode == 0:
            return True, result.stdout.strip()
        return False, result.stderr.strip()
    except Exception as e:
        return False, str(e)


def compare_outputs() -> list[tuple[str, bool, str]]:
    """Compare Python and Go CLI outputs for key commands."""
    commands = ["status", "backlog", "health"]
    results: list[tuple[str, bool, str]] = []

    go_binary = shutil.which("rtmx-go")
    if not go_binary:
        return results

    for cmd in commands:
        try:
            # Run Python CLI
            py_result = subprocess.run(
                ["python", "-m", "rtmx", cmd],
                capture_output=True,
                text=True,
                timeout=30,
            )

            # Run Go CLI
            go_result = subprocess.run(
                [go_binary, cmd],
                capture_output=True,
                text=True,
                timeout=30,
            )

            # Compare exit codes
            match = py_result.returncode == go_result.returncode
            detail = (
                "exit codes match"
                if match
                else f"Python={py_result.returncode}, Go={go_result.returncode}"
            )
            results.append((cmd, match, detail))

        except Exception as e:
            results.append((cmd, False, str(e)))

    return results


def run_migrate(
    verify_only: bool = False,
    install_dir: Path | None = None,
    alias: bool = False,
) -> int:
    """Run the migration process.

    Args:
        verify_only: Just verify, don't install.
        install_dir: Custom installation directory.
        alias: Offer to create shell alias.

    Returns:
        Exit code (0 for success).
    """
    print(f"{Colors.BOLD}=== RTMX Migration: Python CLI → Go CLI ==={Colors.RESET}")
    print()

    # Step 1: Detect platform
    system, machine = detect_platform()
    print(f"{Colors.CYAN}Step 1: Platform Detection{Colors.RESET}")
    print(f"  Platform: {system} {machine}")

    url = get_download_url()
    if not url:
        print(f"  {Colors.RED}Unsupported platform: {system} {machine}{Colors.RESET}")
        print()
        print("Manual installation options:")
        print("  go install github.com/rtmx-ai/rtmx-go/cmd/rtmx@latest")
        print("  https://github.com/rtmx-ai/rtmx-go/releases/latest")
        return 1

    print(f"  Binary: {url.split('/')[-1]}")
    print()

    # Step 2: Check if Go CLI already installed
    print(f"{Colors.CYAN}Step 2: Checking Existing Installation{Colors.RESET}")
    existing = shutil.which("rtmx-go")
    if existing:
        print(f"  {Colors.GREEN}Found: {existing}{Colors.RESET}")
        ok, version = verify_installation(Path(existing))
        if ok:
            print(f"  Version: {version}")
    else:
        print("  Not installed")
    print()

    if verify_only:
        if existing:
            print(f"{Colors.CYAN}Step 3: Verification{Colors.RESET}")
            results = compare_outputs()
            for cmd, match, detail in results:
                status = (
                    f"{Colors.GREEN}✓{Colors.RESET}" if match else f"{Colors.RED}✗{Colors.RESET}"
                )
                print(f"  {status} rtmx {cmd}: {detail}")
            return 0 if all(m for _, m, _ in results) else 1
        else:
            print(
                f"{Colors.YELLOW}Go CLI not installed. Run without --verify-only to install.{Colors.RESET}"
            )
            return 1

    # Step 3: Download
    print(f"{Colors.CYAN}Step 3: Download{Colors.RESET}")
    with tempfile.TemporaryDirectory() as tmpdir:
        tmppath = Path(tmpdir)
        archive_name = url.split("/")[-1]
        archive_path = tmppath / archive_name

        if not download_file(url, archive_path):
            return 1
        print(f"  {Colors.GREEN}Downloaded{Colors.RESET}")

        # Step 4: Extract
        print()
        print(f"{Colors.CYAN}Step 4: Extract{Colors.RESET}")
        binary = extract_archive(archive_path, tmppath)
        if not binary:
            return 1
        print(f"  {Colors.GREEN}Extracted{Colors.RESET}")

        # Step 5: Install
        print()
        print(f"{Colors.CYAN}Step 5: Install{Colors.RESET}")
        dest_dir = install_dir or get_install_dir()
        installed = install_binary(binary, dest_dir)
        if not installed:
            return 1
        print(f"  {Colors.GREEN}Installed to: {installed}{Colors.RESET}")

        # Step 6: Verify
        print()
        print(f"{Colors.CYAN}Step 6: Verify{Colors.RESET}")
        ok, version = verify_installation(installed)
        if ok:
            print(f"  {Colors.GREEN}✓ Working: {version}{Colors.RESET}")
        else:
            print(f"  {Colors.RED}✗ Verification failed: {version}{Colors.RESET}")
            return 1

    # Step 7: PATH instructions
    print()
    print(f"{Colors.CYAN}Step 7: PATH Configuration{Colors.RESET}")
    dest_str = str(dest_dir)
    path_env = os.environ.get("PATH", "")
    if dest_str in path_env:
        print(f"  {Colors.GREEN}✓ {dest_dir} is in PATH{Colors.RESET}")
    else:
        print(f"  {Colors.YELLOW}Add to PATH:{Colors.RESET}")
        if sys.platform == "win32":
            print(f'    $env:PATH = "{dest_dir};$env:PATH"')
        else:
            print(f'    export PATH="{dest_dir}:$PATH"')
            print()
            print("  Or add to shell profile:")
            shell = os.environ.get("SHELL", "")
            if "zsh" in shell:
                print(f"    echo 'export PATH=\"{dest_dir}:$PATH\"' >> ~/.zshrc")
            else:
                print(f"    echo 'export PATH=\"{dest_dir}:$PATH\"' >> ~/.bashrc")

    # Step 8: Alias suggestion
    if alias:
        print()
        print(f"{Colors.CYAN}Step 8: Shell Alias{Colors.RESET}")
        print("  To use 'rtmx' for Go CLI:")
        if sys.platform == "win32":
            print(f'    Set-Alias rtmx "{installed}"')
        else:
            print(f'    alias rtmx="{installed}"')

    print()
    print(f"{Colors.GREEN}=== Migration Complete ==={Colors.RESET}")
    print()
    print("Next steps:")
    print(f"  1. Add {dest_dir} to PATH (if needed)")
    print("  2. Run: rtmx-go status")
    print("  3. Once verified, alias rtmx=rtmx-go")
    print()
    print("To suppress Python CLI deprecation warnings:")
    print("  export RTMX_SUPPRESS_DEPRECATION=1")

    return 0
