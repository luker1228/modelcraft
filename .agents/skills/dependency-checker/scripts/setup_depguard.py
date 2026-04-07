#!/usr/bin/env python3
"""
Setup depguard configuration in .golangci.yml
"""

import sys
import yaml
from pathlib import Path


def setup_depguard(golangci_path: str, depguard_config_path: str) -> bool:
    """
    Add or update depguard configuration in .golangci.yml

    Args:
        golangci_path: Path to .golangci.yml file
        depguard_config_path: Path to depguard configuration template

    Returns:
        True if successful, False otherwise
    """
    try:
        # Read existing golangci config
        with open(golangci_path, 'r') as f:
            golangci_config = yaml.safe_load(f) or {}

        # Read depguard config template
        with open(depguard_config_path, 'r') as f:
            depguard_config = yaml.safe_load(f)

        # Ensure linters-settings exists
        if 'linters-settings' not in golangci_config:
            golangci_config['linters-settings'] = {}

        # Add or update depguard settings
        golangci_config['linters-settings']['depguard'] = depguard_config['depguard']

        # Ensure linters.enable exists and add depguard if not present
        if 'linters' not in golangci_config:
            golangci_config['linters'] = {}
        if 'enable' not in golangci_config['linters']:
            golangci_config['linters']['enable'] = []

        if 'depguard' not in golangci_config['linters']['enable']:
            golangci_config['linters']['enable'].append('depguard')

        # Write back to .golangci.yml
        with open(golangci_path, 'w') as f:
            yaml.dump(golangci_config, f, default_flow_style=False, sort_keys=False)

        print(f"✅ Successfully configured depguard in {golangci_path}")
        return True

    except Exception as e:
        print(f"❌ Error setting up depguard: {e}", file=sys.stderr)
        return False


def main():
    if len(sys.argv) != 3:
        print("Usage: setup_depguard.py <path-to-.golangci.yml> <path-to-depguard-config>")
        sys.exit(1)

    golangci_path = sys.argv[1]
    depguard_config_path = sys.argv[2]

    if not Path(golangci_path).exists():
        print(f"❌ Error: {golangci_path} does not exist", file=sys.stderr)
        sys.exit(1)

    if not Path(depguard_config_path).exists():
        print(f"❌ Error: {depguard_config_path} does not exist", file=sys.stderr)
        sys.exit(1)

    success = setup_depguard(golangci_path, depguard_config_path)
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
