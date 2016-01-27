FROM golang:alpine

ADD . /go/src/github.com/TF2Stadium/Helen

ENV GO15VENDOREXPERIMENT 1
RUN go install -v github.com/TF2Stadium/Helen
RUN cp -R /go/src/github.com/TF2Stadium/Helen/views /go/
RUN cp -R /go/src/github.com/TF2Stadium/Helen/static /go/
RUN cp /go/src/github.com/TF2Stadium/Helen/lobbySettingsData.json /go/
RUN rm -rf /go/src /go/pkg

ENTRYPOINT /go/bin/Helen
