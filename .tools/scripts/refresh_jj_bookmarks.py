import argparse
import subprocess
import sys

def get_current_branch():
    result = subprocess.run(['git', 'rev-parse', '--abbrev-ref', 'HEAD'], stdout=subprocess.PIPE, text=True)
    return result.stdout.strip()

def check_jj_repo():
    """Check if the current directory is a jj repository."""
    result = subprocess.run(['jj', 'status'], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    if result.returncode != 0:
        if 'no jj repo' in result.stderr.lower():
            print("Error: This directory is not initialized as a Jujutsu repository.")
            print("\nTo initialize Jujutsu for this Git repository, run:")
            print("  jj git init --colocate")
            print("\nNote: Use --colocate to work alongside Git in the same directory.")
            sys.exit(1)
        else:
            print(f"Error running jj: {result.stderr}")
            sys.exit(1)
    return True

def get_commit_chain(top_branch, base_branch='main'):
    result = subprocess.run(['git', 'rev-list', top_branch, '--not', base_branch], stdout=subprocess.PIPE, text=True)
    return result.stdout.strip().split('\n')

def get_branch_for_commit(commit_hash):
    result = subprocess.run(['git', 'branch', '--contains', commit_hash], stdout=subprocess.PIPE, text=True)
    branches = [line.strip().lstrip('* ').strip() for line in result.stdout.strip().split('\n')]
    return branches

def find_intermediate_branches(top_branch, base_branch='main'):
    commit_chain = get_commit_chain(top_branch, base_branch)
    bookmarks = []
    for commit in commit_chain:
        branches = get_branch_for_commit(commit)
        filtered = [b for b in branches if b not in [base_branch, top_branch]]
        bookmarks.extend(filtered)
    seen = set()
    unique_bookmarks = []
    for b in bookmarks:
        if b not in seen:
            seen.add(b)
            unique_bookmarks.append(b)
    return unique_bookmarks

def has_uncommitted_changes():
    """Check if there are uncommitted changes in the jj working copy."""
    result = subprocess.run(['jj', 'status'], stdout=subprocess.PIPE, text=True)
    output = result.stdout.strip()
    # Check for working copy changes - jj status shows "Working copy changes:" when there are uncommitted changes.
    return 'Working copy changes:' in output and 'No changes.' not in output

def create_new_revision():
    """Create a new revision with uncommitted changes."""
    subprocess.run(['jj', 'new'])

def create_jj_bookmark(branch_name):
    subprocess.run(['jj', 'bookmark', 'set', branch_name, f'refs/heads/{branch_name}'])

def create_all_jj_bookmarks(branches):
    for branch in branches:
        print(f"Creating Jujutsu bookmark for branch: {branch}")
        create_jj_bookmark(branch)
    print("All bookmarks created.")

EPILOG ='''
Examples:
  %(prog)s                    Create bookmarks for the current branch
  %(prog)s --base develop     Use 'develop' as the base branch instead of 'main'
  %(prog)s --yes              Auto-confirm bookmark creation
            '''


class Main:

    def __init__(self):
        parser = argparse.ArgumentParser(
            description='Create Jujutsu bookmarks for intermediate Git branches between the current branch and main.',
            formatter_class=argparse.RawDescriptionHelpFormatter,
            epilog=EPILOG
        )
        parser.add_argument(
            '--base',
            default='main',
            help='Base branch to compare against (default: main)'
        )
        parser.add_argument(
            '--dry-run', '-n',
            action='store_true',
            help='Show which bookmarks would be created without actually creating them'
        )
        parser.add_argument(
            '--yes', '-y',
            action='store_true',
            help='Automatically confirm bookmark creation without prompting'
        )
        self.args = parser.parse_args()

    def main(self, args=None):
        # Check if jj is initialized first.
        check_jj_repo()
        top_branch = get_current_branch()
        print(f"Current branch (TOP_OF_STACK): {top_branch}")
        # Check for uncommitted changes.
        if has_uncommitted_changes():
            print("\nUncommitted changes detected in working copy.")
            if self.args.yes:
                create_uncommitted = 'y'
            else:
                create_uncommitted = input("Create a new revision for uncommitted changes? (y/n): ").strip().lower()
            if create_uncommitted == 'y':
                if not self.args.dry_run:
                    print("Creating new revision with uncommitted changes...")
                    create_new_revision()
                else:
                    print("[Dry run] Would create new revision with uncommitted changes.")
            print()
        bookmarks_to_create = find_intermediate_branches(top_branch, self.args.base)
        if not bookmarks_to_create:
            print(f"No intermediate branches found between TOP_OF_STACK and '{self.args.base}'.")
            return
        print("\nThe following bookmarks are suggested for creation:")
        for b in bookmarks_to_create:
            print(f"  - {b}")
        if not self.args.dry_run:
            if self.args.yes:
                confirm = 'y'
            else:
                confirm = input("\nDo you want to create these bookmarks in Jujutsu? (y/n): ").strip().lower()
            if confirm == 'y':
                create_all_jj_bookmarks(bookmarks_to_create)
            else:
                print("No bookmarks were created.")
        else:
            print()
            print("Dry run mode enabled. No bookmarks were created.")


if __name__ == "__main__":
    Main().main()

