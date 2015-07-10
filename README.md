# TeamPlayTF
TeamPlayTF Server is the backend server component for the TeamPlay.tf project written in Go.

### Setup
The project uses MongoDB as a database. Default production account data can be found at `database/setup.txt`

### Structure
The code is divided into multiple packages that follow the usual web application structure:
* models go in `models`
* controllers go in `controllers`
* routes go in `routes/routes.go`
* TODO views currently go to static, until work on frontend code starts

### Contribution guidelines
The project uses the Pull Request workflow to contribute code. More info on that here: https://help.github.com/articles/using-pull-requests/.

**Each pull request must pass all existing tests and include new appropriate tests.**

Submitted code has to be formatted with `go fmt`.
