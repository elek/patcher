# Patcher

Patcher is a simple tool to handle ASF patches:

 * Download patch from Jira and apply
 * Create commit message based on the Jira assignee and description.
 * Create patch and upload it to the Jira.

Typical workflow:

 * `git checkout -b HDDS-12`
 * `patcher apply` (downloads and applies the patch)
 * `git add`
 * `patcher commit` (downloads Jira summary and creates commit message and commit)
 * `patcher create` (creates the patch file and save to /tmp/)
 * `patcher create --upload` (uploads the patch to the Apache Jira)

Use `--help` to check the available parameters.

On high level we need the following information:

 * Name of the current Jira issue. By default it comes from the branch name, but could be specified.
 * Name of the working branch. By default patcher tries to find a remote branch with apache prefix in the last 40 commits. It will be used as the working branch (could be adjusted by cli parameters). You always need to rebase or merge to the latest apache branch!
 * Base commit of the patch. Patcher supports multi-commit patches. By default the diff will be created between the working branch (eg. apache/trunk) and the current HEAD.
