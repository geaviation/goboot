#!/usr/bin/env bash

##
pkgs=(
    github.com/stretchr/testify/assert
#    github.com/Sirupsen/logrus
#    github.com/gosemver/aws_aws-sdk-go_v1.4.3-1-g1f24fa1/aws
#    github.com/gosemver/aws_aws-sdk-go_v1.4.3-1-g1f24fa1/aws/credential
#    github.com/gosemver/aws_aws-sdk-go_v1.4.3-1-g1f24fa1/aws/request
#    github.com/gosemver/aws_aws-sdk-go_v1.4.3-1-g1f24fa1/aws/session
#    github.com/gosemver/aws_aws-sdk-go_v1.4.3-1-g1f24fa1/service/s3
#    github.com/gosemver/aws_aws-sdk-go_v1.4.3-1-g1f24fa1/service/s3/s3manager
#    github.com/go-xorm/xorm
#    github.com/lib/pq
#    gopkg.in/olivere/elastic.v3
#    github.com/tylerb/graceful
#    github.com/ant0ine/go-json-rest/rest
#    github.com/emicklei/go-restful
)

for pkg in ${pkgs[@]}; do
    echo "getting $pkg"
    go get -d -insecure $pkg
done

go get -v -d -insecure ./...

godep save
