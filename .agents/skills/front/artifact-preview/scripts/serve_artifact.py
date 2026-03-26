#!/usr/bin/env python3
"""
Start HTTP server for artifact preview with automatic port conflict resolution.
"""
import os
import sys
import signal
import subprocess
import time
from pathlib import Path

PORT = 9999

def find_process_using_port(port):
    """Find process ID using the specified port."""
    try:
        result = subprocess.run(
            ["lsof", "-ti", f":{port}"],
            capture_output=True,
            text=True,
            check=False
        )
        if result.returncode == 0 and result.stdout.strip():
            return result.stdout.strip().split('\n')
        return []
    except FileNotFoundError:
        # lsof not available, try netstat
        try:
            result = subprocess.run(
                ["netstat", "-tuln"],
                capture_output=True,
                text=True,
                check=False
            )
            for line in result.stdout.split('\n'):
                if f':{port}' in line and 'LISTEN' in line:
                    return ["unknown"]
            return []
        except FileNotFoundError:
            return []

def kill_processes(pids):
    """Kill processes by PID."""
    for pid in pids:
        try:
            os.kill(int(pid), signal.SIGTERM)
            print(f"✓ Killed process {pid}")
        except (ProcessLookupError, ValueError):
            pass

def start_server(directory):
    """Start HTTP server in the specified directory."""
    os.chdir(directory)

    # Check for port conflict
    pids = find_process_using_port(PORT)
    if pids:
        print(f"⚠ Port {PORT} is in use by process(es): {', '.join(pids)}")
        print(f"✓ Killing conflicting process(es)...")
        kill_processes(pids)
        time.sleep(1)  # Wait for port to be released

    # Start server
    print(f"\n🚀 Starting HTTP server...")
    print(f"📁 Serving directory: {directory}")
    print(f"🌐 Access at: http://localhost:{PORT}/bundle.html\n")
    print("Press Ctrl+C to stop the server\n")

    try:
        subprocess.run([sys.executable, "-m", "http.server", str(PORT)])
    except KeyboardInterrupt:
        print("\n\n✓ Server stopped")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python3 serve_artifact.py <artifact-directory>")
        sys.exit(1)

    artifact_dir = Path(sys.argv[1]).resolve()

    if not artifact_dir.exists():
        print(f"❌ Error: Directory not found: {artifact_dir}")
        sys.exit(1)

    bundle_file = artifact_dir / "bundle.html"
    if not bundle_file.exists():
        print(f"⚠ Warning: bundle.html not found in {artifact_dir}")
        print("Continuing anyway...")

    start_server(artifact_dir)
