"""
Test user setup utility for integration tests.

Provides functions to automatically create and manage test users with owner roles.
"""

import os
import sys
from pathlib import Path
import pymysql
from pymysql.cursors import DictCursor


def get_setup_sql_path():
    """Get the path to the test user setup SQL script."""
    tests_dir = Path(__file__).parent.parent
    return tests_dir / "setup_test_user.sql"


def execute_test_user_setup(db_config):
    """
    Execute the test user setup SQL script.

    Creates a test user with owner role if not exists (idempotent).

    Args:
        db_config (dict): Database configuration with keys:
            - host: Database host
            - port: Database port
            - user: Database user
            - password: Database password
            - database: Database name

    Returns:
        dict: Test user info {id, external_id, name, org_name, role_name, status}

    Raises:
        RuntimeError: If SQL execution fails or user setup is incomplete
    """
    import subprocess

    sql_path = get_setup_sql_path()

    if not sql_path.exists():
        raise RuntimeError(f"Test user setup SQL not found: {sql_path}")

    # Use MySQL command line for reliable SQL execution
    # This handles complex SQL scripts with variables, comments, etc.
    mysql_cmd = [
        'mysql',
        '-h', db_config.get('host', 'localhost'),
        '-P', str(db_config.get('port', 3306)),
        '-u', db_config.get('user', 'root'),
        f"-p{db_config.get('password', '')}",
        db_config.get('database', 'modelcraft')
    ]

    try:
        print(f"✅ Executing test user setup SQL: {sql_path}")

        # Execute SQL script via mysql command
        with open(sql_path, 'r', encoding='utf-8') as f:
            result = subprocess.run(
                mysql_cmd,
                stdin=f,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                check=True
            )

        # Check if execution was successful by querying the user
        connection = None
        try:
            connection = pymysql.connect(
                host=db_config.get('host', 'localhost'),
                port=int(db_config.get('port', 3306)),
                user=db_config.get('user', 'root'),
                password=db_config.get('password', ''),
                database=db_config.get('database', 'modelcraft'),
                cursorclass=DictCursor
            )

            with connection.cursor() as cursor:
                # Query test user to verify setup
                cursor.execute("""
                    SELECT
                        u.id AS user_id,
                        u.external_id,
                        u.name AS user_name,
                        o.name AS org_name,
                        r.name AS role_name,
                        uo.status
                    FROM users u
                    JOIN user_organizations uo ON u.id = uo.user_id
                    JOIN organizations o ON uo.org_id = o.id
                    JOIN user_roles ur ON u.id = ur.user_id AND o.name = ur.org_name
                    JOIN roles r ON ur.role_id = r.id
                    WHERE u.external_id = 'test-integration'
                    LIMIT 1
                """)
                user_info = cursor.fetchone()

                if not user_info:
                    # Check what's missing
                    cursor.execute("SELECT COUNT(*) as count FROM users WHERE external_id = 'test-integration'")
                    user_exists = cursor.fetchone()['count'] > 0

                    cursor.execute("SELECT COUNT(*) as count FROM organizations WHERE name = 'modelcraft'")
                    org_exists = cursor.fetchone()['count'] > 0

                    cursor.execute("SELECT COUNT(*) as count FROM roles WHERE name = 'owner' AND org_name = '__SYSTEM__'")
                    role_exists = cursor.fetchone()['count'] > 0

                    error_parts = ["Test user setup failed:"]
                    if not user_exists:
                        error_parts.append("  - Test user was not created")
                    if not org_exists:
                        error_parts.append("  - Organization 'modelcraft' does not exist")
                        error_parts.append("    → SQL script should have created it automatically")
                    if not role_exists:
                        error_parts.append("  - Role 'owner' does not exist")
                        error_parts.append("    → Application must initialize system roles on startup")
                        error_parts.append("    → Or manually: INSERT INTO roles (name, description, is_system, org_name) VALUES ('owner', ..., TRUE, '__SYSTEM__')")
                    if user_exists and org_exists and role_exists:
                        error_parts.append("  - User exists but role assignment failed")
                        error_parts.append("    → Check user_roles and user_organizations tables")

                    raise RuntimeError('\n'.join(error_parts))

                return {
                    'id': user_info['user_id'],
                    'external_id': user_info['external_id'],
                    'name': user_info['user_name'],
                    'org_name': user_info['org_name'],
                    'role_name': user_info['role_name'],
                    'status': user_info.get('status', 'active')
                }

        finally:
            if connection:
                connection.close()
                print("✅ Database connection closed")

    except subprocess.CalledProcessError as e:
        error_details = [
            f"❌ MySQL command failed: {e}",
            f"STDERR: {e.stderr}",
            f"Database: {db_config.get('host')}:{db_config.get('port')}/{db_config.get('database')}",
            "Please ensure:",
            "  1. MySQL client is installed (mysql command)",
            "  2. Database is running and accessible",
            "  3. Database credentials are correct",
            "  4. System roles have been initialized (owner, admin, editor, viewer)"
        ]
        raise RuntimeError('\n'.join(error_details))

    except pymysql.Error as e:
        error_details = [
            f"❌ Database error during test user verification: {e}",
            f"Database: {db_config.get('host')}:{db_config.get('port')}/{db_config.get('database')}",
            "SQL script execution may have succeeded, but verification failed."
        ]
        raise RuntimeError('\n'.join(error_details))


def cleanup_test_user(db_config, user_id='487101d6-92bb-459e-b4f1-426255126d27'):
    """
    Clean up test user and associated data with cascading deletions.

    Args:
        db_config (dict): Database configuration
        user_id (str): Test user UUID to delete

    Raises:
        RuntimeError: If cleanup fails (non-critical, logs warning)
    """
    connection = None
    try:
        connection = pymysql.connect(
            host=db_config.get('host', 'localhost'),
            port=int(db_config.get('port', 3306)),
            user=db_config.get('user', 'root'),
            password=db_config.get('password', ''),
            database=db_config.get('database', 'modelcraft'),
            cursorclass=DictCursor,
            autocommit=False
        )

        print(f"🧹 Cleaning up test user: {user_id}")

        with connection.cursor() as cursor:
            # Delete user_roles (will cascade from user FK)
            ur_deleted = cursor.execute(
                "DELETE FROM user_roles WHERE user_id = %s",
                (user_id,)
            )
            if ur_deleted > 0:
                print(f"   Deleted {ur_deleted} user_roles record(s)")

            # Delete user_organizations associations (will cascade from user FK)
            uo_deleted = cursor.execute(
                "DELETE FROM user_organizations WHERE user_id = %s",
                (user_id,)
            )
            if uo_deleted > 0:
                print(f"   Deleted {uo_deleted} user_organizations record(s)")

            # Delete user (cascading will handle roles and orgs)
            user_deleted = cursor.execute(
                "DELETE FROM users WHERE id = %s",
                (user_id,)
            )
            if user_deleted > 0:
                print(f"   Deleted user record")
            else:
                print(f"   ⚠️  User record not found (may have been already deleted)")

            connection.commit()
            print(f"✅ Cleanup completed successfully")

    except pymysql.Error as e:
        error_msg = f"⚠️  Warning: Failed to cleanup test user: {e}"
        if connection:
            connection.rollback()
        print(error_msg)
        # Don't raise - cleanup failure should not break tests
        # Just log the warning

    finally:
        if connection:
            connection.close()


if __name__ == '__main__':
    """
    Standalone CLI execution for manual testing/debugging.

    Usage:
        python tests/common/test_user_setup.py              # Setup test user
        python tests/common/test_user_setup.py --cleanup     # Cleanup test user
        python tests/common/test_user_setup.py --help
    """
    import argparse
    
    parser = argparse.ArgumentParser(
        description='Test user setup utility for integration tests',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  python tests/common/test_user_setup.py              # Create test user
  python tests/common/test_user_setup.py --cleanup     # Cleanup test user
  python tests/common/test_user_setup.py --user-id UUID  # Use specific user ID
        """
    )
    parser.add_argument('--cleanup', '-c', action='store_true',
                        help='Cleanup test user instead of creating')
    parser.add_argument('--user-id', '-u', 
                        default='487101d6-92bb-459e-b4f1-426255126d27',
                        help='User ID to cleanup (default: test-integration user)')
    
    args = parser.parse_args()
    
    # Add tests directory to path
    sys.path.insert(0, str(Path(__file__).parent.parent))

    from common.config import TestConfig, load_env_from_root

    _env_file = os.environ.get('ENV_FILE', '.env')
    load_env_from_root(_env_file)
    config = TestConfig()
    db_config = config.get_db_config()
    print(f"🗄️  Database: {db_config['user']}@{db_config['host']}:{db_config['port']}/{db_config['database']}")

    if args.cleanup:
        print("🧹 Cleaning up test user...")
        try:
            cleanup_test_user(db_config, args.user_id)
            print(f"✅ Test user cleanup completed")
            sys.exit(0)
        except Exception as e:
            print(f"❌ Test user cleanup failed: {e}")
            sys.exit(1)
    else:
        print("🚀 Setting up test user...")
        try:
            user_info = execute_test_user_setup(db_config)
            print(f"✅ Test user created successfully:")
            print(f"   ID: {user_info['id']}")
            print(f"   External ID: {user_info['external_id']}")
            print(f"   Name: {user_info['name']}")
            print(f"   Organization: {user_info['org_name']}")
            print(f"   Role: {user_info['role_name']}")
            print(f"   Status: {user_info['status']}")
            sys.exit(0)
        except Exception as e:
            print(f"❌ Test user setup failed: {e}")
            sys.exit(1)
