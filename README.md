# furbnicator

A tool to simplify everyday tasks for the common corporate programmer.

## What

furbnicator consists of various modules, each supporting a distinct system or
task. Each module provides a list of actions and batarang is in the end not much
more than a simple interface to search & execute the tasks provided by all
activated modules.

### Bitbucket

The bitbucket module can index repositories from a single bitbucket server
installation. Especially handy if you regularly need to browse or clone
repositories. If you work with a single repo most of the time, this module might
not help you very much.

#### Tasks

- Clone a repository
- Browse a repository

### Jenkins

The jenkins meodule can index the jobs in a single jenkins installation.

#### Tasks

- Run a job
- Browse a job

### Timestamps

Displays the current unix or java timestamp.

## Demo run

tbd.

## Configuration

tbd.

## Todo

### RobinModule

The RobinModule will support external task definitions e.g. shell scripts. You
might wonder why you need a task runner to run shell scripts if you already have
... a shell. Well, don't ask me :D

## Build

`go build`

## License

EUPL-1.2

- See [LICENSE](LICENSE)
- See https://joinup.ec.europa.eu/collection/eupl
