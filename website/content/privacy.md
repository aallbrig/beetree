---
title: "Privacy"
description: "What BeeTreeCraft stores about you, and what it does not."
url: "/privacy/"
---

BeeTreeCraft is a hobby project with no advertising and no tracking. This page says what the code actually stores.

## Without an account

You can use the builder and Explore without signing in. Your unsaved draft tree stays in your browser's local storage and never leaves your machine, and so does the Communitree list of trees you saved before signing up. A session-storage flag remembers that you dismissed the sign-up nudge. The site sets no tracking cookies and loads no analytics scripts.

When you sign in, your session token is kept in local storage (`beetree.session`) until you sign out.

## With an account

Sign-up uses Amazon Cognito. Cognito holds your email address and your birthdate. The birthdate is used once, to confirm you are an adult, and cannot be changed afterwards. Passwords are handled by Cognito; BeeTreeCraft never sees them.

The API stores the following in Amazon DynamoDB:

- your profile: user id, display name, bio and creation time (your email is visible only to you)
- the trees you save, with their public or private setting
- your likes, comments and forum posts, shown with your display name
- API keys you create: only a SHA-256 hash and the first eight characters are kept, so the full key cannot be recovered

Public trees, comments and posts are visible to everyone. Private trees are visible only to you.

## Logs and servers

Requests to the API pass through AWS, which keeps standard access and error logs for operating the service. They are not used for profiling and are not sold or shared.

## Deleting your data

Delete a tree, comment or post in the app to remove it. To have your account and its data removed, [open an issue](https://github.com/aallbrig/beetreecraft/issues/new) asking for deletion without including personal details, and the maintainer will arrange a private way to confirm it is you.
