#! /bin/sh

(cd cmd/grnc-bind && go install)
(cd cmd/grnc-ctl && go install)
(cd cmd/grnc-project && go install)