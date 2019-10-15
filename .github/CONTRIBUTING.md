# Development methodology and workflow

This document describes the development methodology and workflow adopted for the P2PDerivatives project.
It is open for comments and suggestions for improvements are welcome.

## Methodology

### Kanban/Scrum

The development methodology is based on Kanban and "scrum flavored".
The main project management tool is [ZenHub](https://app.zenhub.com/workspaces/p2p-derivatives-5d490a9dca85b60001ddc9f4/board?repos=190134999).

The Kanban board incorporates three types of items:
* Epic: represents a big chunk of work with one common objective (for example a new feature to be implemented).
They are first class citizens in Zenhub.
* User Story: represents a small piece of work towards the completion of an objective. User Stories should usually be defined using the [Role-feature-reason template](https://www.agilealliance.org/glossary/user-story-template/).
They are represented as *issues* in Zenhub.
* Task: a smaller chunk of work towards the accomplishment of a user story.
They can be represented either as a markdown checklist in the user story to which they belong, or as a separate issue (mostly in the case that multiple PR would be created for a single user story).

### Sprint

Development work happens in increments during a two weeks period referred to as a *sprint*.

#### Sprint Planning

What work will be carried out during a sprint is decided in the sprint planning meeting.
Two days before the planning meeting, an (github) issue will be created stating:
* The sprint goal,
* A list of proposed issues to be worked on during the sprint (those issues should be detailed enough for team members to understand their scope and complexity).

Team members are welcome to comment on the issue to mention other issues that they think should be added or prioritized, ask clarification on some items, expresses an interest in working on one of the issues, or ask questions.

The list will then serve as a basis for the sprint planning meeting, in which the issues will be estimated and added to the sprint backlog.

#### Daily Scrum

The daily scrum is meant to review the progress of the team to reach the sprint goal, and plan the work for the upcoming day.
As not all team members will be able to participate in the daily scrum, everybody is expected to post a daily update on the #p2pderivatives-daily channel, using the following format:
```
What did I do yesterday:
What will I do today:
(Blockers:)
(I would like to discuss with:)
(Kudos:)
```
As an example:
```
What did I do yesterday:
* Reviewed PR #13,
* Implemented unit tests for the X functionality,
* Read this great article: https://awesomearticle.com

What will I do today:
* Review PR #15,
* Start working on issue #34.

Blockers:
* Nobody commented on my PR #18, please give it some time so that it can be merged before the end of the sprint!

I would like to discuss with:
* @John regarding his PR #14 as I have some questions that are difficult to express in writing.

Kudos:
* To @Jane for spotting a bug in my PR yesterday!
```

#### Sprint Review

The sprint review meeting is held with all the stakeholders to validate the changes made to the product.
Ideally, the development team should demonstrate any new functionality (or changes to existing functionality).
**Make sure to keep that in mind while developing**.
Adding a simple example that can easily be run on the command line can go a long way to showing the development results (nobody wants to see code in a meeting).

Another goal of the sprint review is to confirm the progress of the project with all stakeholders, and review the [product roadmap](https://docs.google.com/spreadsheets/d/1APJzIUnomUA-m9O_mA7ygaTpI44CDJroMe2JAbwsmxU) and update it if necessary.

#### Sprint Retrospective

The sprint retrospective is an opportunity for team members to highlight:
* What went well,
* What didn't go well (and ideally how to improve it),
* What should be improved.

Few days before the end of a sprint, an (github) issue will be created enabling team members to add to any of the above mentioned categories.
Each added item will be discussed during the sprint retrospective meeting, and whenever possible, action items will be derived to try and improve things.

Note that it is ok to mention anything during the team retrospective, whether it be about team collaboration, development process, coding standards...

## Work Items

Different type of issues can be worked on during the sprint, each contributing to the progress of the project in different ways.

### Product Development

Product development work items represent a functionality to be added (or modified) to the system.

#### Definition of done:
A product development work item is considered done when:
* The functionality is working according to the specification (usually a user story),
* The written code has been reviewed and approved, and the corresponding PR has been merged into the development branch,
* Necessary unit and integration tests have been written,
* The necessary documentation has been written or updated (should be checked during code review).

### Infrastructure Development

Infrastructure development work items represent a task aimed at improving the development environment (think CI/CD, linter, testing tools...).

#### Definition of done:
An infrastructure development work item is considered done when:
* The implementation corresponds to the expected outcome,
* Any written code has been reviewed, approved and merged.

### Bug fix

Similar to a development work item.

#### Definition of done:
A bug fix is considered done when:
* The bug no longer occurs in the system,
* Unit or integration tests have been written to ensure that the bug does not re-appear.

> Note that it is usually a good idea to start by writing tests that highlight the bug occurrence before fixing it.

### Design

Before developing software functionalities or components, it is often desirable to provide a design to guide the development work, and get an agreement from all (interested) stakeholders on how it will be implemented.
Design work items aim at producing a design document providing such information.

#### Definition of done:
A design work item it considered done when:
* The design document has been created, reviewed and approved.

### Documentation

Ideally, each software artifact should emanate from a design document, which would serve as a high level documentation for the artifact.
However, this is not always the case, and it is sometimes useful or necessary to add documentation to help new or existing team members or other stakeholders to get an understanding of the system.

#### Definition of done:
A documentation work item is considered done when:
* The document has been written, reviewed and approved.

### Research

To evolve the product, it is important to gain knowledge and understanding of the latest trends.
Research work items aim at obtaining such knowledge and documenting it to make it available to the team members.

#### Definition of done:
A research work item is considered done when:
* The research was conducted and documented,
* The produced document was reviewed and approved.

## Development workflow

### Developing

Development workflow is based on GitFlow.
A good explanation of GitFlow can be found [here](https://datasift.github.io/gitflow/IntroducingGitFlow.html).

The workflow is as follows:
* Pick an issue from the sprint backlog, assign it to yourself and move it to the "In Progress" column on the Kanban board.
* Create a feature branch named after the issue, starting with the issue number ([unfortunately GitHub doesn't enable automating this at the moment](https://github.com/isaacs/github/issues/1125)).
For example, if the issue is "#13 Fix double spend bug", the branch name should be `13-fix-double-spend-bug`.
* Work on you branch adding commits (see [commits](#commits)).
* Once you are happy with your implementation, push your branch to the remote and create a pull request (see [pull requests](#pull-requests)) using the appropriate pull request template.
* On the Kanban board, move your issue to "Review/QA" and tag it with "Review Required". **Do not** move your issue until all the CI tests pass (rules are meant to be broken, if there is a good reason to have code reviewed that doesn't pass CI state it in the PR).

#### Commits

Each commit should represent a single logical unit of work.
Commit titles should be brief and concise and explanatory of the commit content.
Optionally you can add a body if the commit requires more explanations.

Follow these [seven rules](https://chris.beams.io/posts/git-commit/#seven-rules) when creating your commits:
1. Separate subject from body with a blank line
2. Limit the subject line to 50 characters
3. Capitalize the subject line
4. Do not end the subject line with a period
5. Use the imperative mood in the subject line
6. Wrap the body at 72 characters
7. Use the body to explain what and why vs. how

#### Pull requests

Try to keep the amount of changes and number of commits in your pull requests low.
Don't be afraid to mention to the team if you think an issue you have taken is to big, or to break it down and create multiple sub-tasks/sub-issues for it, even after having started development.
Having smaller PR makes [code review](#code-reviewing) easier and the entire development cycle faster.

When addressing code review comments, **do not** create new commits in your feature branch (e.g. "Fixes from review comments").
Instead, make the appropriate changes to the commit on which the reviewer commented (see [here](#rebasing-to-address-review-comments) for how to do that).
This is sometimes (often?) tedious to do, but ensures a clean git history and makes it easier to find bugs (when a feature is implemented in multiple commits it's difficult to find the root cause) and get rid of them (in the worst case a commit can be reverted).

In order to get merged, your feature branch need to be based on top of the `devel` branch to enable fast forward (and have a [clean git history](https://medium.com/@catalinaturlea/clean-git-history-a-step-by-step-guide-eefc0ad8696d)).
See [here](#rebasing-feature-branch-on-devel) for how to do it.

Before opening a pull request, be sure to review your own code to get rid of basic mistakes (typos, dead code...) so as not to waste reviewers' time.
Do not hesitate to comment on the PR to give explanations to the reviewers if you know some things might be difficult to understand.

Remember to update the changelog to document the changes you have made to the system (in the case where the changes are not visible from a user perspective, add the `No Changelog` tag to indicate it).

### Code reviewing

Code reviews are the most important step of the development process.
Be sure to prioritize code reviews over development work.
There is little point in writing code if it never ends up getting merged because nobody reviews it.

When possible, review a PR commit by commit.
Feel free to comment on the PR structure if you find it difficult to review (although being constructive by indicating how the structure can be improved is better).

Code reviewers are expected to verify that:
* The code implements the expected functionality,
* Appropriate unit and integration tests have been written to test the functionality,
* Necessary documentation has been written or updated,
* The changelog has been adequately updated.

Don't be shy of writing **positive** comments on PRs too!

### Coding style

Each project should come equipped with a linter, to make sure that the coding style is consistent.
If something is not covered by the linter, it can be discussed, and a rule can be decided **if the majority of the team agrees with it**.
Ideally, any new rule should come with a tool for enforcing it to avoid wasting time during code review.

#### Variable and function names

Prefer descriptive names for variables and function names to make them easily intelligible.

#### Comments

Avoid using comments that simply restate what the code is doing.
Before adding a comment, ask yourself the following:
* Can I extract part of the code to a function to make it clearer and easier to understand?
* Can I rename some variables or functions to make the code easier to understand?

Acceptable/necessary comments are:
* Comments on public interfaces used for generating documentation,
* Comments to document difficult to understand part of the code (again consider refactoring when possible),
* *TODO* and *FIXME* to document incomplete, missing or broken code that cannot be immediately fixed (don't abuse them).

Always comment on **what** the code is doing rather than on *how* it is doing it.

## Appendix

### Rebasing feature branch on devel

To rebase a feature branch on `devel`, use the following command (while being locally on the feature branch):

```
git fetch
git rebase origin/devel
```

If any conflict pops up during the rebase, fix them on each conflicting commit.
After fixing the conflict, run:

```
git rebase --cont
```

(Do not use `git commit --amend` as this will put all the changes in the previous commit)

It is usually a good practice to re-run the tests after fixing a conflict to make sure the fix didn't introduce any issue.

If something goes wrong during the rebase, you can [abort it](#aborting-a-rebase).

After finishing the rebase, you will have to [force push](#force-pushing-a-branch) your branch to the remote.

### Rebasing to address review comments

Assume you have the following feature branch:

```
git log
commit a4f983e4...
Author: ...
Date: ...
  Commit1
commit b45c980b...
Author: ...
Date: ...
  Commit2
commit e4f8c443...
Author: ...
Date: ...
  Commit3
```

and you get a comment to fix something on Commit2.
Then run the following command:

```
git rebase -i b45c980b^
```

You will then get something like the following:

```
pick b45c980b Commit2
pick e4f8c443 Commit1

#...
#...
```

Change it to:

```
e b45c980b Commit2
pick e4f8c443 Commit1

#...
#...
```

This tells git that you want to edit `Commit2`.

Apply the changes that you want to do.
Once you are happy with the changes, use:

```
git commit --amend --no-edit
```

This will commit your changes to `Commit2` and preserve the commit name (if you want to change the commit name, drop the `--no-edit` part).

Then run:

```
git rebase --cont
```

If some conflicts arise, fix them and then run:

```
git rebase --cont
```

again.

Note again that you shouldn't use `git commit --amend` after fixing a rebase conflict.

If something goes wrong during a rebase, you can [abort it](#aborting-a-rebase)

After finishing the rebase, you will have to [force push](#force-pushing-a-branch) your branch to the remote.

### Aborting a rebase

If something goes wrong during a rebase, run:

```
git rebase --abort
```

to abort the rebase.
This will bring you back to the state before you started the rebase.

### Force pushing a branch

When you rebase your feature branch, it will get out of sync and conflict with the feature branch on the remote.
To get the remote branch in sync, you will have to force push it using:

```
git push -f
```

**WARNING**

Never force push a branch that is not **your** feature branch or a branch that is used by anybody else than you.

### Resetting out of sync branch

If one of your local branch is out of sync with the remote (for example if you review someone's PR and he forced push some changes), you will have to reset it using:

```
git fetch
git reset --hard origin/feature-branch
```

where `feature-branch` is the name of the branch you want to reset.

(Note that `git fetch` ensures that you have retrieved the latest version from the remote.)

**WARNING**

Hard resetting a branch means you will lose all local changes to the branch.
You should usually only do this for branches on which you are not working, or have not yet started working on.
