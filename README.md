# gitzup

[![Build Status](https://travis-ci.com/kfirz/gitzup.svg?branch=master)](https://travis-ci.com/kfirz/gitzup)
[![Go Report Card](https://goreportcard.com/badge/github.com/kfirz/gitzup?style=flat-square)](https://goreportcard.com/report/github.com/kfirz/gitzup)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/kfirz/gitzup)
[![Release](https://img.shields.io/github/release/kfirz/gitzup.svg?style=flat-square)](https://github.com/kfirz/gitzup/releases/latest)

Gitzup is an opinionated DevOps platform for managing all development aspects, from infrastructure, through coding, to deployment.

## Installation

TBD.

## Usage

### Concepts & Brief

- `GitHubRepository`

  Users can register GitHub repositories from the user interface after installing the Gitzup GitHub App. OInce installed, the user will be able to enable one or more GitHub repositories (public and/or private).
  
- `Project`

  Projects serve as containers for a set of inter-related pipelines. In general a project will usually represent a logical project or application, and not necessarily a single GitHub repository. The exact interpretation, however, is left to the user.
  
- `Pipeline`

  Pipelines are essentially a binding between one or more triggers & desired-state manifests (or desired-state manifest providers). In essence, a pipeline is executed when one or more of its triggers fire; a pipeline execution means **_applying_** the pipeline's manifest (see below for details on manifests).
  
- `Trigger`

  Triggers detect whether certain conditions are met, and if so, they fire an execution of the pipeline they are attached to.
  
  **Supported triggers:**
  
  - GitHub commit pushed
  - GitHub tag pushed
  - GitHub PR created, synced, merged or closed
  - GitHub issue created or closed 
  - Another pipeline successful/failed execution
  - JIRA release created/released                    
  - JIRA issue created or transitioned                    
  - Manual invocation (from the UI or via API)
  
- `Manifest` (aka. desired-resource-state manifests)

  Manifests declare the *desired state* of resources. The idea is that Gitzup will query these resources for their *current state*, and will automatically **_transition_** these resources from their current state to the desired state.

## Components

Gitzup is composed of the following building blocks:

- User interface: a single-page application (SPA) built with Angular/React (TBD)

- API server: used by the UI and/or external clients.

- Web-hooks server: invoked by external services to notify Gitzup of important events (eg. GitHub webhooks calling Gitzup on `push` events).

- Agent: a daemon that executes pipelines.

## Contributing

Please see our [contributing](./CONTRIBUTING.md) document.
