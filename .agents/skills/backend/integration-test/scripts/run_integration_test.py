#!/usr/bin/env python3
"""
Run integration tests for ModelCraft design-time modules using pytest.
Generates HTML test reports and provides clean terminal output.
Optionally serves reports via HTTP server for remote access.
"""

import argparse
import subprocess
import sys
import socket
import time
import shutil
from pathlib import Path
from datetime import datetime
from typing import Tuple, Optional


FIXED_REPORT_PORT = 10001


def kill_port(port: int) -> None:
    """
    Kill any process listening on the given port.

    Args:
        port: Port number to free up
    """
    try:
        result = subprocess.run(
            ["ss", "-tlnp", f"sport = :{port}"],
            capture_output=True, text=True, timeout=3
        )
        # Extract PIDs from ss output lines like: users:(("python",pid=12345,fd=4))
        import re
        pids = re.findall(r'pid=(\d+)', result.stdout)
        for pid in set(pids):
            try:
                subprocess.run(["kill", "-9", pid], capture_output=True, timeout=3)
                print(f"🔪 Killed process {pid} on port {port}")
            except Exception:
                pass
        if pids:
            time.sleep(0.5)  # Give OS time to release the port
    except Exception as e:
        print(f"⚠️  Could not kill process on port {port}: {e}")


def get_server_ip() -> str:
    """
    Get the server's IP address for remote access.

    Returns:
        Server IP address as string
    """
    try:
        # Create a socket to find the IP address
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(("8.8.8.8", 80))
        ip = s.getsockname()[0]
        s.close()
        return ip
    except Exception:
        return "127.0.0.1"


def is_port_listening(port: int, max_retries: int = 10, retry_delay: float = 0.5) -> bool:
    """
    Check if a port is listening using lsof command.

    Args:
        port: Port number to check
        max_retries: Maximum number of retry attempts
        retry_delay: Delay between retries in seconds

    Returns:
        True if port is listening, False otherwise
    """
    for attempt in range(max_retries):
        try:
            # Use lsof to check if port is listening
            result = subprocess.run(
                ["lsof", "-i", f":{port}"],
                capture_output=True,
                text=True,
                timeout=2
            )
            # If lsof returns success and has output, port is listening
            if result.returncode == 0 and result.stdout.strip():
                return True
        except (subprocess.TimeoutExpired, FileNotFoundError):
            pass

        # Wait before retry
        if attempt < max_retries - 1:
            time.sleep(retry_delay)

    return False


def clear_reports() -> None:
    """
    Clear all test reports from the reports directory.
    """
    # Determine project root (assume script is in .claude/skills/integration-test/scripts/)
    project_root = Path(__file__).parent.parent.parent.parent.parent
    report_dir = project_root / "tests" / "reports"

    if not report_dir.exists():
        print(f"📁 Reports directory does not exist: {report_dir}")
        return

    # Count reports before deletion
    reports = list(report_dir.glob("test_report_*.html"))
    report_count = len(reports)

    if report_count == 0:
        print("✨ No reports to clear. Reports directory is already empty.")
        return

    # Ask for confirmation
    print(f"🗑️  Found {report_count} report(s) in {report_dir}")
    print("\nReports to be deleted:")
    for report in reports[:5]:  # Show first 5
        print(f"  - {report.name}")
    if report_count > 5:
        print(f"  ... and {report_count - 5} more")

    response = input(f"\n⚠️  Delete all {report_count} report(s)? [y/N]: ").strip().lower()

    if response in ['y', 'yes']:
        deleted_count = 0
        for report in reports:
            try:
                report.unlink()
                deleted_count += 1
            except Exception as e:
                print(f"❌ Failed to delete {report.name}: {e}")

        print(f"\n✅ Successfully deleted {deleted_count}/{report_count} report(s)")

        # Also clean up assets directory if it exists
        assets_dir = report_dir / "assets"
        if assets_dir.exists():
            try:
                shutil.rmtree(assets_dir)
                print("✅ Cleaned up assets directory")
            except Exception as e:
                print(f"⚠️  Failed to clean assets directory: {e}")
    else:
        print("❌ Deletion cancelled")


def start_http_server(report_dir: Path, port: int) -> None:
    """
    Start HTTP server as a nohup background process to serve test reports.

    Args:
        report_dir: Directory containing HTML reports
        port: Port number to serve on
    """
    import os
    log_file = report_dir / "http_server.log"
    subprocess.Popen(
        [
            "nohup",
            sys.executable, "-m", "http.server", str(port),
        ],
        cwd=str(report_dir),
        stdout=open(log_file, "a"),
        stderr=subprocess.STDOUT,
        start_new_session=True,
    )


def run_pytest(test_path: str, verbose: bool = True, env_file: Optional[str] = None) -> Tuple[int, str, int]:
    """
    Run pytest on specified path with HTML report generation.
    Always starts an HTTP server on FIXED_REPORT_PORT (killing any existing one first).

    Args:
        test_path: Path to test directory or file (e.g., 'tests/design/project')
        verbose: Enable verbose output
        env_file: Name of .env file to load (e.g., '.env.autotest'). Defaults to '.env'.

    Returns:
        Tuple of (return_code, report_path, server_port)
    """
    # Determine project root (assume script is in .claude/skills/integration-test/scripts/)
    project_root = Path(__file__).parent.parent.parent.parent.parent
    test_full_path = project_root / test_path

    # Validate test path exists
    if not test_full_path.exists():
        print(f"❌ Error: Test path does not exist: {test_full_path}")
        sys.exit(1)

    # Generate report filename with timestamp
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    test_name = Path(test_path).name or "all"
    report_dir = project_root / "tests" / "reports"
    report_dir.mkdir(exist_ok=True)
    report_path = report_dir / f"test_report_{test_name}_{timestamp}.html"

    # Build pytest command
    cmd = [
        "pytest",
        str(test_full_path),
        f"--html={report_path}",
        "--self-contained-html",
    ]

    if verbose:
        cmd.append("-v")

    # Build environment: pass ENV_FILE if specified
    import os
    env = os.environ.copy()
    if env_file:
        env["ENV_FILE"] = env_file
        print(f"🌍 Using env file: {env_file}")

    # Run test user setup before pytest
    setup_script = project_root / "tests" / "common" / "test_user_setup.py"
    if setup_script.exists():
        print(f"👤 Checking test user setup...")
        setup_result = subprocess.run(
            [sys.executable, str(setup_script)],
            cwd=project_root,
            env=env,
        )
        if setup_result.returncode != 0:
            print(f"❌ Test user setup failed, aborting tests.")
            sys.exit(1)
        print("-" * 80)
    else:
        print(f"⚠️  test_user_setup.py not found at {setup_script}, skipping.")

    # Run pytest
    print(f"🔍 Running tests: {test_path}")
    print(f"📊 Report will be generated at: {report_path}")
    print("-" * 80)

    result = subprocess.run(cmd, cwd=project_root, env=env)

    print("-" * 80)
    if result.returncode == 0:
        print(f"✅ All tests passed!")
    else:
        print(f"❌ Some tests failed (exit code: {result.returncode})")

    print(f"📄 HTML report: {report_path}")

    # Always start HTTP server on fixed port, killing any existing server first
    server_port = FIXED_REPORT_PORT
    try:
        kill_port(server_port)
        start_http_server(report_dir, server_port)
        server_ip = get_server_ip()
        report_filename = report_path.name

        print(f"\n⏳ Starting HTTP server on port {server_port}...")
        if is_port_listening(server_port):
            print("\n" + "=" * 80)
            print("🌐 HTTP Server Started (background)")
            print("=" * 80)
            print(f"🔗 Report URL:")
            print(f"   http://{server_ip}:{server_port}/{report_filename}")
            print(f"\n💡 All reports: http://{server_ip}:{server_port}/")
            print("=" * 80)
        else:
            print(f"⚠️  Warning: Server may not have started on port {server_port}")
    except Exception as e:
        print(f"\n⚠️  Failed to start HTTP server: {e}")

    return result.returncode, str(report_path), server_port


def main():
    parser = argparse.ArgumentParser(
        description="Run ModelCraft integration tests with HTML reporting",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Test specific module
  %(prog)s tests/design/project

  # Test all design-time modules
  %(prog)s tests/design/

  # Test without verbose output
  %(prog)s tests/design/project --no-verbose

  # Clear all test reports
  %(prog)s --clear
        """
    )

    parser.add_argument(
        "test_path",
        nargs='?',  # Make test_path optional
        help="Path to test directory or file (e.g., 'tests/design/project' or 'tests/design/')"
    )

    parser.add_argument(
        "--no-verbose",
        action="store_true",
        help="Disable verbose pytest output"
    )

    parser.add_argument(
        "--env-file",
        default=None,
        help="Name of .env file to load (e.g., '.env.autotest'). Defaults to '.env'."
    )

    parser.add_argument(
        "--clear",
        action="store_true",
        help="Clear all test reports from tests/reports/ directory"
    )

    args = parser.parse_args()

    # Handle clear command
    if args.clear:
        clear_reports()
        sys.exit(0)

    # Validate test_path is provided when not clearing
    if not args.test_path:
        parser.error("test_path is required unless using --clear")

    return_code, report_path, server_port = run_pytest(
        test_path=args.test_path,
        verbose=not args.no_verbose,
        env_file=args.env_file,
    )

    sys.exit(return_code)


if __name__ == "__main__":
    main()
