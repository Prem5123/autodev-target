# autodev-target

A simple Go playground that **multica-autodev** uses as the codebase its
agents collaborate on. Agents (CEO / Judge / backend-dev / tester / ...)
clone this repo, write code into it, commit to `autodev/issue-*-cycle-*`
branches, push back, and the autodev daemon promotes or rolls back
those branches based on the Judge's score.

This repo is intentionally minimal — its job is to be a playground, not
a production codebase.

## Layout

- `go.mod` — Go module
- `greet/` — sample package agents will be asked to extend
- `.autodev/` — created by agents per cycle (rubric.json, role-plan.json, score.json)
