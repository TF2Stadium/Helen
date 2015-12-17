#!/bin/sh

go test -coverprofile=handler.out ./controllers/socket/internal/handler/
go tool cover -html=handler.out -o handler.html

go test -coverprofile=models.out ./models/
go tool cover -html=models.out -o models.html

git clone git@github.com:TF2Stadium/coverage.git
cp handler.html models.html ./coverage/
cd coverage
git config --global user.email "this@is.bot"
git config --global user.name "circleci deploy"
git add models.html handler.html
git commit -m "Update coverage" && git push -f
