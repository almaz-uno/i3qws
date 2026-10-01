# Agent Instructions

## Project Language

The primary language for this project is **English**.

All code, documentation, comments, commit messages, and communication should be in English.

## Logging

The project uses **logrus** for structured logging.

Use `github.com/sirupsen/logrus` with fields (`logrus.WithField`, `logrus.WithFields`) rather than the standard library `log` package.

## Spec-driven development

A change that will outlive a single sitting gets a written specification in
`specs/` before any code is written.

The directory layout, the set of documents a particular change needs and the
rules a specification must satisfy are described in `specs/README.adoc`. Rules
that apply to every change regardless of its specification are collected in
`specs/constitution.adoc` — read it before starting work, and keep the
specification consistent with the code within a single change set.
